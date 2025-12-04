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
	Category     string     `json:"category"` // "miner", "buyer", "seller", or "unknown"
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
	TotalUsers        int     `json:"total_users"`
	TotalMiners       int     `json:"total_miners"`
	TotalBuyers       int     `json:"total_buyers"`
	TotalSellers      int     `json:"total_sellers"`
	TotalIncome       float64 `json:"total_income"`
	TotalExpenses     float64 `json:"total_expenses"`
	TotalProfit       float64 `json:"total_profit"`
	TotalProduction   float64 `json:"total_production"`
	TotalReceivables  float64 `json:"total_receivables"`
	TotalPayables     float64 `json:"total_payables"`
	ActiveUsers       int     `json:"active_users"`
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
	Month     string  `json:"month"`
	Income    float64 `json:"income"`
	Expenses  float64 `json:"expenses"`
	Profit    float64 `json:"profit"`
	Users     int     `json:"users"`
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
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
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
			Category:    category,
			Amount:      amount,
			Count:       categoryCounts[category],
			Percentage:  percentage,
		})
	}

	utils.WriteSuccessResponse(w, "Category breakdown retrieved successfully", breakdown)
}

