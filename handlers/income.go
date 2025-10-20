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

// IncomeHandler handles income-related requests
type IncomeHandler struct {
	IncomeRepo data.IncomeInterface
}

// NewIncomeHandler creates a new IncomeHandler
func NewIncomeHandler(incomeRepo data.IncomeInterface) *IncomeHandler {
	return &IncomeHandler{
		IncomeRepo: incomeRepo,
	}
}

// CreateIncomeRequest represents a create income request
type CreateIncomeRequest struct {
	Date            string  `json:"date"`
	MineralType     string  `json:"mineral_type"`
	Quantity        float64 `json:"quantity"`
	Unit            string  `json:"unit"`
	PricePerUnit    float64 `json:"price_per_unit"`
	CustomerName    string  `json:"customer_name"`
	CustomerContact string  `json:"customer_contact"`
	PaymentStatus   string  `json:"payment_status"`
	AmountPaid      float64 `json:"amount_paid"`
	Notes           string  `json:"notes,omitempty"`
}

// UpdateIncomeRequest represents an update income request
type UpdateIncomeRequest struct {
	Date            string  `json:"date"`
	MineralType     string  `json:"mineral_type"`
	Quantity        float64 `json:"quantity"`
	Unit            string  `json:"unit"`
	PricePerUnit    float64 `json:"price_per_unit"`
	CustomerName    string  `json:"customer_name"`
	CustomerContact string  `json:"customer_contact"`
	PaymentStatus   string  `json:"payment_status"`
	AmountPaid      float64 `json:"amount_paid"`
	Notes           string  `json:"notes,omitempty"`
}

// GetAllIncomes retrieves all income records for the authenticated user
func (h *IncomeHandler) GetAllIncomes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	incomes, err := h.IncomeRepo.GetAll(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve income records")
		return
	}

	utils.WriteSuccessResponse(w, "Income records retrieved successfully", incomes)
}

// GetIncome retrieves a specific income record
func (h *IncomeHandler) GetIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid income ID")
		return
	}

	income, err := h.IncomeRepo.GetOne(uint(id), userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Income record not found")
		return
	}

	utils.WriteSuccessResponse(w, "Income record retrieved successfully", income)
}

// CreateIncome creates a new income record
func (h *IncomeHandler) CreateIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	var req CreateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate input
	if !utils.ValidateRequired(req.Date) {
		utils.WriteValidationError(w, "Date is required")
		return
	}
	if !utils.ValidateRequired(req.MineralType) {
		utils.WriteValidationError(w, "Mineral type is required")
		return
	}
	if !utils.ValidatePositiveNumber(req.Quantity) {
		utils.WriteValidationError(w, "Quantity must be positive")
		return
	}
	if !utils.ValidateRequired(req.Unit) {
		utils.WriteValidationError(w, "Unit is required")
		return
	}
	if !utils.ValidatePositiveNumber(req.PricePerUnit) {
		utils.WriteValidationError(w, "Price per unit must be positive")
		return
	}
	if !utils.ValidateRequired(req.CustomerName) {
		utils.WriteValidationError(w, "Customer name is required")
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

	// Validate mineral type
	mineralType := data.MineralType(req.MineralType)
	if mineralType != data.MineralGold && mineralType != data.MineralCopper &&
		mineralType != data.MineralCobalt && mineralType != data.MineralDiamond &&
		mineralType != data.MineralOther {
		utils.WriteValidationError(w, "Invalid mineral type")
		return
	}

	// Validate payment status
	paymentStatus := data.PaymentStatus(req.PaymentStatus)
	if paymentStatus != data.PaymentPaid && paymentStatus != data.PaymentUnpaid &&
		paymentStatus != data.PaymentPartial {
		utils.WriteValidationError(w, "Invalid payment status")
		return
	}

	// Create income record
	income := &data.Income{
		Date:            date,
		MineralType:     mineralType,
		Quantity:        req.Quantity,
		Unit:            req.Unit,
		PricePerUnit:    req.PricePerUnit,
		CustomerName:    req.CustomerName,
		CustomerContact: req.CustomerContact,
		PaymentStatus:   paymentStatus,
		AmountPaid:      req.AmountPaid,
		UserID:          userID,
	}
	if req.Notes != "" {
		income.Notes = &req.Notes
	}

	incomeID, err := h.IncomeRepo.Insert(income)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create income record")
		return
	}

	income.ID = incomeID
	utils.WriteSuccessResponse(w, "Income record created successfully", income)
}

// UpdateIncome updates an existing income record
func (h *IncomeHandler) UpdateIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid income ID")
		return
	}

	var req UpdateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Get existing income record
	income, err := h.IncomeRepo.GetOne(uint(id), userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Income record not found")
		return
	}

	// Validate and update fields
	if !utils.ValidateRequired(req.Date) {
		utils.WriteValidationError(w, "Date is required")
		return
	}
	if !utils.ValidateRequired(req.MineralType) {
		utils.WriteValidationError(w, "Mineral type is required")
		return
	}
	if !utils.ValidatePositiveNumber(req.Quantity) {
		utils.WriteValidationError(w, "Quantity must be positive")
		return
	}
	if !utils.ValidateRequired(req.Unit) {
		utils.WriteValidationError(w, "Unit is required")
		return
	}
	if !utils.ValidatePositiveNumber(req.PricePerUnit) {
		utils.WriteValidationError(w, "Price per unit must be positive")
		return
	}
	if !utils.ValidateRequired(req.CustomerName) {
		utils.WriteValidationError(w, "Customer name is required")
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

	// Validate mineral type
	mineralType := data.MineralType(req.MineralType)
	if mineralType != data.MineralGold && mineralType != data.MineralCopper &&
		mineralType != data.MineralCobalt && mineralType != data.MineralDiamond &&
		mineralType != data.MineralOther {
		utils.WriteValidationError(w, "Invalid mineral type")
		return
	}

	// Validate payment status
	paymentStatus := data.PaymentStatus(req.PaymentStatus)
	if paymentStatus != data.PaymentPaid && paymentStatus != data.PaymentUnpaid &&
		paymentStatus != data.PaymentPartial {
		utils.WriteValidationError(w, "Invalid payment status")
		return
	}

	// Calculate total amount and amount due
	totalAmount := req.Quantity * req.PricePerUnit
	amountDue := totalAmount - req.AmountPaid

	// Update income record
	income.Date = date
	income.MineralType = mineralType
	income.Quantity = req.Quantity
	income.Unit = req.Unit
	income.PricePerUnit = req.PricePerUnit
	income.TotalAmount = totalAmount
	income.CustomerName = req.CustomerName
	income.CustomerContact = req.CustomerContact
	income.PaymentStatus = paymentStatus
	income.AmountPaid = req.AmountPaid
	income.AmountDue = amountDue
	if req.Notes != "" {
		income.Notes = &req.Notes
	} else {
		income.Notes = nil
	}

	err = h.IncomeRepo.Update(income)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to update income record")
		return
	}

	utils.WriteSuccessResponse(w, "Income record updated successfully", income)
}

// DeleteIncome deletes an income record
func (h *IncomeHandler) DeleteIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid income ID")
		return
	}

	err = h.IncomeRepo.Delete(uint(id), userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to delete income record")
		return
	}

	utils.WriteSuccessResponse(w, "Income record deleted successfully", nil)
}

// GetIncomeByDateRange retrieves income records within a date range
func (h *IncomeHandler) GetIncomeByDateRange(w http.ResponseWriter, r *http.Request) {
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

	incomes, err := h.IncomeRepo.GetByDateRange(userID, startDate, endDate)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve income records")
		return
	}

	utils.WriteSuccessResponse(w, "Income records retrieved successfully", incomes)
}
