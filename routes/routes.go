package routes

import (
	"mineral/handlers"
	"mineral/pkg/middleware"
	"net/http"

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
) http.Handler {
	r := chi.NewRouter()

	// CORS configuration using chi's built-in CORS
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002", "http://localhost:8086"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
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
		// Authentication routes (no auth required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/signup", authHandler.Signup)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
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

			// Admin routes (require admin role)
			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminMiddleware)
				// Add admin-specific routes here if needed
			})
		})
	})

	return r
}
