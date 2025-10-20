package handlers

import (
	"encoding/json"
	"mineral/data"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// InventoryHandler handles inventory-related requests
type InventoryHandler struct {
	InventoryRepo data.InventoryInterface
}

// NewInventoryHandler creates a new InventoryHandler
func NewInventoryHandler(inventoryRepo data.InventoryInterface) *InventoryHandler {
	return &InventoryHandler{
		InventoryRepo: inventoryRepo,
	}
}

// CreateInventoryRequest represents a create inventory request
type CreateInventoryRequest struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	MinStockLevel float64 `json:"min_stock_level"`
	CurrentValue  float64 `json:"current_value"`
}

// UpdateInventoryRequest represents an update inventory request
type UpdateInventoryRequest struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	MinStockLevel float64 `json:"min_stock_level"`
	CurrentValue  float64 `json:"current_value"`
}

// UpdateQuantityRequest represents an update quantity request
type UpdateQuantityRequest struct {
	Quantity float64 `json:"quantity"`
}

// GetAllInventory retrieves all inventory items for the authenticated user
func (h *InventoryHandler) GetAllInventory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	items, err := h.InventoryRepo.GetAll(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve inventory items")
		return
	}

	utils.WriteSuccessResponse(w, "Inventory items retrieved successfully", items)
}

// GetInventoryItem retrieves a specific inventory item
func (h *InventoryHandler) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid inventory item ID")
		return
	}

	item, err := h.InventoryRepo.GetOne(uint(id), userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Inventory item not found")
		return
	}

	utils.WriteSuccessResponse(w, "Inventory item retrieved successfully", item)
}

// CreateInventoryItem creates a new inventory item
func (h *InventoryHandler) CreateInventoryItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	var req CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate input
	if !utils.ValidateRequired(req.Name) {
		utils.WriteValidationError(w, "Name is required")
		return
	}
	if !utils.ValidateRequired(req.Type) {
		utils.WriteValidationError(w, "Type is required")
		return
	}
	if req.Type != "mineral" && req.Type != "supply" {
		utils.WriteValidationError(w, "Type must be either 'mineral' or 'supply'")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.Quantity) {
		utils.WriteValidationError(w, "Quantity cannot be negative")
		return
	}
	if !utils.ValidateRequired(req.Unit) {
		utils.WriteValidationError(w, "Unit is required")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.MinStockLevel) {
		utils.WriteValidationError(w, "Minimum stock level cannot be negative")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.CurrentValue) {
		utils.WriteValidationError(w, "Current value cannot be negative")
		return
	}

	// Create inventory item
	item := &data.InventoryItem{
		Name:          req.Name,
		Type:          req.Type,
		Quantity:      req.Quantity,
		Unit:          req.Unit,
		MinStockLevel: req.MinStockLevel,
		CurrentValue:  req.CurrentValue,
		UserID:        userID,
	}

	itemID, err := h.InventoryRepo.Insert(item)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create inventory item")
		return
	}

	item.ID = itemID
	utils.WriteSuccessResponse(w, "Inventory item created successfully", item)
}

// UpdateInventoryItem updates an existing inventory item
func (h *InventoryHandler) UpdateInventoryItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid inventory item ID")
		return
	}

	var req UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Get existing inventory item
	item, err := h.InventoryRepo.GetOne(uint(id), userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Inventory item not found")
		return
	}

	// Validate and update fields
	if !utils.ValidateRequired(req.Name) {
		utils.WriteValidationError(w, "Name is required")
		return
	}
	if !utils.ValidateRequired(req.Type) {
		utils.WriteValidationError(w, "Type is required")
		return
	}
	if req.Type != "mineral" && req.Type != "supply" {
		utils.WriteValidationError(w, "Type must be either 'mineral' or 'supply'")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.Quantity) {
		utils.WriteValidationError(w, "Quantity cannot be negative")
		return
	}
	if !utils.ValidateRequired(req.Unit) {
		utils.WriteValidationError(w, "Unit is required")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.MinStockLevel) {
		utils.WriteValidationError(w, "Minimum stock level cannot be negative")
		return
	}
	if !utils.ValidateNonNegativeNumber(req.CurrentValue) {
		utils.WriteValidationError(w, "Current value cannot be negative")
		return
	}

	// Update inventory item
	item.Name = req.Name
	item.Type = req.Type
	item.Quantity = req.Quantity
	item.Unit = req.Unit
	item.MinStockLevel = req.MinStockLevel
	item.CurrentValue = req.CurrentValue

	err = h.InventoryRepo.Update(item)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to update inventory item")
		return
	}

	utils.WriteSuccessResponse(w, "Inventory item updated successfully", item)
}

// DeleteInventoryItem deletes an inventory item
func (h *InventoryHandler) DeleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid inventory item ID")
		return
	}

	err = h.InventoryRepo.Delete(uint(id), userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to delete inventory item")
		return
	}

	utils.WriteSuccessResponse(w, "Inventory item deleted successfully", nil)
}

// GetLowStockItems retrieves items that are below minimum stock level
func (h *InventoryHandler) GetLowStockItems(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	items, err := h.InventoryRepo.GetLowStockItems(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve low stock items")
		return
	}

	utils.WriteSuccessResponse(w, "Low stock items retrieved successfully", items)
}

// UpdateQuantity updates the quantity of an inventory item
func (h *InventoryHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid inventory item ID")
		return
	}

	var req UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate quantity
	if !utils.ValidateNonNegativeNumber(req.Quantity) {
		utils.WriteValidationError(w, "Quantity cannot be negative")
		return
	}

	err = h.InventoryRepo.UpdateQuantity(uint(id), userID, req.Quantity)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to update quantity")
		return
	}

	// Get updated item
	item, err := h.InventoryRepo.GetOne(uint(id), userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve updated item")
		return
	}

	utils.WriteSuccessResponse(w, "Quantity updated successfully", item)
}
