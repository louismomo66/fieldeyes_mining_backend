package routes

import (
	"mineral/data"
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
	complianceHandler *handlers.ComplianceHandler,
	traceabilityHandler *handlers.TraceabilityHandler,
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

	// Security headers on every response
	r.Use(middleware.SecurityHeadersMiddleware)

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

		// Public lot verification (QR target — no login, reg. 53).
		r.Get("/public/verify/{code}", complianceHandler.GetPublicVerification)

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

			// ICGLR compliance routes
			r.Route("/compliance", func(r chi.Router) {
				r.Get("/summary", complianceHandler.GetSummary)

				r.Route("/mine-site-certifications", func(r chi.Router) {
					r.Get("/", complianceHandler.GetMineSiteCertifications)
					r.Post("/", complianceHandler.CreateMineSiteCertification)
					r.Put("/{id}", complianceHandler.UpdateMineSiteCertification)
					r.Delete("/{id}", complianceHandler.DeleteMineSiteCertification)
				})

				r.Route("/coc-lots", func(r chi.Router) {
					r.Get("/", complianceHandler.GetCoCLots)
					r.Post("/", complianceHandler.CreateCoCLot)
					r.Put("/{id}", complianceHandler.UpdateCoCLot)
					r.Delete("/{id}", complianceHandler.DeleteCoCLot)
					r.Post("/{id}/production-records", complianceHandler.LinkProductionRecords)
					r.Get("/{id}/passport", complianceHandler.GetLotPassport)
					// Multi-party custody handover — transporters/exporters typically use this.
					r.With(middleware.RequireChainRole(data.ChainOperator, data.ChainTransporter, data.ChainExporter)).
						Post("/{id}/handover", complianceHandler.HandoverCoCLot)
					r.Get("/production-records", complianceHandler.GetAvailableProductionRecords)
					r.Get("/production-records/by-pit", complianceHandler.GetProductionRecordsByPit)
				})

				// Merge real input lots into one real, linked output lot —
				// distinct from the older /processing-records endpoint, which
				// records a processing note with no link to any actual lot.
				r.Post("/processing-runs", complianceHandler.CreateProcessingRun)

				r.Route("/export-shipments", func(r chi.Router) {
					r.Get("/", complianceHandler.GetExportShipments)
					r.Post("/", complianceHandler.CreateExportShipment)
					r.Put("/{id}", complianceHandler.UpdateExportShipment)
					r.Delete("/{id}", complianceHandler.DeleteExportShipment)
				})

				r.Route("/due-diligence-reports", func(r chi.Router) {
					r.Get("/", complianceHandler.GetDueDiligenceReports)
					r.Post("/", complianceHandler.CreateDueDiligenceReport)
					r.Put("/{id}", complianceHandler.UpdateDueDiligenceReport)
					r.Delete("/{id}", complianceHandler.DeleteDueDiligenceReport)
				})

				r.Route("/third-party-audits", func(r chi.Router) {
					r.Get("/", complianceHandler.GetThirdPartyAudits)
					r.Post("/", complianceHandler.CreateThirdPartyAudit)
					r.Put("/{id}", complianceHandler.UpdateThirdPartyAudit)
					r.Delete("/{id}", complianceHandler.DeleteThirdPartyAudit)
				})

				r.Route("/documents", func(r chi.Router) {
					r.Get("/", complianceHandler.GetComplianceDocuments)
					r.Post("/", complianceHandler.CreateComplianceDocument)
					r.Put("/{id}", complianceHandler.UpdateComplianceDocument)
					r.Delete("/{id}", complianceHandler.DeleteComplianceDocument)
				})
			})

			// Enhanced Traceability routes
			r.Route("/traceability", func(r chi.Router) {
				// Transport records
				r.Route("/transport", func(r chi.Router) {
					r.Get("/", traceabilityHandler.GetTransportRecords)
					r.Post("/", traceabilityHandler.CreateTransportRecord)
					r.Put("/{id}/status", traceabilityHandler.UpdateTransportStatus)
				})

				// Processing records
				r.Route("/processing", func(r chi.Router) {
					r.Get("/", traceabilityHandler.GetProcessingRecords)
					r.Post("/", traceabilityHandler.CreateProcessingRecord)
				})

				// Real-time tracking
				r.Route("/tracking", func(r chi.Router) {
					r.Get("/", traceabilityHandler.GetRealTimeTracking)
					r.Post("/location", traceabilityHandler.UpdateLotLocation)
				})

				// Stakeholders
				r.Route("/stakeholders", func(r chi.Router) {
					r.Get("/", traceabilityHandler.GetStakeholders)
					r.Post("/", traceabilityHandler.CreateStakeholder)
				})

				// Tracking alerts
				r.Route("/alerts", func(r chi.Router) {
					r.Get("/", traceabilityHandler.GetTrackingAlerts)
					r.Post("/", traceabilityHandler.CreateTrackingAlert)
				})
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
