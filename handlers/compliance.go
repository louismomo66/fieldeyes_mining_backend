package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mineral/data"
	"mineral/pkg/middleware"
	"mineral/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// ComplianceHandler handles ICGLR compliance records.
type ComplianceHandler struct {
	ComplianceRepo   data.ComplianceInterface
	TraceabilityRepo data.TraceabilityInterface
}

// NewComplianceHandler creates a ComplianceHandler.
func NewComplianceHandler(complianceRepo data.ComplianceInterface, traceabilityRepo data.TraceabilityInterface) *ComplianceHandler {
	return &ComplianceHandler{ComplianceRepo: complianceRepo, TraceabilityRepo: traceabilityRepo}
}

type MineSiteCertificationRequest struct {
	MineSiteInfoID      *uint   `json:"mine_site_info_id,omitempty"`
	MineSiteName        string  `json:"mine_site_name"`
	RMDIdentification   *string `json:"rmd_identification,omitempty"`
	InspectionDate      *string `json:"inspection_date,omitempty"`
	InspectorName       *string `json:"inspector_name,omitempty"`
	Status              string  `json:"status"`
	ReportReference     *string `json:"report_reference,omitempty"`
	Findings            *string `json:"findings,omitempty"`
	CorrectiveActions   *string `json:"corrective_actions,omitempty"`
	GracePeriodEndsAt   *string `json:"grace_period_ends_at,omitempty"`
	FollowUpRequestedAt *string `json:"follow_up_requested_at,omitempty"`
	FollowUpDueAt       *string `json:"follow_up_due_at,omitempty"`
	IsSuspended         bool    `json:"is_suspended"`
	SuspensionLiftedAt  *string `json:"suspension_lifted_at,omitempty"`
}

// SourceLotInput describes one source lot blended into a mixed lot
// (Schedule 5, paragraph 3 of S.I. No. 23 of 2023).
type SourceLotInput struct {
	SourceLotID         uint     `json:"source_lot_id"`
	WeightContributed   float64  `json:"weight_contributed"`
	GradeContributed    *float64 `json:"grade_contributed,omitempty"`
	PurchaseOrderNumber *string  `json:"purchase_order_number,omitempty"`
}

type CoCLotRequest struct {
	LotNumber              string   `json:"lot_number"`
	ProductionRecordIDs    *[]uint  `json:"production_record_ids,omitempty"`
	ParentLotNumbers       *string  `json:"parent_lot_numbers,omitempty"`
	SourceLots             []SourceLotInput `json:"source_lots,omitempty"`
	MineSiteCertificationID *uint   `json:"mine_site_certification_id,omitempty"`
	PurchaseOrderNumber    *string  `json:"purchase_order_number,omitempty"`
	PurchaseDate           *string  `json:"purchase_date,omitempty"`
	AcceptedBy             *string  `json:"accepted_by,omitempty"`
	MineralType            string   `json:"mineral_type"`
	OreType                *string  `json:"ore_type,omitempty"`
	Weight                 float64  `json:"weight"`
	Unit                   string   `json:"unit"`
	Grade                  *string  `json:"grade,omitempty"`
	GradeValue             *float64 `json:"grade_value,omitempty"`
	GradeUnit              *string  `json:"grade_unit,omitempty"`
	NumberOfSacks          *int     `json:"number_of_sacks,omitempty"`
	SourceMineSite         string   `json:"source_mine_site"`
	MineSiteStatus         string   `json:"mine_site_status"`
	MineOperatorName       *string  `json:"mine_operator_name,omitempty"`
	MinerName              *string  `json:"miner_name,omitempty"`
	MinerNationalID        *string  `json:"miner_national_id,omitempty"`
	ArtisanalLicenseNumber *string  `json:"artisanal_license_number,omitempty"`
	CoCSystem              string   `json:"coc_system"`
	SealNumber             *string  `json:"seal_number,omitempty"`
	RegisteredAt           *string  `json:"registered_at,omitempty"`
	SealedAt               *string `json:"sealed_at,omitempty"`
	ShippedAt              *string `json:"shipped_at,omitempty"`
	TransporterName        *string `json:"transporter_name,omitempty"`
	TransportRoute         *string `json:"transport_route,omitempty"`
	CurrentCustodian       *string `json:"current_custodian,omitempty"`
	UpstreamActors         *string `json:"upstream_actors,omitempty"`
	TaxesFeesRoyalties     *string `json:"taxes_fees_royalties,omitempty"`
	VerificationOfficer    *string `json:"verification_officer,omitempty"`
	DocumentationReference *string `json:"documentation_reference,omitempty"`
	IsExported             bool    `json:"is_exported"`
}

type ExportShipmentRequest struct {
	ExporterLotNumber      string  `json:"exporter_lot_number"`
	ExporterName           string  `json:"exporter_name"`
	ExporterLicenseNumber  *string `json:"exporter_license_number,omitempty"`
	ExporterStatus         string  `json:"exporter_status"`
	CustomerName           string  `json:"customer_name"`
	CustomerAddress        *string `json:"customer_address,omitempty"`
	DestinationCountry     string  `json:"destination_country"`
	MaterialDescription    string  `json:"material_description"`
	Weight                 float64 `json:"weight"`
	Unit                   string  `json:"unit"`
	Grade                  *string `json:"grade,omitempty"`
	IncomingLotNumbers     string  `json:"incoming_lot_numbers"`
	IncomingLotWeights     *string `json:"incoming_lot_weights,omitempty"`
	CoCLotIDs              []uint  `json:"coc_lot_ids,omitempty"`
	TaxesFeesRoyalties     *string `json:"taxes_fees_royalties,omitempty"`
	SealedAt               *string `json:"sealed_at,omitempty"`
	ShippedAt              *string `json:"shipped_at,omitempty"`
	TransporterName        *string `json:"transporter_name,omitempty"`
	TransportRoute         *string `json:"transport_route,omitempty"`
	AuthorisedOfficer      *string `json:"authorised_officer,omitempty"`
	ApplicationStatus      string  `json:"application_status"`
	ICGLRCertificateNumber *string `json:"icglr_certificate_number,omitempty"`
	ICGLRCertificateFile   *string `json:"icglr_certificate_file,omitempty"`
	CertificateIssuedAt    *string `json:"certificate_issued_at,omitempty"`
	CertificateExpiresAt   *string `json:"certificate_expires_at,omitempty"`
}

type DueDiligenceReportRequest struct {
	ReportingPeriodStart   string  `json:"reporting_period_start"`
	ReportingPeriodEnd     string  `json:"reporting_period_end"`
	Frequency              string  `json:"frequency"`
	ResponsiblePerson      string  `json:"responsible_person"`
	MineralChainPolicy     *string `json:"mineral_chain_policy,omitempty"`
	ManagementSystem       *string `json:"management_system,omitempty"`
	RiskAssessmentSummary  *string `json:"risk_assessment_summary,omitempty"`
	RiskMitigationPlan     *string `json:"risk_mitigation_plan,omitempty"`
	GrievanceMechanism     *string `json:"grievance_mechanism,omitempty"`
	SupplierCapacity       *string `json:"supplier_capacity,omitempty"`
	GovernmentPayments     *string `json:"government_payments,omitempty"`
	BeneficialOwnership    *string `json:"beneficial_ownership,omitempty"`
	PublishedAt            *string `json:"published_at,omitempty"`
	SubmittedToDirectorate bool    `json:"submitted_to_directorate"`
	SubmittedAt            *string `json:"submitted_at,omitempty"`
	AttachmentReference    *string `json:"attachment_reference,omitempty"`
}

type ThirdPartyAuditRequest struct {
	ExporterName        string  `json:"exporter_name"`
	AuditorName         *string `json:"auditor_name,omitempty"`
	AuditRequestedAt    *string `json:"audit_requested_at,omitempty"`
	AuditStartedAt      *string `json:"audit_started_at,omitempty"`
	AuditCompletedAt    *string `json:"audit_completed_at,omitempty"`
	Status              string  `json:"status"`
	StatusExpiresAt     *string `json:"status_expires_at,omitempty"`
	ReportReference     *string `json:"report_reference,omitempty"`
	Findings            *string `json:"findings,omitempty"`
	CorrectiveActions   *string `json:"corrective_actions,omitempty"`
	FollowUpRequestedAt *string `json:"follow_up_requested_at,omitempty"`
	FollowUpDueAt       *string `json:"follow_up_due_at,omitempty"`
}

type ComplianceDocumentRequest struct {
	DocumentType  string  `json:"document_type"`
	Title         string  `json:"title"`
	Reference     string  `json:"reference"`
	RelatedEntity string  `json:"related_entity,omitempty"`
	RelatedID     *uint   `json:"related_id,omitempty"`
	IssuedAt      *string `json:"issued_at,omitempty"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

func (h *ComplianceHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}

	certs, err := h.ComplianceRepo.GetMineSiteCertifications(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve mine site certifications")
		return
	}
	lots, err := h.ComplianceRepo.GetCoCLots(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve CoC lots")
		return
	}
	exports, err := h.ComplianceRepo.GetExportShipments(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve export shipments")
		return
	}
	reports, err := h.ComplianceRepo.GetDueDiligenceReports(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve due diligence reports")
		return
	}
	audits, err := h.ComplianceRepo.GetThirdPartyAudits(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve third party audits")
		return
	}
	docs, err := h.ComplianceRepo.GetComplianceDocuments(userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve compliance documents")
		return
	}

	blockedLots := 0
	for _, lot := range lots {
		if lot.MineSiteStatus == data.StatusRed {
			blockedLots++
		}
	}

	utils.WriteSuccessResponse(w, "Compliance summary retrieved successfully", map[string]interface{}{
		"mine_site_certifications": len(certs),
		"coc_lots":                 len(lots),
		"export_shipments":         len(exports),
		"due_diligence_reports":    len(reports),
		"third_party_audits":       len(audits),
		"documents":                len(docs),
		"blocked_red_status_lots":  blockedLots,
	})
}

func (h *ComplianceHandler) GetMineSiteCertifications(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	records, err := h.ComplianceRepo.GetMineSiteCertifications(userID)
	writeList(w, records, err, "Mine site certifications retrieved successfully")
}

func (h *ComplianceHandler) CreateMineSiteCertification(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	var req MineSiteCertificationRequest
	if !decodeJSON(w, r, &req) || !validStatus(w, req.Status) {
		return
	}
	if req.MineSiteName == "" {
		utils.WriteValidationError(w, "Mine site name is required")
		return
	}
	record := mineSiteCertificationFromRequest(req, userID)
	id, err := h.ComplianceRepo.InsertMineSiteCertification(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create mine site certification")
		return
	}
	record.ID = id
	utils.WriteSuccessResponse(w, "Mine site certification created successfully", record)
}

func (h *ComplianceHandler) UpdateMineSiteCertification(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	var req MineSiteCertificationRequest
	if !decodeJSON(w, r, &req) || !validStatus(w, req.Status) {
		return
	}
	record, err := h.ComplianceRepo.GetMineSiteCertification(id, userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Mine site certification not found")
		return
	}
	applyMineSiteCertificationRequest(record, req)
	if err := h.ComplianceRepo.UpdateMineSiteCertification(record); err != nil {
		utils.WriteInternalServerError(w, "Failed to update mine site certification")
		return
	}
	utils.WriteSuccessResponse(w, "Mine site certification updated successfully", record)
}

func (h *ComplianceHandler) DeleteMineSiteCertification(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	err := h.ComplianceRepo.DeleteMineSiteCertification(id, userID)
	writeDeleted(w, err, "Mine site certification deleted successfully")
}

func (h *ComplianceHandler) GetCoCLots(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	records, err := h.ComplianceRepo.GetCoCLots(userID)
	writeList(w, records, err, "CoC lots retrieved successfully")
}

func (h *ComplianceHandler) CreateCoCLot(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	var req CoCLotRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// Derive the mine site status from the linked certification record —
	// the certification is the single source of truth, never re-typed.
	var cert *data.MineSiteCertification
	if req.MineSiteCertificationID != nil {
		var err error
		cert, err = h.ComplianceRepo.GetMineSiteCertification(*req.MineSiteCertificationID, userID)
		if err != nil {
			utils.WriteNotFoundError(w, "Mine site certification not found")
			return
		}
	} else if req.MineSiteStatus == "" {
		if latest, err := h.ComplianceRepo.GetLatestCertificationForMineSite(userID); err == nil {
			cert = latest
		}
	}
	if cert != nil {
		certID := cert.ID
		req.MineSiteCertificationID = &certID
		req.MineSiteStatus = string(cert.Status)
		if req.SourceMineSite == "" {
			req.SourceMineSite = cert.MineSiteName
		}
	}
	if req.MineSiteStatus == "" {
		req.MineSiteStatus = string(data.StatusBlue)
	}
	if !validStatus(w, req.MineSiteStatus) {
		return
	}

	// Validate and auto-aggregate mixed-lot composition (Schedule 5 para 3).
	composition, compWeight, err := h.buildComposition(req, userID)
	if err != nil {
		utils.WriteValidationError(w, err.Error())
		return
	}
	if req.Weight <= 0 && compWeight > 0 {
		req.Weight = compWeight
	}

	// Some clients (the workflow-first lot registration) derive everything
	// about a lot and never collect a lot number from the user — mint one
	// rather than reject the request. A client that does collect its own
	// (the manual compliance form) is left untouched.
	if req.LotNumber == "" {
		req.LotNumber = generateLotNumber()
	}

	prodIDs := []uint{}
	if req.ProductionRecordIDs != nil {
		prodIDs = *req.ProductionRecordIDs
	}

	if err := validateCoCLot(req, len(prodIDs) > 0); err != nil {
		utils.WriteValidationError(w, err.Error())
		return
	}
	record := coCLotFromRequest(req, userID)
	// Assign a public verification code so the lot is QR-verifiable (reg. 53).
	code := generateVerifyCode()
	record.VerifyCode = &code

	id, err := h.ComplianceRepo.InsertCoCLotWithLinks(record, prodIDs, composition)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create CoC lot: "+err.Error())
		return
	}
	record.ID = id
	created, err := h.ComplianceRepo.GetCoCLot(id, userID)
	if err == nil {
		utils.WriteSuccessResponse(w, "CoC lot created successfully", created)
		return
	}
	utils.WriteSuccessResponse(w, "CoC lot created successfully", record)
}

// ProcessingRunRequest describes a real processing run: consume named input
// lots, produce one output lot. Distinct from the legacy "quick" processing
// forms, which recorded a free-typed input/output weight with no link to any
// actual CoCLot — those still exist and still work, this is additive.
type ProcessingRunRequest struct {
	InputLotIDs      []uint   `json:"input_lot_ids"`
	FacilityName     string   `json:"facility_name"`
	ProcessType      []string `json:"process_type"`
	Operator         string   `json:"operator"`
	Supervisor       string   `json:"supervisor"`
	Duration         int      `json:"duration"`
	SamplesCollected int      `json:"samples_collected"`
	QualityNotes     string   `json:"quality_notes"`

	// Output lot. Weight is optional — left blank, it's input total minus
	// waste, the same "derive, don't retype" rule lot registration already
	// follows. Mineral, site and CoC system are never asked for: they're
	// inherited from the inputs, since an output batch can't have a
	// different origin than what went into it.
	OutputWeight     float64  `json:"output_weight"`
	WasteGenerated   float64  `json:"waste_generated"`
	OutputGradeValue *float64 `json:"output_grade_value,omitempty"`
	OutputGradeUnit  *string  `json:"output_grade_unit,omitempty"`
	SealNumber       *string  `json:"seal_number,omitempty"`
}

// statusRank orders site status from most to least restrictive, so a run
// blending lots from several sites inherits the worst one — a clean input
// can't launder a red-flagged one by sharing a batch with it.
var statusRank = map[data.ComplianceStatus]int{
	data.StatusBlue:   0,
	data.StatusYellow: 1,
	data.StatusGreen:  2,
}

// CreateProcessingRun is the real merge-lots-into-one-batch feature: pick
// real input lots, get a real output lot that carries their combined
// traceability forward — not a note that happens to mention a weight.
func (h *ComplianceHandler) CreateProcessingRun(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	var req ProcessingRunRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.InputLotIDs) == 0 {
		utils.WriteValidationError(w, "Select at least one input lot")
		return
	}
	if strings.TrimSpace(req.FacilityName) == "" {
		utils.WriteValidationError(w, "Facility is required")
		return
	}

	inputs, err := h.ComplianceRepo.GetCoCLotsByIDs(req.InputLotIDs, userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to load input lots")
		return
	}
	if len(inputs) != len(req.InputLotIDs) {
		utils.WriteValidationError(w, "One or more input lots were not found, or are not in your custody")
		return
	}

	mineral := inputs[0].MineralType
	totalInputKg := 0.0
	var chosenCert *uint
	var chosenSite string
	var chosenCoCSystem string
	bestRank := -1
	for _, lot := range inputs {
		if lot.IsExported {
			utils.WriteValidationError(w, fmt.Sprintf("Lot %s has already been exported and cannot be processed", lot.LotNumber))
			return
		}
		if lot.MineralType != mineral {
			utils.WriteValidationError(w, "Input lots must all be the same designated mineral")
			return
		}
		if lot.MineSiteStatus == data.StatusRed {
			utils.WriteValidationError(w, fmt.Sprintf("Lot %s originates from a red-status mine site and cannot be processed", lot.LotNumber))
			return
		}
		factor, ok := data.UnitToKg[strings.ToLower(lot.Unit)]
		if !ok {
			utils.WriteValidationError(w, fmt.Sprintf("Lot %s is in a unit that can't be converted to kg (%s)", lot.LotNumber, lot.Unit))
			return
		}
		totalInputKg += lot.Weight * factor

		if rank, ok := statusRank[lot.MineSiteStatus]; ok && (bestRank == -1 || rank < bestRank) {
			bestRank = rank
			chosenCert = lot.MineSiteCertificationID
			chosenSite = lot.SourceMineSite
			chosenCoCSystem = lot.CoCSystem
		}
	}

	outputWeight := req.OutputWeight
	if outputWeight <= 0 {
		outputWeight = totalInputKg - req.WasteGenerated
	}
	if outputWeight <= 0 {
		utils.WriteValidationError(w, "Output weight must be greater than zero once waste is subtracted")
		return
	}
	if outputWeight+req.WasteGenerated > totalInputKg+0.0001 {
		utils.WriteValidationError(w, fmt.Sprintf(
			"Output plus waste (%.2f kg) cannot exceed the total input weight (%.2f kg)", outputWeight+req.WasteGenerated, totalInputKg))
		return
	}

	outputStatus := data.StatusBlue
	if chosenSite == "" {
		chosenSite = inputs[0].SourceMineSite
		chosenCoCSystem = inputs[0].CoCSystem
	}
	for status, rank := range statusRank {
		if rank == bestRank {
			outputStatus = status
		}
	}

	lotNumber := generateLotNumber()
	verifyCode := generateVerifyCode()
	outputLot := &data.CoCLot{
		LotNumber:               lotNumber,
		VerifyCode:              &verifyCode,
		MineralType:             mineral,
		Weight:                  outputWeight,
		Unit:                    "kg",
		GradeValue:              req.OutputGradeValue,
		GradeUnit:               req.OutputGradeUnit,
		SourceMineSite:          chosenSite,
		MineSiteStatus:          outputStatus,
		MineSiteCertificationID: chosenCert,
		CoCSystem:               chosenCoCSystem,
		SealNumber:              req.SealNumber,
		UserID:                  userID,
	}

	processTypeJSON, _ := json.Marshal(req.ProcessType)
	inputBatchesJSON, _ := json.Marshal(inputs)
	outputItemsJSON, _ := json.Marshal([]map[string]interface{}{
		{"name": lotNumber, "weight": outputWeight, "unit": "kg"},
	})
	now := time.Now()
	rec := &data.ProcessingRecord{
		FacilityID:       "",
		FacilityName:     strings.TrimSpace(req.FacilityName),
		ProcessType:      string(processTypeJSON),
		InputBatches:     string(inputBatchesJSON),
		Equipment:        "[]",
		Parameters:       "",
		Duration:         req.Duration,
		Operator:         req.Operator,
		Supervisor:       req.Supervisor,
		StartTime:        now,
		EndTime:          &now, // recorded after the fact, not scheduled ahead of it
		OutputItems:      string(outputItemsJSON),
		Yield:            (outputWeight / totalInputKg) * 100,
		WasteGenerated:   req.WasteGenerated,
		SamplesCollected: req.SamplesCollected,
		QualityNotes:     &req.QualityNotes,
		// TrackingStatus is really a lot's supply-chain position, reused here
		// for lack of a dedicated processing-record status; "scheduled" is
		// what both existing (unlinked) processing forms already use.
		Status: "scheduled",
	}

	result, err := h.ComplianceRepo.CreateProcessingRun(userID, req.InputLotIDs, outputLot, rec)
	if err != nil {
		utils.WriteValidationError(w, err.Error())
		return
	}
	utils.WriteSuccessResponse(w, fmt.Sprintf("Processing run recorded — output lot %s created", result.OutputLot.LotNumber), result)
}

// buildComposition validates the source lots of a mixed lot and returns the
// composition rows plus the summed contributed weight.
func (h *ComplianceHandler) buildComposition(req CoCLotRequest, userID uint) ([]data.LotComposition, float64, error) {
	if len(req.SourceLots) == 0 {
		return nil, 0, nil
	}
	ids := make([]uint, 0, len(req.SourceLots))
	for _, s := range req.SourceLots {
		if s.SourceLotID == 0 {
			return nil, 0, validationErr("Each source lot requires a source_lot_id")
		}
		if s.WeightContributed <= 0 {
			return nil, 0, validationErr("Each source lot requires the weight it contributed to the mixed lot")
		}
		ids = append(ids, s.SourceLotID)
	}
	lots, err := h.ComplianceRepo.GetCoCLotsByIDs(ids, userID)
	if err != nil || len(lots) != len(ids) {
		return nil, 0, validationErr("One or more source lots were not found")
	}
	lotByID := map[uint]*data.CoCLot{}
	for _, lot := range lots {
		lotByID[lot.ID] = lot
	}
	total := 0.0
	composition := make([]data.LotComposition, 0, len(req.SourceLots))
	for _, s := range req.SourceLots {
		lot := lotByID[s.SourceLotID]
		if lot.IsExported {
			return nil, 0, validationErr(fmt.Sprintf("Source lot %s has already been exported", lot.LotNumber))
		}
		status := lot.MineSiteStatus
		if lot.MineSiteCertification != nil {
			status = lot.MineSiteCertification.Status
		}
		if status == data.StatusRed {
			return nil, 0, validationErr(fmt.Sprintf("Source lot %s originates from a red-status mine site and cannot be blended", lot.LotNumber))
		}
		if s.WeightContributed > lot.Weight {
			return nil, 0, validationErr(fmt.Sprintf("Source lot %s cannot contribute more weight (%.2f) than it contains (%.2f)", lot.LotNumber, s.WeightContributed, lot.Weight))
		}
		grade := s.GradeContributed
		if grade == nil {
			grade = lot.GradeValue
		}
		composition = append(composition, data.LotComposition{
			SourceLotID:         s.SourceLotID,
			WeightContributed:   s.WeightContributed,
			GradeContributed:    grade,
			PurchaseOrderNumber: s.PurchaseOrderNumber,
		})
		total += s.WeightContributed
	}
	return composition, total, nil
}

func (h *ComplianceHandler) UpdateCoCLot(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	var req CoCLotRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	// Re-derive the status from a linked certification, if any.
	if req.MineSiteCertificationID != nil {
		if cert, err := h.ComplianceRepo.GetMineSiteCertification(*req.MineSiteCertificationID, userID); err == nil {
			req.MineSiteStatus = string(cert.Status)
			if req.SourceMineSite == "" {
				req.SourceMineSite = cert.MineSiteName
			}
		}
	}
	if req.MineSiteStatus == "" {
		req.MineSiteStatus = string(data.StatusBlue)
	}
	if !validStatus(w, req.MineSiteStatus) {
		return
	}
	if err := validateCoCLot(req, false); err != nil {
		utils.WriteValidationError(w, err.Error())
		return
	}
	record, err := h.ComplianceRepo.GetCoCLot(id, userID)
	if err != nil {
		utils.WriteNotFoundError(w, "CoC lot not found")
		return
	}
	applyCoCLotRequest(record, req)
	if err := h.ComplianceRepo.UpdateCoCLot(record); err != nil {
		utils.WriteInternalServerError(w, "Failed to update CoC lot")
		return
	}
	utils.WriteSuccessResponse(w, "CoC lot updated successfully", record)
}

func (h *ComplianceHandler) DeleteCoCLot(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	err := h.ComplianceRepo.DeleteCoCLot(id, userID)
	writeDeleted(w, err, "CoC lot deleted successfully")
}

func (h *ComplianceHandler) GetExportShipments(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	records, err := h.ComplianceRepo.GetExportShipments(userID)
	writeList(w, records, err, "Export shipments retrieved successfully")
}

func (h *ComplianceHandler) CreateExportShipment(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	var req ExportShipmentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ExporterStatus == "" {
		req.ExporterStatus = string(data.StatusBlue)
	}
	if !validStatus(w, req.ExporterStatus) {
		return
	}

	// Build the shipment from the selected CoC lots: lot numbers, weights and
	// grades are aggregated automatically instead of being re-typed, and the
	// legal export conditions (reg. 37 & 39, S.I. No. 23 of 2023) are enforced.
	var lots []*data.CoCLot
	var lotTotal float64
	if len(req.CoCLotIDs) > 0 {
		var err error
		lots, err = h.ComplianceRepo.GetCoCLotsByIDs(req.CoCLotIDs, userID)
		if err != nil || len(lots) != len(req.CoCLotIDs) {
			utils.WriteValidationError(w, "One or more CoC lots were not found")
			return
		}
		lotNumbers := make([]string, 0, len(lots))
		type lotWeight struct {
			LotNumber string  `json:"lot_number"`
			Weight    float64 `json:"weight"`
			Unit      string  `json:"unit"`
		}
		lotWeights := make([]lotWeight, 0, len(lots))
		mineralTypes := map[string]bool{}
		for _, lot := range lots {
			if lot.IsExported {
				utils.WriteValidationError(w, fmt.Sprintf("Lot %s has already been exported", lot.LotNumber))
				return
			}
			if lot.ExportShipmentID != nil {
				utils.WriteValidationError(w, fmt.Sprintf("Lot %s is already assigned to another export shipment", lot.LotNumber))
				return
			}
			status := lot.MineSiteStatus
			if lot.MineSiteCertification != nil {
				status = lot.MineSiteCertification.Status
			}
			if status == data.StatusRed {
				utils.WriteValidationError(w, fmt.Sprintf("Lot %s originates from a red-status mine site and cannot be exported (reg. 37, S.I. No. 23 of 2023)", lot.LotNumber))
				return
			}
			lotTotal += lot.Weight
			lotNumbers = append(lotNumbers, lot.LotNumber)
			lotWeights = append(lotWeights, lotWeight{LotNumber: lot.LotNumber, Weight: lot.Weight, Unit: lot.Unit})
			mineralTypes[string(lot.MineralType)] = true
		}

		// Auto-fill from the lots — single entry, no repetition.
		if req.IncomingLotNumbers == "" {
			req.IncomingLotNumbers = strings.Join(lotNumbers, ", ")
		}
		if req.IncomingLotWeights == nil {
			if b, err := json.Marshal(lotWeights); err == nil {
				s := string(b)
				req.IncomingLotWeights = &s
			}
		}
		if req.Weight <= 0 {
			req.Weight = lotTotal
		}
		if req.Unit == "" {
			req.Unit = lots[0].Unit
		}
		if req.MaterialDescription == "" {
			types := make([]string, 0, len(mineralTypes))
			for t := range mineralTypes {
				types = append(types, t)
			}
			req.MaterialDescription = fmt.Sprintf("%s — %d lot(s), %.2f %s", strings.Join(types, ", "), len(lots), lotTotal, req.Unit)
		}
		if req.Grade == nil && len(lots) == 1 && lots[0].Grade != nil {
			req.Grade = lots[0].Grade
		}

		// Mass balance: a shipment cannot weigh more than the lots feeding it.
		if lotTotal > 0 && req.Weight > lotTotal*1.02 {
			utils.WriteValidationError(w, fmt.Sprintf("Declared weight (%.2f) exceeds the combined weight of the linked lots (%.2f) — possible infiltration of untraced material", req.Weight, lotTotal))
			return
		}
	}

	if err := validateExportShipment(req); err != nil {
		utils.WriteValidationError(w, err.Error())
		return
	}
	record := exportShipmentFromRequest(req, userID)
	if lotTotal > 0 {
		disc := (req.Weight - lotTotal) / lotTotal * 100
		record.WeightDiscrepancyPct = &disc
	}
	id, err := h.ComplianceRepo.InsertExportShipment(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create export shipment")
		return
	}
	record.ID = id

	if len(req.CoCLotIDs) > 0 {
		if err := h.ComplianceRepo.AttachLotsToShipment(id, req.CoCLotIDs, userID); err != nil {
			utils.WriteInternalServerError(w, "Shipment created but failed to link CoC lots")
			return
		}
		// Raise a compliance alert when the declared weight deviates from the
		// combined lot weight by more than 2% (mass-balance discrepancy).
		if record.WeightDiscrepancyPct != nil && (*record.WeightDiscrepancyPct > 2 || *record.WeightDiscrepancyPct < -2) && h.TraceabilityRepo != nil {
			for _, lot := range lots {
				alert := &data.TrackingAlert{
					LotID:     lot.ID,
					LotType:   "coc_lot",
					AlertType: data.AlertCompliance,
					Severity:  data.SeverityHigh,
					Message:   fmt.Sprintf("Export shipment %s declares %.2f %s but linked lots total %.2f — mass-balance discrepancy of %.1f%%", record.ExporterLotNumber, record.Weight, record.Unit, lotTotal, *record.WeightDiscrepancyPct),
					UserID:    userID,
				}
				_, _ = h.TraceabilityRepo.InsertTrackingAlert(alert)
			}
		}
	}

	// If the shipment is already marked shipped, close the chain for its lots.
	if record.ShippedAt != nil || isShippedStatus(record.ApplicationStatus) {
		shippedAt := time.Now()
		if record.ShippedAt != nil {
			shippedAt = *record.ShippedAt
		}
		_ = h.ComplianceRepo.MarkShipmentLotsExported(id, userID, shippedAt)
	}

	utils.WriteSuccessResponse(w, "Export shipment created successfully", record)
}

// isShippedStatus reports whether an application status means the minerals left the country.
func isShippedStatus(status string) bool {
	switch strings.ToLower(status) {
	case "shipped", "exported", "completed":
		return true
	}
	return false
}

func (h *ComplianceHandler) UpdateExportShipment(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	var req ExportShipmentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ExporterStatus == "" {
		req.ExporterStatus = string(data.StatusBlue)
	}
	if !validStatus(w, req.ExporterStatus) {
		return
	}
	if err := validateExportShipment(req); err != nil {
		utils.WriteValidationError(w, err.Error())
		return
	}
	record, err := h.ComplianceRepo.GetExportShipment(id, userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Export shipment not found")
		return
	}
	wasShipped := record.ShippedAt != nil || isShippedStatus(record.ApplicationStatus)
	applyExportShipmentRequest(record, req)
	if err := h.ComplianceRepo.UpdateExportShipment(record); err != nil {
		utils.WriteInternalServerError(w, "Failed to update export shipment")
		return
	}
	// When the shipment transitions to shipped, close the chain of custody for
	// every linked lot automatically.
	nowShipped := record.ShippedAt != nil || isShippedStatus(record.ApplicationStatus)
	if nowShipped && !wasShipped {
		shippedAt := time.Now()
		if record.ShippedAt != nil {
			shippedAt = *record.ShippedAt
		}
		_ = h.ComplianceRepo.MarkShipmentLotsExported(record.ID, userID, shippedAt)
	}
	utils.WriteSuccessResponse(w, "Export shipment updated successfully", record)
}

// generateVerifyCode returns a URL-safe random code for public lot verification.
func generateVerifyCode() string {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("v%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// generateLotNumber mints a lot number for clients that derive everything else
// about a lot (weight, grade, site status) and were never given a field to
// type one in — the workflow-first registration flow, not the older manual
// form. Same shape MineTrace uses: LOT-<year>-<3 random uppercase chars>.
// LotNumber has no unique constraint in this schema, so a collision is
// cosmetic, not a write failure — the random suffix just keeps it unlikely.
func generateLotNumber() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I — easy to misread on a printed tag
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("LOT-%d-%d", time.Now().Year(), time.Now().UnixNano()%1000)
	}
	suffix := make([]byte, 3)
	for i, v := range b {
		suffix[i] = alphabet[int(v)%len(alphabet)]
	}
	return fmt.Sprintf("LOT-%d-%s", time.Now().Year(), string(suffix))
}

// GetPublicVerification returns a sanitized, login-free view of a lot's chain of
// custody for anyone who scans its QR code (reg. 53 — access to mine site/status).
// No user IDs, names, prices or account data are exposed.
func (h *ComplianceHandler) GetPublicVerification(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		utils.WriteValidationError(w, "Verification code is required")
		return
	}
	view, err := h.ComplianceRepo.GetPublicLotView(code)
	if err != nil {
		utils.WriteNotFoundError(w, "No lot matches this verification code")
		return
	}
	utils.WriteSuccessResponse(w, "Lot verified", view)
}

// HandoverCoCLot transfers custody of a lot to another user by email, recording
// an immutable custody transfer. Enables the multi-party chain of custody where
// minerals move between separate accounts (operator → transporter → exporter).
func (h *ComplianceHandler) HandoverCoCLot(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	var req struct {
		ToEmail  string `json:"to_email"`
		Note     string `json:"note"`
		Location string `json:"location"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.ToEmail) == "" {
		utils.WriteValidationError(w, "The recipient's registered email is required")
		return
	}
	result, err := h.ComplianceRepo.HandoverLot(id, userID, req.ToEmail, req.Note, req.Location)
	if err != nil {
		utils.WriteValidationError(w, err.Error())
		return
	}
	utils.WriteSuccessResponse(w, "Custody handed over to "+result.ToName, result)
}

// GetLotPassport returns the complete production→transport→export trace for one CoC lot.
func (h *ComplianceHandler) GetLotPassport(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	passport, err := h.ComplianceRepo.GetLotPassport(id, userID)
	if err != nil {
		utils.WriteNotFoundError(w, "CoC lot not found")
		return
	}
	utils.WriteSuccessResponse(w, "Lot passport retrieved successfully", passport)
}

func (h *ComplianceHandler) DeleteExportShipment(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	err := h.ComplianceRepo.DeleteExportShipment(id, userID)
	writeDeleted(w, err, "Export shipment deleted successfully")
}

func (h *ComplianceHandler) GetDueDiligenceReports(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	records, err := h.ComplianceRepo.GetDueDiligenceReports(userID)
	writeList(w, records, err, "Due diligence reports retrieved successfully")
}

func (h *ComplianceHandler) CreateDueDiligenceReport(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	var req DueDiligenceReportRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	record, ok := dueDiligenceReportFromRequest(w, req, userID)
	if !ok {
		return
	}
	id, err := h.ComplianceRepo.InsertDueDiligenceReport(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create due diligence report")
		return
	}
	record.ID = id
	utils.WriteSuccessResponse(w, "Due diligence report created successfully", record)
}

func (h *ComplianceHandler) UpdateDueDiligenceReport(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	var req DueDiligenceReportRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	record, err := h.ComplianceRepo.GetDueDiligenceReport(id, userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Due diligence report not found")
		return
	}
	if !applyDueDiligenceReportRequest(w, record, req) {
		return
	}
	if err := h.ComplianceRepo.UpdateDueDiligenceReport(record); err != nil {
		utils.WriteInternalServerError(w, "Failed to update due diligence report")
		return
	}
	utils.WriteSuccessResponse(w, "Due diligence report updated successfully", record)
}

func (h *ComplianceHandler) DeleteDueDiligenceReport(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	err := h.ComplianceRepo.DeleteDueDiligenceReport(id, userID)
	writeDeleted(w, err, "Due diligence report deleted successfully")
}

func (h *ComplianceHandler) GetThirdPartyAudits(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	records, err := h.ComplianceRepo.GetThirdPartyAudits(userID)
	writeList(w, records, err, "Third party audits retrieved successfully")
}

func (h *ComplianceHandler) CreateThirdPartyAudit(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	var req ThirdPartyAuditRequest
	if !decodeJSON(w, r, &req) || !validStatus(w, req.Status) {
		return
	}
	if req.ExporterName == "" {
		utils.WriteValidationError(w, "Exporter name is required")
		return
	}
	record := thirdPartyAuditFromRequest(req, userID)
	id, err := h.ComplianceRepo.InsertThirdPartyAudit(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create third party audit")
		return
	}
	record.ID = id
	utils.WriteSuccessResponse(w, "Third party audit created successfully", record)
}

func (h *ComplianceHandler) UpdateThirdPartyAudit(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	var req ThirdPartyAuditRequest
	if !decodeJSON(w, r, &req) || !validStatus(w, req.Status) {
		return
	}
	record, err := h.ComplianceRepo.GetThirdPartyAudit(id, userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Third party audit not found")
		return
	}
	applyThirdPartyAuditRequest(record, req)
	if err := h.ComplianceRepo.UpdateThirdPartyAudit(record); err != nil {
		utils.WriteInternalServerError(w, "Failed to update third party audit")
		return
	}
	utils.WriteSuccessResponse(w, "Third party audit updated successfully", record)
}

func (h *ComplianceHandler) DeleteThirdPartyAudit(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	err := h.ComplianceRepo.DeleteThirdPartyAudit(id, userID)
	writeDeleted(w, err, "Third party audit deleted successfully")
}

func (h *ComplianceHandler) GetComplianceDocuments(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	records, err := h.ComplianceRepo.GetComplianceDocuments(userID)
	writeList(w, records, err, "Compliance documents retrieved successfully")
}

func (h *ComplianceHandler) CreateComplianceDocument(w http.ResponseWriter, r *http.Request) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return
	}
	var req ComplianceDocumentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.DocumentType == "" || req.Title == "" || req.Reference == "" {
		utils.WriteValidationError(w, "Document type, title and reference are required")
		return
	}
	record := complianceDocumentFromRequest(req, userID)
	id, err := h.ComplianceRepo.InsertComplianceDocument(record)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to create compliance document")
		return
	}
	record.ID = id
	utils.WriteSuccessResponse(w, "Compliance document created successfully", record)
}

func (h *ComplianceHandler) UpdateComplianceDocument(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	var req ComplianceDocumentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	record, err := h.ComplianceRepo.GetComplianceDocument(id, userID)
	if err != nil {
		utils.WriteNotFoundError(w, "Compliance document not found")
		return
	}
	applyComplianceDocumentRequest(record, req)
	if err := h.ComplianceRepo.UpdateComplianceDocument(record); err != nil {
		utils.WriteInternalServerError(w, "Failed to update compliance document")
		return
	}
	utils.WriteSuccessResponse(w, "Compliance document updated successfully", record)
}

func (h *ComplianceHandler) DeleteComplianceDocument(w http.ResponseWriter, r *http.Request) {
	userID, id := userAndID(w, r)
	if userID == 0 || id == 0 {
		return
	}
	err := h.ComplianceRepo.DeleteComplianceDocument(id, userID)
	writeDeleted(w, err, "Compliance document deleted successfully")
}

func authenticatedUserID(w http.ResponseWriter, r *http.Request) uint {
	userID := middleware.GetUserIDFromRequest(r)
	if userID == 0 {
		utils.WriteUnauthorizedError(w, "User not authenticated")
	}
	return userID
}

func userAndID(w http.ResponseWriter, r *http.Request) (uint, uint) {
	userID := authenticatedUserID(w, r)
	if userID == 0 {
		return 0, 0
	}
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid record ID")
		return 0, 0
	}
	return userID, uint(id)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return false
	}
	return true
}

func writeList(w http.ResponseWriter, records interface{}, err error, message string) {
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve compliance records")
		return
	}
	utils.WriteSuccessResponse(w, message, records)
}

func writeDeleted(w http.ResponseWriter, err error, message string) {
	if err != nil {
		utils.WriteNotFoundError(w, "Compliance record not found")
		return
	}
	utils.WriteSuccessResponse(w, message, nil)
}

func validStatus(w http.ResponseWriter, status string) bool {
	switch data.ComplianceStatus(status) {
	case data.StatusGreen, data.StatusYellow, data.StatusRed, data.StatusBlue:
		return true
	default:
		utils.WriteValidationError(w, "Status must be one of green, yellow, red or blue")
		return false
	}
}

func parseOptionalDate(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	if parsed, err := time.Parse("2006-01-02", *value); err == nil {
		return &parsed
	}
	if parsed, err := time.Parse(time.RFC3339, *value); err == nil {
		return &parsed
	}
	return nil
}

func parseRequiredDate(w http.ResponseWriter, value string, field string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		utils.WriteValidationError(w, field+" must use YYYY-MM-DD format")
		return time.Time{}, false
	}
	return parsed, true
}

func mineSiteCertificationFromRequest(req MineSiteCertificationRequest, userID uint) *data.MineSiteCertification {
	record := &data.MineSiteCertification{UserID: userID}
	applyMineSiteCertificationRequest(record, req)
	return record
}

func applyMineSiteCertificationRequest(record *data.MineSiteCertification, req MineSiteCertificationRequest) {
	record.MineSiteInfoID = req.MineSiteInfoID
	record.MineSiteName = req.MineSiteName
	record.RMDIdentification = req.RMDIdentification
	record.InspectionDate = parseOptionalDate(req.InspectionDate)
	record.InspectorName = req.InspectorName
	record.Status = data.ComplianceStatus(req.Status)
	record.ReportReference = req.ReportReference
	record.Findings = req.Findings
	record.CorrectiveActions = req.CorrectiveActions
	record.GracePeriodEndsAt = parseOptionalDate(req.GracePeriodEndsAt)
	record.FollowUpRequestedAt = parseOptionalDate(req.FollowUpRequestedAt)
	record.FollowUpDueAt = parseOptionalDate(req.FollowUpDueAt)
	record.IsSuspended = req.IsSuspended || record.Status == data.StatusRed
	record.SuspensionLiftedAt = parseOptionalDate(req.SuspensionLiftedAt)
}

func validateCoCLot(req CoCLotRequest, hasProductionRecords bool) error {
	if req.LotNumber == "" {
		return validationErr("Lot number is required")
	}
	if req.MineralType == "" {
		return validationErr("Mineral type is required")
	}
	// Weight may be derived automatically from linked production records
	// or from the mixed-lot composition.
	if req.Weight <= 0 && !hasProductionRecords && len(req.SourceLots) == 0 {
		return validationErr("Weight must be greater than zero, or link production records / source lots so it can be derived")
	}
	if req.SourceMineSite == "" || req.CoCSystem == "" {
		return validationErr("Source mine site and CoC system are required")
	}
	if req.Unit == "" && !hasProductionRecords {
		return validationErr("Unit is required")
	}
	if data.ComplianceStatus(req.MineSiteStatus) == data.StatusRed {
		return validationErr("Designated minerals from red-status mine sites cannot be traded or exported")
	}
	return nil
}

func coCLotFromRequest(req CoCLotRequest, userID uint) *data.CoCLot {
	record := &data.CoCLot{UserID: userID}
	applyCoCLotRequest(record, req)
	return record
}

func applyCoCLotRequest(record *data.CoCLot, req CoCLotRequest) {
	record.LotNumber = req.LotNumber
	record.ParentLotNumbers = req.ParentLotNumbers
	record.MineSiteCertificationID = req.MineSiteCertificationID
	record.PurchaseOrderNumber = req.PurchaseOrderNumber
	record.PurchaseDate = parseOptionalDate(req.PurchaseDate)
	record.AcceptedBy = req.AcceptedBy
	record.MineralType = data.MineralType(req.MineralType)
	record.OreType = req.OreType
	record.Weight = req.Weight
	record.Unit = req.Unit
	record.Grade = req.Grade
	record.GradeValue = req.GradeValue
	record.GradeUnit = req.GradeUnit
	record.NumberOfSacks = req.NumberOfSacks
	record.SourceMineSite = req.SourceMineSite
	record.MineSiteStatus = data.ComplianceStatus(req.MineSiteStatus)
	record.MineOperatorName = req.MineOperatorName
	record.MinerName = req.MinerName
	record.MinerNationalID = req.MinerNationalID
	record.ArtisanalLicenseNumber = req.ArtisanalLicenseNumber
	record.CoCSystem = req.CoCSystem
	record.SealNumber = req.SealNumber
	record.RegisteredAt = parseOptionalDate(req.RegisteredAt)
	record.SealedAt = parseOptionalDate(req.SealedAt)
	record.ShippedAt = parseOptionalDate(req.ShippedAt)
	record.TransporterName = req.TransporterName
	record.TransportRoute = req.TransportRoute
	record.CurrentCustodian = req.CurrentCustodian
	record.UpstreamActors = req.UpstreamActors
	record.TaxesFeesRoyalties = req.TaxesFeesRoyalties
	record.VerificationOfficer = req.VerificationOfficer
	record.DocumentationReference = req.DocumentationReference
	record.IsExported = req.IsExported
}

func validateExportShipment(req ExportShipmentRequest) error {
	if req.ExporterLotNumber == "" || req.ExporterName == "" || req.CustomerName == "" || req.DestinationCountry == "" {
		return validationErr("Exporter lot number, exporter name, customer name and destination country are required")
	}
	if req.MaterialDescription == "" || req.IncomingLotNumbers == "" {
		return validationErr("Material description and incoming lot numbers are required")
	}
	if req.Weight <= 0 || req.Unit == "" {
		return validationErr("Weight and unit are required")
	}
	if data.ComplianceStatus(req.ExporterStatus) == data.StatusRed {
		return validationErr("Red-status exporters are not eligible to certify export shipments")
	}
	return nil
}

func exportShipmentFromRequest(req ExportShipmentRequest, userID uint) *data.ExportShipment {
	record := &data.ExportShipment{UserID: userID}
	applyExportShipmentRequest(record, req)
	return record
}

func applyExportShipmentRequest(record *data.ExportShipment, req ExportShipmentRequest) {
	status := req.ApplicationStatus
	if status == "" {
		status = "draft"
	}
	record.ExporterLotNumber = req.ExporterLotNumber
	record.ExporterName = req.ExporterName
	record.ExporterLicenseNumber = req.ExporterLicenseNumber
	record.ExporterStatus = data.ComplianceStatus(req.ExporterStatus)
	record.CustomerName = req.CustomerName
	record.CustomerAddress = req.CustomerAddress
	record.DestinationCountry = req.DestinationCountry
	record.MaterialDescription = req.MaterialDescription
	record.Weight = req.Weight
	record.Unit = req.Unit
	record.Grade = req.Grade
	record.IncomingLotNumbers = req.IncomingLotNumbers
	record.IncomingLotWeights = req.IncomingLotWeights
	record.TaxesFeesRoyalties = req.TaxesFeesRoyalties
	record.SealedAt = parseOptionalDate(req.SealedAt)
	record.ShippedAt = parseOptionalDate(req.ShippedAt)
	record.TransporterName = req.TransporterName
	record.TransportRoute = req.TransportRoute
	record.AuthorisedOfficer = req.AuthorisedOfficer
	record.ApplicationStatus = status
	record.ICGLRCertificateNumber = req.ICGLRCertificateNumber
	record.ICGLRCertificateFile = req.ICGLRCertificateFile
	record.CertificateIssuedAt = parseOptionalDate(req.CertificateIssuedAt)
	record.CertificateExpiresAt = parseOptionalDate(req.CertificateExpiresAt)
}

func dueDiligenceReportFromRequest(w http.ResponseWriter, req DueDiligenceReportRequest, userID uint) (*data.DueDiligenceReport, bool) {
	record := &data.DueDiligenceReport{UserID: userID}
	if !applyDueDiligenceReportRequest(w, record, req) {
		return nil, false
	}
	return record, true
}

func applyDueDiligenceReportRequest(w http.ResponseWriter, record *data.DueDiligenceReport, req DueDiligenceReportRequest) bool {
	start, ok := parseRequiredDate(w, req.ReportingPeriodStart, "Reporting period start")
	if !ok {
		return false
	}
	end, ok := parseRequiredDate(w, req.ReportingPeriodEnd, "Reporting period end")
	if !ok {
		return false
	}
	if req.ResponsiblePerson == "" {
		utils.WriteValidationError(w, "Responsible person is required")
		return false
	}
	frequency := data.ReportFrequency(req.Frequency)
	if frequency == "" {
		frequency = data.FrequencyAnnual
	}
	record.ReportingPeriodStart = start
	record.ReportingPeriodEnd = end
	record.Frequency = frequency
	record.ResponsiblePerson = req.ResponsiblePerson
	record.MineralChainPolicy = req.MineralChainPolicy
	record.ManagementSystem = req.ManagementSystem
	record.RiskAssessmentSummary = req.RiskAssessmentSummary
	record.RiskMitigationPlan = req.RiskMitigationPlan
	record.GrievanceMechanism = req.GrievanceMechanism
	record.SupplierCapacity = req.SupplierCapacity
	record.GovernmentPayments = req.GovernmentPayments
	record.BeneficialOwnership = req.BeneficialOwnership
	record.PublishedAt = parseOptionalDate(req.PublishedAt)
	record.SubmittedToDirectorate = req.SubmittedToDirectorate
	record.SubmittedAt = parseOptionalDate(req.SubmittedAt)
	record.AttachmentReference = req.AttachmentReference
	return true
}

func thirdPartyAuditFromRequest(req ThirdPartyAuditRequest, userID uint) *data.ThirdPartyAudit {
	record := &data.ThirdPartyAudit{UserID: userID}
	applyThirdPartyAuditRequest(record, req)
	return record
}

func applyThirdPartyAuditRequest(record *data.ThirdPartyAudit, req ThirdPartyAuditRequest) {
	record.ExporterName = req.ExporterName
	record.AuditorName = req.AuditorName
	record.AuditRequestedAt = parseOptionalDate(req.AuditRequestedAt)
	record.AuditStartedAt = parseOptionalDate(req.AuditStartedAt)
	record.AuditCompletedAt = parseOptionalDate(req.AuditCompletedAt)
	record.Status = data.ComplianceStatus(req.Status)
	record.StatusExpiresAt = parseOptionalDate(req.StatusExpiresAt)
	record.ReportReference = req.ReportReference
	record.Findings = req.Findings
	record.CorrectiveActions = req.CorrectiveActions
	record.FollowUpRequestedAt = parseOptionalDate(req.FollowUpRequestedAt)
	record.FollowUpDueAt = parseOptionalDate(req.FollowUpDueAt)
}

func complianceDocumentFromRequest(req ComplianceDocumentRequest, userID uint) *data.ComplianceDocument {
	record := &data.ComplianceDocument{UserID: userID}
	applyComplianceDocumentRequest(record, req)
	return record
}

func applyComplianceDocumentRequest(record *data.ComplianceDocument, req ComplianceDocumentRequest) {
	record.DocumentType = req.DocumentType
	record.Title = req.Title
	record.Reference = req.Reference
	record.RelatedEntity = req.RelatedEntity
	record.RelatedID = req.RelatedID
	record.IssuedAt = parseOptionalDate(req.IssuedAt)
	record.ExpiresAt = parseOptionalDate(req.ExpiresAt)
	record.Notes = req.Notes
}

type validationErr string

func (e validationErr) Error() string {
	return string(e)
}

// GetAvailableProductionRecords returns production records available for linking to chain of custody lots
func (h *ComplianceHandler) GetAvailableProductionRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)

	mineralType := r.URL.Query().Get("mineral_type")
	if mineralType == "" {
		utils.WriteValidationError(w, "Mineral type is required")
		return
	}

	records, err := h.ComplianceRepo.GetAvailableProductionRecords(userID, data.MineralType(mineralType))
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve available production records")
		return
	}

	utils.WriteSuccessResponse(w, "Available production records retrieved successfully", records)
}

// GetProductionRecordsByPit returns production records from a specific pit
func (h *ComplianceHandler) GetProductionRecordsByPit(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)

	pitNumber := r.URL.Query().Get("pit_number")
	if pitNumber == "" {
		utils.WriteValidationError(w, "Pit number is required")
		return
	}

	records, err := h.ComplianceRepo.GetProductionRecordsByPit(userID, pitNumber)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve production records by pit")
		return
	}

	utils.WriteSuccessResponse(w, "Production records by pit retrieved successfully", records)
}

// LinkProductionRecordsRequest represents the request to link production records to a CoC lot
type LinkProductionRecordsRequest struct {
	ProductionRecordIDs []uint `json:"production_record_ids"`
}

// LinkProductionRecords links existing production records to a chain of custody lot
func (h *ComplianceHandler) LinkProductionRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromRequest(r)

	lotID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		utils.WriteValidationError(w, "Invalid lot ID")
		return
	}

	// Verify the lot belongs to the user
	_, err = h.ComplianceRepo.GetCoCLot(uint(lotID), userID)
	if err != nil {
		if err.Error() == "record not found" {
			utils.WriteNotFoundError(w, "Chain of custody lot not found")
		} else {
			utils.WriteInternalServerError(w, "Failed to retrieve chain of custody lot")
		}
		return
	}

	var req LinkProductionRecordsRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		utils.WriteValidationError(w, "Invalid request body")
		return
	}

	if len(req.ProductionRecordIDs) == 0 {
		utils.WriteValidationError(w, "At least one production record ID is required")
		return
	}

	// Link the production records
	err = h.ComplianceRepo.LinkProductionRecordsToCoCLot(uint(lotID), req.ProductionRecordIDs)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to link production records to chain of custody lot")
		return
	}

	// Return the updated lot with production records
	updatedLot, err := h.ComplianceRepo.GetCoCLot(uint(lotID), userID)
	if err != nil {
		utils.WriteInternalServerError(w, "Failed to retrieve updated chain of custody lot")
		return
	}

	utils.WriteSuccessResponse(w, "Production records linked to chain of custody lot successfully", updatedLot)
}