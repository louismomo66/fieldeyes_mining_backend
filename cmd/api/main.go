package main

import (
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

	// Set JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // Default for development
	}
	utils.SetJWTSecret(jwtSecret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(app.Models.User, app.Mailer)
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

	// Create server
	server := &http.Server{
		Addr:         ":9006",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
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

	// Graceful shutdown
	if err := server.Shutdown(nil); err != nil {
		app.ErrorLog.Fatalf("Server forced to shutdown: %v", err)
	}

	app.InfoLog.Println("Server exited")
}
