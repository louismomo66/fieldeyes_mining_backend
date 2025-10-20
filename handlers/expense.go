package handlers

import (
	"encoding/json"
	"mineral/data"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// ExpenseHandler handles expense-related requests
type ExpenseHandler struct {
	ExpenseRepo data.ExpenseInterface
}

// NewExpenseHandler creates a new ExpenseHandler
func NewExpenseHandler(expenseRepo data.ExpenseInterface) *ExpenseHandler {
	return &ExpenseHandler{
		ExpenseRepo: expenseRepo,
	}
}

// CreateExpenseRequest represents a create expense request
type CreateExpenseRequest struct {
	Date            string  `json:"date"`
	Category        string  `json:"category"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	SupplierName    string  `json:"supplier_name"`
	SupplierContact string  `json:"supplier_contact,omitempty"`
	PaymentStatus   string  `json:"payment_status"`
	AmountPaid      float64 `json:"amount_paid"`
	Notes           string  `json:"notes,omitempty"`
}

// UpdateExpenseRequest represents an update expense request
type UpdateExpenseRequest struct {
	Date            string  `json:"date"`
	Category        string  `json:"category"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	SupplierName    string  `json:"supplier_name"`
	SupplierContact string  `json:"supplier_contact,omitempty"`
	PaymentStatus   string  `json:"payment_status"`
	AmountPaid      float64 `json:"amount_paid"`
	Notes           string  `json:"notes,omitempty"`
}

// GetAllExpenses retrieves all expense records for the authenticated user
func (h *ExpenseHandler) GetAllExpenses(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	expenses, err := h.ExpenseRepo.GetAll(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve expense records")
		return
	}

	utils.WriteSuccessResponse(w, "Expense records retrieved successfully", expenses)
}

// GetExpense retrieves a specific expense record
func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid expense ID")
		return
	}

	expense, err := h.ExpenseRepo.GetOne(uint(id), userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Expense record not found")
		return
	}

	utils.WriteSuccessResponse(w, "Expense record retrieved successfully", expense)
}

// CreateExpense creates a new expense record
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	var req CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate input
	if !utils.ValidateRequired(req.Date) {
		utils.WriteValidationError(w, "Date is required")
		return
	}
	if !utils.ValidateRequired(req.Category) {
		utils.WriteValidationError(w, "Category is required")
		return
	}
	if !utils.ValidateRequired(req.Description) {
		utils.WriteValidationError(w, "Description is required")
		return
	}
	if !utils.ValidatePositiveNumber(req.Amount) {
		utils.WriteValidationError(w, "Amount must be positive")
		return
	}
	if !utils.ValidateRequired(req.SupplierName) {
		utils.WriteValidationError(w, "Supplier name is required")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.AmountPaid) {
		utils.WriteValidationError(w, "Amount paid cannot be negative")
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		utils.WriteValidationError(w, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Validate category
	category := data.ExpenseCategory(req.Category)
	if category != data.ExpenseEquipment && category != data.ExpenseLabor &&
		category != data.ExpenseChemicals && category != data.ExpenseFuel &&
		category != data.ExpenseMaintenance && category != data.ExpenseTransport &&
		category != data.ExpenseOther {
		utils.WriteValidationError(w, "Invalid expense category")
		return
	}

	// Validate payment status
	paymentStatus := data.PaymentStatus(req.PaymentStatus)
	if paymentStatus != data.PaymentPaid && paymentStatus != data.PaymentUnpaid &&
		paymentStatus != data.PaymentPartial {
		utils.WriteValidationError(w, "Invalid payment status")
		return
	}

	// Create expense record
	expense := &data.Expense{
		Date:          date,
		Category:      category,
		Description:   req.Description,
		Amount:        req.Amount,
		SupplierName:  req.SupplierName,
		PaymentStatus: paymentStatus,
		AmountPaid:    req.AmountPaid,
		UserID:        userID,
	}
	if req.SupplierContact != "" {
		expense.SupplierContact = &req.SupplierContact
	}
	if req.Notes != "" {
		expense.Notes = &req.Notes
	}

	expenseID, err := h.ExpenseRepo.Insert(expense)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create expense record")
		return
	}

	expense.ID = expenseID
	utils.WriteSuccessResponse(w, "Expense record created successfully", expense)
}

// UpdateExpense updates an existing expense record
func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid expense ID")
		return
	}

	var req UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Get existing expense record
	expense, err := h.ExpenseRepo.GetOne(uint(id), userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Expense record not found")
		return
	}

	// Validate and update fields
	if !utils.ValidateRequired(req.Date) {
		utils.WriteValidationError(w, "Date is required")
		return
	}
	if !utils.ValidateRequired(req.Category) {
		utils.WriteValidationError(w, "Category is required")
		return
	}
	if !utils.ValidateRequired(req.Description) {
		utils.WriteValidationError(w, "Description is required")
		return
	}
	if !utils.ValidatePositiveNumber(req.Amount) {
		utils.WriteValidationError(w, "Amount must be positive")
		return
	}
	if !utils.ValidateRequired(req.SupplierName) {
		utils.WriteValidationError(w, "Supplier name is required")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.AmountPaid) {
		utils.WriteValidationError(w, "Amount paid cannot be negative")
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		utils.WriteValidationError(w, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Validate category
	category := data.ExpenseCategory(req.Category)
	if category != data.ExpenseEquipment && category != data.ExpenseLabor &&
		category != data.ExpenseChemicals && category != data.ExpenseFuel &&
		category != data.ExpenseMaintenance && category != data.ExpenseTransport &&
		category != data.ExpenseOther {
		utils.WriteValidationError(w, "Invalid expense category")
		return
	}

	// Validate payment status
	paymentStatus := data.PaymentStatus(req.PaymentStatus)
	if paymentStatus != data.PaymentPaid && paymentStatus != data.PaymentUnpaid &&
		paymentStatus != data.PaymentPartial {
		utils.WriteValidationError(w, "Invalid payment status")
		return
	}

	// Calculate amount due
	amountDue := req.Amount - req.AmountPaid

	// Update expense record
	expense.Date = date
	expense.Category = category
	expense.Description = req.Description
	expense.Amount = req.Amount
	expense.SupplierName = req.SupplierName
	expense.PaymentStatus = paymentStatus
	expense.AmountPaid = req.AmountPaid
	expense.AmountDue = amountDue
	if req.SupplierContact != "" {
		expense.SupplierContact = &req.SupplierContact
	} else {
		expense.SupplierContact = nil
	}
	if req.Notes != "" {
		expense.Notes = &req.Notes
	} else {
		expense.Notes = nil
	}

	err = h.ExpenseRepo.Update(expense)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to update expense record")
		return
	}

	utils.WriteSuccessResponse(w, "Expense record updated successfully", expense)
}

// DeleteExpense deletes an expense record
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid expense ID")
		return
	}

	err = h.ExpenseRepo.Delete(uint(id), userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to delete expense record")
		return
	}

	utils.WriteSuccessResponse(w, "Expense record deleted successfully", nil)
}

// GetExpenseByDateRange retrieves expense records within a date range
func (h *ExpenseHandler) GetExpenseByDateRange(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		utils.WriteValidationError(w, "Start date and end date are required")
		return
	}

	expenses, err := h.ExpenseRepo.GetByDateRange(userID, startDate, endDate)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve expense records")
		return
	}

	utils.WriteSuccessResponse(w, "Expense records retrieved successfully", expenses)
}

// GetExpenseCategoryBreakdown retrieves expense breakdown by category
func (h *ExpenseHandler) GetExpenseCategoryBreakdown(w http.ResponseWriter, r *http.Request) {
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
