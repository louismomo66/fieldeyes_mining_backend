package handlers

import (
	"encoding/json"
	"mineral/data"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
)

// MineSiteHandler handles mine site information requests
type MineSiteHandler struct {
	MineSiteRepo data.MineSiteInterface
}

// NewMineSiteHandler creates a new MineSiteHandler
func NewMineSiteHandler(mineSiteRepo data.MineSiteInterface) *MineSiteHandler {
	return &MineSiteHandler{
		MineSiteRepo: mineSiteRepo,
	}
}

// MineSiteRequest represents a mine site information request
type MineSiteRequest struct {
	Owner           string   `json:"owner"`
	License         *string  `json:"license,omitempty"`
	Location        string   `json:"location"`
	Size            *float64 `json:"size,omitempty"`
	NumberOfPits    *int     `json:"number_of_pits,omitempty"`
	Commodities     *string  `json:"commodities,omitempty"`
	Equipment       *string  `json:"equipment,omitempty"`
	Employees       *int     `json:"employees,omitempty"`
	EstablishedYear *int     `json:"established_year,omitempty"`
	Contact         *string  `json:"contact,omitempty"`
}

// GetMineSiteInfo retrieves mine site information for the authenticated user
func (h *MineSiteHandler) GetMineSiteInfo(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	info, err := h.MineSiteRepo.GetByUserID(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve mine site information")
		return
	}

	if info == nil {
		// Return empty response if not found
		utils.WriteSuccessResponse(w, "Mine site information not found", nil)
		return
	}

	utils.WriteSuccessResponse(w, "Mine site information retrieved successfully", info)
}

// CreateOrUpdateMineSiteInfo creates or updates mine site information for the authenticated user
func (h *MineSiteHandler) CreateOrUpdateMineSiteInfo(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
		return
	}

	var req MineSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Owner == "" {
		utils.WriteValidationError(w, "Owner is required")
		return
	}
	if req.Location == "" {
		utils.WriteValidationError(w, "Location is required")
		return
	}

	// Check if mine site info already exists
	existingInfo, err := h.MineSiteRepo.GetByUserID(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to check existing mine site information")
		return
	}

	if existingInfo != nil {
		// Update existing record
		existingInfo.Owner = req.Owner
		existingInfo.License = req.License
		existingInfo.Location = req.Location
		existingInfo.Size = req.Size
		existingInfo.NumberOfPits = req.NumberOfPits
		existingInfo.Commodities = req.Commodities
		existingInfo.Equipment = req.Equipment
		existingInfo.Employees = req.Employees
		existingInfo.EstablishedYear = req.EstablishedYear
		existingInfo.Contact = req.Contact

		if err := h.MineSiteRepo.Update(existingInfo); err != nil {
			utils.WriteInternalServerError(w, "Failed to update mine site information")
			return
		}

		utils.WriteSuccessResponse(w, "Mine site information updated successfully", existingInfo)
		return
	}

	// Create new record
	newInfo := &data.MineSiteInfo{
		Owner:           req.Owner,
		License:         req.License,
		Location:        req.Location,
		Size:            req.Size,
		NumberOfPits:    req.NumberOfPits,
		Commodities:     req.Commodities,
		Equipment:       req.Equipment,
		Employees:       req.Employees,
		EstablishedYear: req.EstablishedYear,
		Contact:         req.Contact,
		UserID:          userID,
	}

	id, err := h.MineSiteRepo.Insert(newInfo)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create mine site information")
		return
	}

	newInfo.ID = id
	utils.WriteSuccessResponse(w, "Mine site information created successfully", newInfo)
}
