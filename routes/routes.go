package routes

import (
	"mineral/handlers"
	"mineral/pkg/middleware"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// SetupRoutes configures all API routes using chi router
func SetupRoutes(
	authHandler *handlers.AuthHandler,
	incomeHandler *handlers.IncomeHandler,
	expenseHandler *handlers.ExpenseHandler,
	inventoryHandler *handlers.InventoryHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	mineSiteHandler *handlers.MineSiteHandler,
	adminHandler *handlers.AdminHandler,
) http.Handler {
	r := chi.NewRouter()

	// Build allowed origins from environment.
	// ALLOWED_ORIGINS should be a comma-separated list, e.g.:
	//   ALLOWED_ORIGINS=https://app.fieldeyes.com,https://admin.fieldeyes.com
	// Falls back to localhost only for local development.
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:3001"}
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		allowedOrigins = strings.Split(raw, ",")
		for i, o := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(o)
		}
	}

	// Rate limiters for sensitive auth endpoints
	// Login / signup: 10 requests per minute per IP
	authLimiter := middleware.NewRateLimiter(10, 60*time.Second)
	// Forgot-password / reset-password: 5 requests per minute per IP
	passwordLimiter := middleware.NewRateLimiter(5, 60*time.Second)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Logging middleware
	r.Use(middleware.LoggingMiddleware)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API version 1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Authentication routes (no auth required, but rate-limited)
		r.Route("/auth", func(r chi.Router) {
			r.With(authLimiter.Middleware).Post("/login", authHandler.Login)
			r.With(authLimiter.Middleware).Post("/signup", authHandler.Signup)
			r.With(passwordLimiter.Middleware).Post("/forgot-password", authHandler.ForgotPassword)
			r.With(passwordLimiter.Middleware).Post("/reset-password", authHandler.ResetPassword)
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			// User profile routes
			r.Get("/profile", authHandler.GetProfile)
			r.Put("/profile", authHandler.UpdateProfile)

			// Income routes
			r.Route("/income", func(r chi.Router) {
				r.Get("/", incomeHandler.GetAllIncomes)
				r.Post("/", incomeHandler.CreateIncome)
				r.Get("/range", incomeHandler.GetIncomeByDateRange)
				r.Get("/{id}", incomeHandler.GetIncome)
				r.Put("/{id}", incomeHandler.UpdateIncome)
				r.Delete("/{id}", incomeHandler.DeleteIncome)
			})

			// Expense routes
			r.Route("/expense", func(r chi.Router) {
				r.Get("/", expenseHandler.GetAllExpenses)
				r.Post("/", expenseHandler.CreateExpense)
				r.Get("/range", expenseHandler.GetExpenseByDateRange)
				r.Get("/breakdown", expenseHandler.GetExpenseCategoryBreakdown)
				r.Get("/{id}", expenseHandler.GetExpense)
				r.Put("/{id}", expenseHandler.UpdateExpense)
				r.Delete("/{id}", expenseHandler.DeleteExpense)
			})

			// Inventory routes
			r.Route("/inventory", func(r chi.Router) {
				r.Get("/", inventoryHandler.GetAllInventory)
				r.Post("/", inventoryHandler.CreateInventoryItem)
				r.Get("/low-stock", inventoryHandler.GetLowStockItems)
				r.Get("/{id}", inventoryHandler.GetInventoryItem)
				r.Put("/{id}", inventoryHandler.UpdateInventoryItem)
				r.Delete("/{id}", inventoryHandler.DeleteInventoryItem)
				r.Patch("/{id}/quantity", inventoryHandler.UpdateQuantity)
			})

			// Analytics routes
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/summary", analyticsHandler.GetFinancialSummary)
				r.Get("/monthly", analyticsHandler.GetMonthlyData)
				r.Get("/expense-breakdown", analyticsHandler.GetExpenseCategoryBreakdown)
			})

			// Mine site info routes
			r.Route("/minesite", func(r chi.Router) {
				r.Get("/", mineSiteHandler.GetMineSiteInfo)
				r.Post("/", mineSiteHandler.CreateOrUpdateMineSiteInfo)
				r.Put("/", mineSiteHandler.CreateOrUpdateMineSiteInfo)
			})

			// User serial number route (available to all authenticated users)
			r.Get("/serial-number", adminHandler.GetUserSerialNumber)

			// Admin routes (require admin role)
			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminMiddleware)
				r.Route("/admin", func(r chi.Router) {
					r.Get("/users", adminHandler.GetAllUsers)
					r.Get("/stats", adminHandler.GetSystemStats)
					r.Get("/trends", adminHandler.GetSystemTrends)
					r.Get("/category-breakdown", adminHandler.GetSystemCategoryBreakdown)
					r.Get("/daily-usage", adminHandler.GetDailyUsage)
				})
			})
		})
	})

	return r
}
