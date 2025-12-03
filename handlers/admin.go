package handlers

import (
	"mineral/data"
	"mineral/pkg/utils"
	"net/http"
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
	User     *data.User `json:"user"`
	Category string     `json:"category"` // "miner", "buyer", "seller", or "unknown"
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

		categorizedUsers = append(categorizedUsers, UserCategory{
			User:     user,
			Category: category,
		})
	}

	utils.WriteSuccessResponse(w, "Users retrieved successfully", categorizedUsers)
}

