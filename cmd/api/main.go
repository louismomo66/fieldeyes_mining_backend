package main

import (
	"context"
	"log"
	"mineral/data"
	"mineral/handlers"
	"mineral/pkg/email"
	"mineral/pkg/utils"
	"mineral/routes"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	app := &Config{
		InfoLog:       log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLog:      log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		Wait:          &sync.WaitGroup{},
		ErrorChan:     make(chan error),
		ErrorChanDone: make(chan bool),
	}

	// Initialize database
	app.DB = app.initDB()
	app.InfoLog.Println("Database connection established")

	// Initialize repositories
	app.Models = data.Models{
		User:      data.NewUserRepository(app.DB),
		Income:    data.NewIncomeRepository(app.DB),
		Expense:   data.NewExpenseRepository(app.DB),
		Inventory: data.NewInventoryRepository(app.DB),
		MineSite:  data.NewMineSiteRepository(app.DB),
	}

	// Initialize mailer (mock for development)
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost != "" && smtpPortStr != "" && smtpUsername != "" && smtpPassword != "" && smtpFrom != "" {
		port, err := strconv.Atoi(smtpPortStr)
		if err != nil {
			app.InfoLog.Printf("Invalid SMTP_PORT (%s), falling back to mock mailer", smtpPortStr)
			app.Mailer = &email.MockMailer{}
		} else {
			app.InfoLog.Println("Using SMTP mailer for OTP delivery")
			app.Mailer = email.NewSMTPMailer(smtpHost, port, smtpUsername, smtpPassword, smtpFrom)
		}
	} else {
		app.InfoLog.Println("SMTP configuration incomplete, using mock mailer")
		app.Mailer = &email.MockMailer{}
	}

	// Set JWT secret from environment — no fallback; app will not start without it
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		app.ErrorLog.Fatal("JWT_SECRET environment variable is required and must not be empty")
	}
	utils.SetJWTSecret(jwtSecret)

	// Load admin registration code from environment (optional — disables admin self-registration if unset)
	adminCode := os.Getenv("ADMIN_REGISTRATION_CODE")
	if adminCode == "" {
		app.InfoLog.Println("ADMIN_REGISTRATION_CODE not set — admin self-registration is disabled")
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(app.Models.User, app.Mailer, adminCode)
	incomeHandler := handlers.NewIncomeHandler(app.Models.Income)
	expenseHandler := handlers.NewExpenseHandler(app.Models.Expense, app.Models.MineSite)
	inventoryHandler := handlers.NewInventoryHandler(app.Models.Inventory)
	analyticsHandler := handlers.NewAnalyticsHandler(app.Models.Income, app.Models.Expense)
	mineSiteHandler := handlers.NewMineSiteHandler(app.Models.MineSite)
	adminHandler := handlers.NewAdminHandler(app.Models.User, app.Models.Income, app.Models.Expense, app.Models.Inventory)

	// Setup routes
	router := routes.SetupRoutes(
		authHandler,
		incomeHandler,
		expenseHandler,
		inventoryHandler,
		analyticsHandler,
		mineSiteHandler,
		adminHandler,
	)

	// Wrap router with a global request body size limit (4 MB) to prevent
	// large payload attacks before they reach any handler.
	limitedRouter := http.MaxBytesHandler(router, 4<<20)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9006"
	}

	// Create server with tuned timeouts for higher concurrency:
	//   ReadHeaderTimeout — fast rejection of slow-header attacks
	//   ReadTimeout       — max time to read full request body
	//   WriteTimeout      — max time to write response
	//   IdleTimeout       — keep-alive connection reuse window
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           limitedRouter,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB header limit
	}

	// Start server in a goroutine
	go func() {
		app.InfoLog.Printf("Starting server on port %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.ErrorLog.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.InfoLog.Println("Server is shutting down...")

	// Give in-flight requests up to 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.ErrorLog.Fatalf("Server forced to shutdown: %v", err)
	}

	app.InfoLog.Println("Server exited")
}
