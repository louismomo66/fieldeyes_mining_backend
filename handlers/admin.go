package handlers

import (
	"fmt"
	"mineral/data"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	UserRepo      data.UserInterface
	IncomeRepo    data.IncomeInterface
	ExpenseRepo   data.ExpenseInterface
	InventoryRepo data.InventoryInterface
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(
	userRepo data.UserInterface,
	incomeRepo data.IncomeInterface,
	expenseRepo data.ExpenseInterface,
	inventoryRepo data.InventoryInterface,
) *AdminHandler {
	return &AdminHandler{
		UserRepo:      userRepo,
		IncomeRepo:    incomeRepo,
		ExpenseRepo:   expenseRepo,
		InventoryRepo: inventoryRepo,
	}
}

// UserCategory represents a user with their category
type UserCategory struct {
	User         *data.User `json:"user"`
	Category     string     `json:"category"`      // "miner", "buyer", "seller", or "unknown"
	SerialNumber string     `json:"serial_number"` // Serial number based on signup date and user ID
}

// GetAllUsers retrieves all users categorized as miners, buyers, or sellers
func (h *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Get all users
	users, err := h.UserRepo.GetAll()
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve users")
		return
	}

	// Categorize users
	var categorizedUsers []UserCategory

	for _, user := range users {
		category := "unknown"

		// Check if user is a miner (has inventory items with minerName or minerSerialNumber)
		inventoryItems, _ := h.InventoryRepo.GetAll(user.ID)
		hasMinerData := false
		for _, item := range inventoryItems {
			if item.MinerName != nil && *item.MinerName != "" {
				hasMinerData = true
				break
			}
			if item.MinerSerialNumber != nil && *item.MinerSerialNumber != "" {
				hasMinerData = true
				break
			}
		}

		// Check if user is a buyer (has income records - they sell to customers)
		incomes, _ := h.IncomeRepo.GetAll(user.ID)
		hasBuyerData := len(incomes) > 0

		// Check if user is a seller (has expense records with suppliers - they buy from suppliers)
		expenses, _ := h.ExpenseRepo.GetAll(user.ID)
		hasSellerData := len(expenses) > 0

		// Determine primary category (miner > buyer > seller)
		if hasMinerData {
			category = "miner"
		} else if hasBuyerData {
			category = "buyer"
		} else if hasSellerData {
			category = "seller"
		}

		// Generate serial number for this user
		serialNumber := GenerateSerialNumber(user.ID, user.CreatedAt)

		categorizedUsers = append(categorizedUsers, UserCategory{
			User:         user,
			Category:     category,
			SerialNumber: serialNumber,
		})
	}

	utils.WriteSuccessResponse(w, "Users retrieved successfully", categorizedUsers)
}

// GenerateSerialNumber generates a serial number for a user based on signup date and user ID
func GenerateSerialNumber(userID uint, signupDate time.Time) string {
	// Format: YYYYMMDD + user ID (3 digits padded)
	// Example: User 1 signed up on 2024-12-04 = 20241204001
	datePart := signupDate.Format("20060102") // YYYYMMDD format
	return datePart + fmt.Sprintf("%03d", userID)
}

// GetUserSerialNumber generates a serial number for a user based on signup date and user ID
func (h *AdminHandler) GetUserSerialNumber(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	// Get user
	user, err := h.UserRepo.GetOne(userID)
	if err != nil {
		utils.WriteNotFoundError(w, "User not found")
		return
	}

	// Generate serial number using the same function
	serialNumber := GenerateSerialNumber(user.ID, user.CreatedAt)

	utils.WriteSuccessResponse(w, "Serial number retrieved successfully", map[string]string{
		"serial_number": serialNumber,
		"user_id":       fmt.Sprintf("%d", user.ID),
		"signup_date":   user.CreatedAt.Format("2006-01-02"),
	})
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers       int     `json:"total_users"`
	TotalMiners      int     `json:"total_miners"`
	TotalBuyers      int     `json:"total_buyers"`
	TotalSellers     int     `json:"total_sellers"`
	TotalIncome      float64 `json:"total_income"`
	TotalExpenses    float64 `json:"total_expenses"`
	TotalProfit      float64 `json:"total_profit"`
	TotalProduction  float64 `json:"total_production"`
	TotalReceivables float64 `json:"total_receivables"`
	TotalPayables    float64 `json:"total_payables"`
	ActiveUsers      int     `json:"active_users"`
}

// GetSystemStats retrieves system-wide statistics for admin
func (h *AdminHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	// Get all users
	users, err := h.UserRepo.GetAll()
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve users")
		return
	}

	stats := SystemStats{
		TotalUsers: len(users),
	}

	// Categorize users and calculate totals
	var totalIncome, totalExpenses, totalReceivables, totalPayables, totalProduction float64
	activeUserSet := make(map[uint]bool)

	for _, user := range users {
		// Get user's income
		incomes, _ := h.IncomeRepo.GetAll(user.ID)
		for _, income := range incomes {
			totalIncome += income.TotalAmount
			totalReceivables += income.AmountDue
			activeUserSet[user.ID] = true
		}

		// Get user's expenses
		expenses, _ := h.ExpenseRepo.GetAll(user.ID)
		for _, expense := range expenses {
			if expense.Category != "fuel" {
				totalExpenses += expense.Amount
				totalPayables += expense.AmountDue
				activeUserSet[user.ID] = true
			}
		}

		// Get user's inventory/production
		inventory, _ := h.InventoryRepo.GetAll(user.ID)
		for _, item := range inventory {
			totalProduction += item.Quantity
			activeUserSet[user.ID] = true
		}
	}

	// Count categories
	for _, user := range users {
		inventoryItems, _ := h.InventoryRepo.GetAll(user.ID)
		hasMinerData := false
		for _, item := range inventoryItems {
			if item.MinerName != nil && *item.MinerName != "" {
				hasMinerData = true
				stats.TotalMiners++
				break
			}
		}
		if !hasMinerData {
			incomes, _ := h.IncomeRepo.GetAll(user.ID)
			if len(incomes) > 0 {
				stats.TotalBuyers++
			} else {
				expenses, _ := h.ExpenseRepo.GetAll(user.ID)
				if len(expenses) > 0 {
					stats.TotalSellers++
				}
			}
		}
	}

	stats.TotalIncome = totalIncome
	stats.TotalExpenses = totalExpenses
	stats.TotalProfit = totalIncome - totalExpenses
	stats.TotalProduction = totalProduction
	stats.TotalReceivables = totalReceivables
	stats.TotalPayables = totalPayables
	stats.ActiveUsers = len(activeUserSet)

	utils.WriteSuccessResponse(w, "System statistics retrieved successfully", stats)
}

// MonthlyTrend represents monthly trend data
type MonthlyTrend struct {
	Month      string  `json:"month"`
	Income     float64 `json:"income"`
	Expenses   float64 `json:"expenses"`
	Profit     float64 `json:"profit"`
	Users      int     `json:"users"`
	Production float64 `json:"production"`
}

// GetSystemTrends retrieves system-wide trends over time
func (h *AdminHandler) GetSystemTrends(w http.ResponseWriter, r *http.Request) {
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

	// Get all users
	users, err := h.UserRepo.GetAll()
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve users")
		return
	}

	// Initialize monthly data
	monthlyData := make(map[string]*MonthlyTrend)
	for i := 1; i <= 12; i++ {
		month := fmt.Sprintf("%d-%02d", year, i)
		monthlyData[month] = &MonthlyTrend{
			Month: month,
		}
	}

	// Aggregate data from all users
	userMonths := make(map[string]map[uint]bool) // Track unique users per month

	for _, user := range users {
		// Get income data
		incomes, _ := h.IncomeRepo.GetAll(user.ID)
		for _, income := range incomes {
			incomeYear := income.Date.Year()
			if incomeYear == year {
				month := income.Date.Format("2006-01")
				if monthlyData[month] != nil {
					monthlyData[month].Income += income.TotalAmount
					if userMonths[month] == nil {
						userMonths[month] = make(map[uint]bool)
					}
					userMonths[month][user.ID] = true
				}
			}
		}

		// Get expense data
		expenses, _ := h.ExpenseRepo.GetAll(user.ID)
		for _, expense := range expenses {
			if expense.Category != "fuel" {
				expenseYear := expense.Date.Year()
				if expenseYear == year {
					month := expense.Date.Format("2006-01")
					if monthlyData[month] != nil {
						monthlyData[month].Expenses += expense.Amount
						if userMonths[month] == nil {
							userMonths[month] = make(map[uint]bool)
						}
						userMonths[month][user.ID] = true
					}
				}
			}
		}

		// Get production data
		inventory, _ := h.InventoryRepo.GetAll(user.ID)
		for _, item := range inventory {
			itemYear := item.LastUpdated.Year()
			if itemYear == year {
				month := item.LastUpdated.Format("2006-01")
				if monthlyData[month] != nil {
					monthlyData[month].Production += item.Quantity
					if userMonths[month] == nil {
						userMonths[month] = make(map[uint]bool)
					}
					userMonths[month][user.ID] = true
				}
			}
		}
	}

	// Calculate profit and user counts
	var result []*MonthlyTrend
	for i := 1; i <= 12; i++ {
		month := fmt.Sprintf("%d-%02d", year, i)
		data := monthlyData[month]
		data.Profit = data.Income - data.Expenses
		if userMonths[month] != nil {
			data.Users = len(userMonths[month])
		}
		result = append(result, data)
	}

	utils.WriteSuccessResponse(w, "System trends retrieved successfully", result)
}

// CategoryBreakdown represents category breakdown for admin
type CategoryBreakdown struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// GetSystemCategoryBreakdown retrieves expense category breakdown across all users
func (h *AdminHandler) GetSystemCategoryBreakdown(w http.ResponseWriter, r *http.Request) {
	// Get all users
	users, err := h.UserRepo.GetAll()
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve users")
		return
	}

	// Aggregate expenses by category
	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)
	var totalAmount float64

	for _, user := range users {
		expenses, _ := h.ExpenseRepo.GetAll(user.ID)
		for _, expense := range expenses {
			if expense.Category != "fuel" {
				category := string(expense.Category)
				categoryTotals[category] += expense.Amount
				categoryCounts[category]++
				totalAmount += expense.Amount
			}
		}
	}

	// Build breakdown
	var breakdown []CategoryBreakdown
	for category, amount := range categoryTotals {
		percentage := 0.0
		if totalAmount > 0 {
			percentage = (amount / totalAmount) * 100
		}
		breakdown = append(breakdown, CategoryBreakdown{
			Category:   category,
			Amount:     amount,
			Count:      categoryCounts[category],
			Percentage: percentage,
		})
	}

	utils.WriteSuccessResponse(w, "Category breakdown retrieved successfully", breakdown)
}

// DailyUsage represents usage statistics for a single day
type DailyUsage struct {
	Date            string  `json:"date"`
	ActiveUsers     int     `json:"active_users"`
	NewUsers        int     `json:"new_users"`
	TotalIncome     float64 `json:"total_income"`
	TotalExpenses   float64 `json:"total_expenses"`
	IncomeCount     int     `json:"income_count"`
	ExpenseCount    int     `json:"expense_count"`
	ProductionCount int     `json:"production_count"`
}

// GetDailyUsage retrieves daily platform usage statistics
func (h *AdminHandler) GetDailyUsage(w http.ResponseWriter, r *http.Request) {
	// Get date range from query parameters (default: last 30 days)
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if startDateStr == "" || endDateStr == "" {
		// Default to last 30 days
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			utils.WriteValidationError(w, "Invalid start_date format. Use YYYY-MM-DD")
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			utils.WriteValidationError(w, "Invalid end_date format. Use YYYY-MM-DD")
			return
		}
	}

	// Get all users
	users, err := h.UserRepo.GetAll()
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve users")
		return
	}

	// Initialize daily data map
	dailyData := make(map[string]*DailyUsage)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dailyData[dateStr] = &DailyUsage{
			Date: dateStr,
		}
	}

	// Track active users per day
	dailyActiveUsers := make(map[string]map[uint]bool)

	// Count new user signups per day
	for _, user := range users {
		signupDate := user.CreatedAt.Format("2006-01-02")
		if data, exists := dailyData[signupDate]; exists {
			data.NewUsers++
		}
	}

	// Aggregate data from all users
	for _, user := range users {
		// Get income data
		incomes, _ := h.IncomeRepo.GetAll(user.ID)
		for _, income := range incomes {
			dateStr := income.Date.Format("2006-01-02")
			if data, exists := dailyData[dateStr]; exists {
				data.TotalIncome += income.TotalAmount
				data.IncomeCount++
				if dailyActiveUsers[dateStr] == nil {
					dailyActiveUsers[dateStr] = make(map[uint]bool)
				}
				dailyActiveUsers[dateStr][user.ID] = true
			}
		}

		// Get expense data
		expenses, _ := h.ExpenseRepo.GetAll(user.ID)
		for _, expense := range expenses {
			dateStr := expense.Date.Format("2006-01-02")
			if data, exists := dailyData[dateStr]; exists {
				data.TotalExpenses += expense.Amount
				data.ExpenseCount++
				if dailyActiveUsers[dateStr] == nil {
					dailyActiveUsers[dateStr] = make(map[uint]bool)
				}
				dailyActiveUsers[dateStr][user.ID] = true
			}
		}

		// Get production data
		inventory, _ := h.InventoryRepo.GetAll(user.ID)
		for _, item := range inventory {
			// Use Date if available, otherwise LastUpdated
			var itemDate time.Time
			if item.Date != nil {
				itemDate = *item.Date
			} else {
				itemDate = item.LastUpdated
			}
			dateStr := itemDate.Format("2006-01-02")
			if data, exists := dailyData[dateStr]; exists {
				data.ProductionCount++
				if dailyActiveUsers[dateStr] == nil {
					dailyActiveUsers[dateStr] = make(map[uint]bool)
				}
				dailyActiveUsers[dateStr][user.ID] = true
			}
		}
	}

	// Set active user counts
	for dateStr, userMap := range dailyActiveUsers {
		if data, exists := dailyData[dateStr]; exists {
			data.ActiveUsers = len(userMap)
		}
	}

	// Convert to sorted slice
	var result []*DailyUsage
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		result = append(result, dailyData[dateStr])
	}

	utils.WriteSuccessResponse(w, "Daily usage retrieved successfully", result)
}
