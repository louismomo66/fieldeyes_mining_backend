package handlers

import (
	"fmt"
	"mineral/data"
	"mineral/pkg/cache"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

// Analytics cache TTLs
const (
	summaryCacheTTL  = 5 * time.Minute
	monthlyCacheTTL  = 10 * time.Minute
	breakdownCacheTTL = 5 * time.Minute
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	IncomeRepo  data.IncomeInterface
	ExpenseRepo data.ExpenseInterface
	Cache       *cache.Client
}

// NewAnalyticsHandler creates a new AnalyticsHandler
func NewAnalyticsHandler(incomeRepo data.IncomeInterface, expenseRepo data.ExpenseInterface, cache *cache.Client) *AnalyticsHandler {
	return &AnalyticsHandler{
		IncomeRepo:  incomeRepo,
		ExpenseRepo: expenseRepo,
		Cache:       cache,
	}
}

// GetFinancialSummary retrieves financial summary for the authenticated user.
// Results are cached in Redis for 5 minutes per user.
func (h *AnalyticsHandler) GetFinancialSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	cacheKey := fmt.Sprintf("analytics:summary:%d", userID)

	// Try cache first
	var summary data.FinancialSummary
	if hit, _ := h.Cache.Get(cacheKey, &summary); hit {
		utils.WriteSuccessResponse(w, "Financial summary retrieved successfully", &summary)
		return
	}

	// Cache miss — query DB
	incomeSummary, err := h.IncomeRepo.GetFinancialSummary(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve income summary")
		return
	}

	expenseSummary, err := h.ExpenseRepo.GetFinancialSummary(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve expense summary")
		return
	}

	netProfit := incomeSummary.TotalIncome - expenseSummary.TotalExpenses
	var profitMargin float64
	if incomeSummary.TotalIncome > 0 {
		profitMargin = (netProfit / incomeSummary.TotalIncome) * 100
	}

	result := &data.FinancialSummary{
		TotalIncome:      incomeSummary.TotalIncome,
		TotalExpenses:    expenseSummary.TotalExpenses,
		NetProfit:        netProfit,
		TotalReceivables: incomeSummary.TotalReceivables,
		TotalPayables:    expenseSummary.TotalPayables,
		ProfitMargin:     profitMargin,
	}

	// Store in cache
	h.Cache.Set(cacheKey, result, summaryCacheTTL)

	utils.WriteSuccessResponse(w, "Financial summary retrieved successfully", result)
}

// GetMonthlyData retrieves monthly financial data for a year.
// Results are cached in Redis for 10 minutes per user per year.
func (h *AnalyticsHandler) GetMonthlyData(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	yearStr := r.URL.Query().Get("year")
	var year int
	if yearStr == "" {
		year = time.Now().Year()
	} else {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil || year < 2000 || year > 3000 {
			utils.WriteValidationError(w, "Invalid year")
			return
		}
	}

	cacheKey := fmt.Sprintf("analytics:monthly:%d:%d", userID, year)

	// Try cache first
	var cached []*data.MonthlyData
	if hit, _ := h.Cache.Get(cacheKey, &cached); hit {
		utils.WriteSuccessResponse(w, "Monthly data retrieved successfully", cached)
		return
	}

	// Cache miss — query DB
	incomeData, err := h.IncomeRepo.GetMonthlyData(userID, year)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve monthly income data")
		return
	}

	expenseData, err := h.ExpenseRepo.GetMonthlyData(userID, year)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve monthly expense data")
		return
	}

	monthlyData := make(map[string]*data.MonthlyData)
	for _, item := range incomeData {
		if monthlyData[item.Month] == nil {
			monthlyData[item.Month] = &data.MonthlyData{Month: item.Month}
		}
		monthlyData[item.Month].Income = item.Income
	}
	for _, item := range expenseData {
		if monthlyData[item.Month] == nil {
			monthlyData[item.Month] = &data.MonthlyData{Month: item.Month}
		}
		monthlyData[item.Month].Expenses = item.Expenses
	}

	var result []*data.MonthlyData
	for _, d := range monthlyData {
		d.Profit = d.Income - d.Expenses
		result = append(result, d)
	}

	// Store in cache
	h.Cache.Set(cacheKey, result, monthlyCacheTTL)

	utils.WriteSuccessResponse(w, "Monthly data retrieved successfully", result)
}

// GetExpenseCategoryBreakdown retrieves expense breakdown by category.
// Results are cached in Redis for 5 minutes per user.
func (h *AnalyticsHandler) GetExpenseCategoryBreakdown(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	cacheKey := fmt.Sprintf("analytics:breakdown:%d", userID)

	// Try cache first
	var cached []*data.CategoryBreakdown
	if hit, _ := h.Cache.Get(cacheKey, &cached); hit {
		utils.WriteSuccessResponse(w, "Expense breakdown retrieved successfully", cached)
		return
	}

	// Cache miss — query DB
	breakdown, err := h.ExpenseRepo.GetCategoryBreakdown(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve expense breakdown")
		return
	}

	// Store in cache
	h.Cache.Set(cacheKey, breakdown, breakdownCacheTTL)

	utils.WriteSuccessResponse(w, "Expense breakdown retrieved successfully", breakdown)
}
