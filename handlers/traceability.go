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

// TraceabilityHandler handles enhanced traceability operations.
type TraceabilityHandler struct {
	TraceabilityRepo data.TraceabilityInterface
}

// NewTraceabilityHandler creates a TraceabilityHandler.
func NewTraceabilityHandler(traceabilityRepo data.TraceabilityInterface) *TraceabilityHandler {
	return &TraceabilityHandler{TraceabilityRepo: traceabilityRepo}
}

// traceUserID extracts the authenticated user ID injected by AuthMiddleware.
// Writes an unauthorized response and returns 0 when missing.
func traceUserID(w http.ResponseWriter, r *http.Request) uint {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
	}
	return userID
}

// Transport Record handlers

func (h *TraceabilityHandler) CreateTransportRecord(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	var requestPayload struct {
		TransportType   string  `json:"transport_type"`
		LicensePlate    string  `json:"license_plate"`
		DriverName      string  `json:"driver_name"`
		DriverLicense   string  `json:"driver_license"`
		VehicleType     string  `json:"vehicle_type"`
		VehicleCapacity float64 `json:"vehicle_capacity"`
		Origin          struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Address   string  `json:"address"`
		} `json:"origin"`
		Destination struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Address   string  `json:"address"`
		} `json:"destination"`
		PlannedDistance    float64   `json:"planned_distance"`
		DepartureTime      time.Time `json:"departure_time"`
		EstimatedDuration  int       `json:"estimated_duration"`
		CoCLotIDs          []uint    `json:"coc_lot_ids"`
		SealNumbers        []string  `json:"seal_numbers"`
		PackagingType      string    `json:"packaging_type"`
		GrossWeight        float64   `json:"gross_weight"`
		NetWeight          float64   `json:"net_weight"`
		HandoverFromName   string    `json:"handover_from_name"`
		HandoverFromID     string    `json:"handover_from_id"`
		HandoverToName     string    `json:"handover_to_name"`
		HandoverToID       string    `json:"handover_to_id"`
		HandoverNotes      string    `json:"handover_notes"`
		SecurityEscort     bool      `json:"security_escort"`
		EscortDetails      string    `json:"escort_details"`
		RouteSecurityLevel string    `json:"route_security_level"`
		TransportPermit    string    `json:"transport_permit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	// Create GPS locations for origin and destination
	originLocation := &data.GPSLocation{
		Latitude:  requestPayload.Origin.Latitude,
		Longitude: requestPayload.Origin.Longitude,
		Address:   &requestPayload.Origin.Address,
		Timestamp: time.Now(),
	}

	destinationLocation := &data.GPSLocation{
		Latitude:  requestPayload.Destination.Latitude,
		Longitude: requestPayload.Destination.Longitude,
		Address:   &requestPayload.Destination.Address,
		Timestamp: time.Now(),
	}

	originID, err := h.TraceabilityRepo.InsertGPSLocation(originLocation)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to save origin location")
		return
	}

	destID, err := h.TraceabilityRepo.InsertGPSLocation(destinationLocation)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to save destination location")
		return
	}

	// Keep the legacy JSON columns populated for backwards compatibility.
	lotIDsJSON, _ := json.Marshal(requestPayload.CoCLotIDs)
	sealNumbersJSON, _ := json.Marshal(requestPayload.SealNumbers)

	record := &data.TransportRecord{
		TransportType:      data.TransportType(requestPayload.TransportType),
		LicensePlate:       requestPayload.LicensePlate,
		DriverName:         requestPayload.DriverName,
		DriverLicense:      requestPayload.DriverLicense,
		VehicleType:        requestPayload.VehicleType,
		VehicleCapacity:    requestPayload.VehicleCapacity,
		OriginID:           originID,
		DestinationID:      destID,
		PlannedDistance:    requestPayload.PlannedDistance,
		DepartureTime:      requestPayload.DepartureTime,
		EstimatedDuration:  requestPayload.EstimatedDuration,
		CoCLotIDs:          string(lotIDsJSON),
		SealNumbers:        string(sealNumbersJSON),
		PackagingType:      requestPayload.PackagingType,
		GrossWeight:        requestPayload.GrossWeight,
		NetWeight:          requestPayload.NetWeight,
		HandoverFromName:   requestPayload.HandoverFromName,
		HandoverFromID:     requestPayload.HandoverFromID,
		HandoverToName:     requestPayload.HandoverToName,
		HandoverToID:       requestPayload.HandoverToID,
		HandoverNotes:      &requestPayload.HandoverNotes,
		SecurityEscort:     requestPayload.SecurityEscort,
		EscortDetails:      &requestPayload.EscortDetails,
		RouteSecurityLevel: requestPayload.RouteSecurityLevel,
		TransportPermit:    &requestPayload.TransportPermit,
		Status:             "scheduled",
		UserID:             userID,
	}

	// Creating the transport record also links the carried CoC lots, writes a
	// custody transfer per lot, moves each lot to in_transit and updates
	// real-time tracking — one entry, whole chain updated.
	id, err := h.TraceabilityRepo.InsertTransportRecordWithLots(record, requestPayload.CoCLotIDs)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create transport record: "+err.Error())
		return
	}
	record.ID = id

	utils.WriteSuccessResponse(w, "Transport record created successfully", record)
}

func (h *TraceabilityHandler) GetTransportRecords(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	records, err := h.TraceabilityRepo.GetTransportRecords(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve transport records")
		return
	}

	utils.WriteSuccessResponse(w, "Transport records retrieved successfully", records)
}

func (h *TraceabilityHandler) UpdateTransportStatus(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid transport record ID")
		return
	}

	var requestPayload struct {
		Status         string     `json:"status"`
		ArrivalTime    *time.Time `json:"arrival_time,omitempty"`
		ActualDistance *float64   `json:"actual_distance,omitempty"`
		ActualDuration *int       `json:"actual_duration,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	record, err := h.TraceabilityRepo.GetTransportRecord(uint(id), userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Transport record not found")
		return
	}

	record.Status = data.TrackingStatus(requestPayload.Status)
	if requestPayload.ArrivalTime != nil {
		record.ArrivalTime = requestPayload.ArrivalTime
	}
	if requestPayload.ActualDistance != nil {
		record.ActualDistance = requestPayload.ActualDistance
	}
	if requestPayload.ActualDuration != nil {
		record.ActualDuration = requestPayload.ActualDuration
	}

	// A delivered/completed transport propagates arrival to every carried lot:
	// custody transfer recorded, lot state moved to stored, tracking updated.
	switch requestPayload.Status {
	case "delivered", "completed", "arrived", "stored":
		err = h.TraceabilityRepo.CompleteTransportDelivery(record)
	default:
		err = h.TraceabilityRepo.UpdateTransportRecord(record)
	}
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to update transport status")
		return
	}

	utils.WriteSuccessResponse(w, "Transport status updated successfully", record)
}

// Processing Record handlers

func (h *TraceabilityHandler) CreateProcessingRecord(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	var requestPayload struct {
		FacilityID   string   `json:"facility_id"`
		FacilityName string   `json:"facility_name"`
		ProcessType  []string `json:"process_type"`
		InputBatches []struct {
			InventoryItemID uint    `json:"inventory_item_id"`
			InputWeight     float64 `json:"input_weight"`
			InputGrade      float64 `json:"input_grade"`
			InputUnit       string  `json:"input_unit"`
		} `json:"input_batches"`
		Equipment   []string  `json:"equipment"`
		Parameters  string    `json:"parameters"`
		Duration    int       `json:"duration"`
		Operator    string    `json:"operator"`
		Supervisor  string    `json:"supervisor"`
		StartTime   time.Time `json:"start_time"`
		OutputItems []struct {
			Name    string  `json:"name"`
			Weight  float64 `json:"weight"`
			Grade   float64 `json:"grade"`
			Unit    string  `json:"unit"`
			Product string  `json:"product"`
		} `json:"output_items"`
		Yield            float64 `json:"yield"`
		Recovery         float64 `json:"recovery"`
		WasteGenerated   float64 `json:"waste_generated"`
		SamplesCollected int     `json:"samples_collected"`
		AssayResults     string  `json:"assay_results"`
		QualityNotes     string  `json:"quality_notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	processTypeJSON, _ := json.Marshal(requestPayload.ProcessType)
	inputBatchesJSON, _ := json.Marshal(requestPayload.InputBatches)
	equipmentJSON, _ := json.Marshal(requestPayload.Equipment)
	outputItemsJSON, _ := json.Marshal(requestPayload.OutputItems)

	record := &data.ProcessingRecord{
		FacilityID:       requestPayload.FacilityID,
		FacilityName:     requestPayload.FacilityName,
		ProcessType:      string(processTypeJSON),
		InputBatches:     string(inputBatchesJSON),
		Equipment:        string(equipmentJSON),
		Parameters:       requestPayload.Parameters,
		Duration:         requestPayload.Duration,
		Operator:         requestPayload.Operator,
		Supervisor:       requestPayload.Supervisor,
		StartTime:        requestPayload.StartTime,
		OutputItems:      string(outputItemsJSON),
		Yield:            requestPayload.Yield,
		Recovery:         &requestPayload.Recovery,
		WasteGenerated:   requestPayload.WasteGenerated,
		SamplesCollected: requestPayload.SamplesCollected,
		AssayResults:     &requestPayload.AssayResults,
		QualityNotes:     &requestPayload.QualityNotes,
		Status:           "scheduled",
		UserID:           userID,
	}

	id, err := h.TraceabilityRepo.InsertProcessingRecord(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create processing record")
		return
	}
	record.ID = id

	utils.WriteSuccessResponse(w, "Processing record created successfully", record)
}

func (h *TraceabilityHandler) GetProcessingRecords(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	records, err := h.TraceabilityRepo.GetProcessingRecords(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve processing records")
		return
	}

	utils.WriteSuccessResponse(w, "Processing records retrieved successfully", records)
}

// Real-time Tracking handlers

func (h *TraceabilityHandler) GetRealTimeTracking(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	records, err := h.TraceabilityRepo.GetRealTimeTracking(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve real-time tracking data")
		return
	}

	utils.WriteSuccessResponse(w, "Real-time tracking data retrieved successfully", records)
}

func (h *TraceabilityHandler) UpdateLotLocation(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	var requestPayload struct {
		LotID            uint    `json:"lot_id"`
		LotType          string  `json:"lot_type"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
		Address          string  `json:"address"`
		CurrentCustodian string  `json:"current_custodian"`
		Status           string  `json:"status"`
		NextDestination  string  `json:"next_destination,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	location := &data.GPSLocation{
		Latitude:  requestPayload.Latitude,
		Longitude: requestPayload.Longitude,
		Address:   &requestPayload.Address,
		Timestamp: time.Now(),
	}

	locationID, err := h.TraceabilityRepo.InsertGPSLocation(location)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to save location")
		return
	}

	existingTracking, err := h.TraceabilityRepo.GetRealTimeTrackingByLot(
		requestPayload.LotID, requestPayload.LotType, userID)

	if err != nil {
		tracking := &data.RealTimeTracking{
			LotID:             requestPayload.LotID,
			LotType:           requestPayload.LotType,
			CurrentLocationID: &locationID,
			CurrentCustodian:  requestPayload.CurrentCustodian,
			Status:            data.TrackingStatus(requestPayload.Status),
			NextDestination:   &requestPayload.NextDestination,
			UserID:            userID,
		}
		_, err = h.TraceabilityRepo.InsertRealTimeTracking(tracking)
	} else {
		existingTracking.CurrentLocationID = &locationID
		existingTracking.CurrentCustodian = requestPayload.CurrentCustodian
		existingTracking.Status = data.TrackingStatus(requestPayload.Status)
		if requestPayload.NextDestination != "" {
			existingTracking.NextDestination = &requestPayload.NextDestination
		}
		err = h.TraceabilityRepo.UpdateRealTimeTracking(existingTracking)
	}

	if err != nil {
		utils.WriteInternalServerError(w, "Failed to update lot location")
		return
	}

	utils.WriteSuccessResponse(w, "Location updated successfully", nil)
}

// Stakeholder handlers

func (h *TraceabilityHandler) GetStakeholders(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	records, err := h.TraceabilityRepo.GetStakeholders(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve stakeholders")
		return
	}

	utils.WriteSuccessResponse(w, "Stakeholders retrieved successfully", records)
}

func (h *TraceabilityHandler) CreateStakeholder(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	var requestPayload struct {
		Name             string `json:"name"`
		Type             string `json:"type"`
		Email            string `json:"email"`
		Phone            string `json:"phone"`
		Address          string `json:"address"`
		ComplianceStatus string `json:"compliance_status"`
		Licenses         []struct {
			LicenseType   string    `json:"license_type"`
			LicenseNumber string    `json:"license_number"`
			IssuedBy      string    `json:"issued_by"`
			ValidFrom     time.Time `json:"valid_from"`
			ValidUntil    time.Time `json:"valid_until"`
			Status        string    `json:"status"`
		} `json:"licenses"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	licensesJSON, _ := json.Marshal(requestPayload.Licenses)

	record := &data.Stakeholder{
		Name:             requestPayload.Name,
		Type:             requestPayload.Type,
		Email:            &requestPayload.Email,
		Phone:            &requestPayload.Phone,
		Address:          &requestPayload.Address,
		Licenses:         string(licensesJSON),
		ComplianceStatus: data.ComplianceStatus(requestPayload.ComplianceStatus),
		UserID:           userID,
	}

	id, err := h.TraceabilityRepo.InsertStakeholder(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create stakeholder")
		return
	}
	record.ID = id

	utils.WriteSuccessResponse(w, "Stakeholder created successfully", record)
}

// Tracking alerts handlers

func (h *TraceabilityHandler) GetTrackingAlerts(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	records, err := h.TraceabilityRepo.GetTrackingAlerts(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve tracking alerts")
		return
	}

	utils.WriteSuccessResponse(w, "Tracking alerts retrieved successfully", records)
}

func (h *TraceabilityHandler) CreateTrackingAlert(w http.ResponseWriter, r *http.Request) {
	userID := traceUserID(w, r)
	if userID == 0 {
		return
	}

	var requestPayload struct {
		LotID     uint   `json:"lot_id"`
		LotType   string `json:"lot_type"`
		AlertType string `json:"alert_type"`
		Severity  string `json:"severity"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	record := &data.TrackingAlert{
		LotID:     requestPayload.LotID,
		LotType:   requestPayload.LotType,
		AlertType: data.AlertType(requestPayload.AlertType),
		Severity:  data.AlertSeverity(requestPayload.Severity),
		Message:   requestPayload.Message,
		UserID:    userID,
	}

	id, err := h.TraceabilityRepo.InsertTrackingAlert(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create tracking alert")
		return
	}
	record.ID = id

	utils.WriteSuccessResponse(w, "Tracking alert created successfully", record)
}
