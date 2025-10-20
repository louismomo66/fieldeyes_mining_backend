package handlers

import (
	"mineral/data"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	IncomeRepo  data.IncomeInterface
	ExpenseRepo data.ExpenseInterface
}

// NewAnalyticsHandler creates a new AnalyticsHandler
func NewAnalyticsHandler(incomeRepo data.IncomeInterface, expenseRepo data.ExpenseInterface) *AnalyticsHandler {
	return &AnalyticsHandler{
		IncomeRepo:  incomeRepo,
		ExpenseRepo: expenseRepo,
	}
}

// GetFinancialSummary retrieves financial summary for the authenticated user
func (h *AnalyticsHandler) GetFinancialSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	// Get income summary
	incomeSummary, err := h.IncomeRepo.GetFinancialSummary(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve income summary")
		return
	}

	// Get expense summary
	expenseSummary, err := h.ExpenseRepo.GetFinancialSummary(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve expense summary")
		return
	}

	// Calculate net profit
	netProfit := incomeSummary.TotalIncome - expenseSummary.TotalExpenses

	// Calculate profit margin
	var profitMargin float64
	if incomeSummary.TotalIncome > 0 {
		profitMargin = (netProfit / incomeSummary.TotalIncome) * 100
	}

	// Combine summaries
	summary := &data.FinancialSummary{
		TotalIncome:      incomeSummary.TotalIncome,
		TotalExpenses:    expenseSummary.TotalExpenses,
		NetProfit:        netProfit,
		TotalReceivables: incomeSummary.TotalReceivables,
		TotalPayables:    expenseSummary.TotalReceivables, // Assuming this field exists in expense summary
		ProfitMargin:     profitMargin,
	}

	utils.WriteSuccessResponse(w, "Financial summary retrieved successfully", summary)
}

// GetMonthlyData retrieves monthly financial data for a year
func (h *AnalyticsHandler) GetMonthlyData(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	// Get year from query parameter, default to current year
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

	// Get monthly income data
	incomeData, err := h.IncomeRepo.GetMonthlyData(userID, year)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve monthly income data")
		return
	}

	// Get monthly expense data
	expenseData, err := h.ExpenseRepo.GetMonthlyData(userID, year)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve monthly expense data")
		return
	}

	// Combine the data
	monthlyData := make(map[string]*data.MonthlyData)

	// Add income data
	for _, item := range incomeData {
		if monthlyData[item.Month] == nil {
			monthlyData[item.Month] = &data.MonthlyData{Month: item.Month}
		}
		monthlyData[item.Month].Income = item.Income
	}

	// Add expense data
	for _, item := range expenseData {
		if monthlyData[item.Month] == nil {
			monthlyData[item.Month] = &data.MonthlyData{Month: item.Month}
		}
		monthlyData[item.Month].Expenses = item.Expenses
	}

	// Calculate profit for each month
	var result []*data.MonthlyData
	for _, data := range monthlyData {
		data.Profit = data.Income - data.Expenses
		result = append(result, data)
	}

	utils.WriteSuccessResponse(w, "Monthly data retrieved successfully", result)
}

// GetExpenseCategoryBreakdown retrieves expense breakdown by category
func (h *AnalyticsHandler) GetExpenseCategoryBreakdown(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	breakdown, err := h.ExpenseRepo.GetCategoryBreakdown(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve expense breakdown")
		return
	}

	utils.WriteSuccessResponse(w, "Expense breakdown retrieved successfully", breakdown)
}
