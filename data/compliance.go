package data

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ComplianceRepository implements ComplianceInterface using GORM.
type ComplianceRepository struct {
	db *gorm.DB
}

// UnitToKg mirrors lib/stock.ts's UNIT_TO_KG on the frontend — keep the two in
// sync. Without this, summing production records recorded in different units
// (2 g + 1 kg) added the raw numbers together (= 3) instead of converting
// first (= 1.002 kg), and copied whichever record's unit happened to be first.
var UnitToKg = map[string]float64{
	"kg":     1,
	"g":      0.001,
	"grams":  0.001,
	"tonnes": 1000,
	"t":      1000,
	// Troy ounce — the precious-metals standard, not the avoirdupois ounce.
	"oz": 0.0311034768,
}

// sumQuantityKg converts every item's quantity to kilograms and adds it up.
// Items in a unit the table doesn't recognise are skipped — reported by
// unconvertibleCount — rather than added in at the wrong scale.
func sumQuantityKg(items []InventoryItem) (totalKg float64, unconvertibleCount int) {
	for _, item := range items {
		factor, ok := UnitToKg[strings.ToLower(item.Unit)]
		if !ok {
			unconvertibleCount++
			continue
		}
		totalKg += item.Quantity * factor
	}
	return totalKg, unconvertibleCount
}

// NewComplianceRepository creates a compliance repository.
func NewComplianceRepository(db *gorm.DB) ComplianceInterface {
	return &ComplianceRepository{db: db}
}

func (r *ComplianceRepository) GetMineSiteCertifications(userID uint) ([]*MineSiteCertification, error) {
	var records []*MineSiteCertification
	result := r.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) GetMineSiteCertification(id uint, userID uint) (*MineSiteCertification, error) {
	var record MineSiteCertification
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *ComplianceRepository) InsertMineSiteCertification(record *MineSiteCertification) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *ComplianceRepository) UpdateMineSiteCertification(record *MineSiteCertification) error {
	return r.db.Save(record).Error
}

func (r *ComplianceRepository) DeleteMineSiteCertification(id uint, userID uint) error {
	return deleteForUser[MineSiteCertification](r.db, id, userID)
}

func (r *ComplianceRepository) GetCoCLots(userID uint) ([]*CoCLot, error) {
	var records []*CoCLot
	// A user sees lots they own, plus any lot currently handed over to them.
	// This is additive — it only ever returns more rows, never fewer.
	result := r.db.Preload("ProductionRecords").Preload("MineSiteCertification").
		Preload("SourceLots").Preload("SourceLots.SourceLot").
		Where("user_id = ? OR current_custodian_user_id = ?", userID, userID).
		Order("updated_at DESC").Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) GetCoCLot(id uint, userID uint) (*CoCLot, error) {
	var record CoCLot
	result := r.db.Preload("ProductionRecords").Preload("MineSiteCertification").
		Preload("SourceLots").Preload("SourceLots.SourceLot").
		Where("id = ? AND (user_id = ? OR current_custodian_user_id = ?)", id, userID, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// GetCoCLotsByIDs looks lots up by ID, scoped the same way GetCoCLots is: a
// user may reach a lot they own or one currently handed over to them.
func (r *ComplianceRepository) GetCoCLotsByIDs(ids []uint, userID uint) ([]*CoCLot, error) {
	var records []*CoCLot
	result := r.db.Preload("MineSiteCertification").
		Where("id IN ? AND (user_id = ? OR current_custodian_user_id = ?)", ids, userID, userID).Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) GetAvailableProductionRecords(userID uint, mineralType MineralType) ([]*InventoryItem, error) {
	var records []*InventoryItem

	mineralName := strings.ToLower(string(mineralType))

	// Exclude records already linked to a CoC lot so the same production
	// cannot be counted twice in the chain of custody.
	result := r.db.Where("user_id = ? AND type = 'mineral'", userID).
		Where("LOWER(name) LIKE ?", "%"+mineralName+"%").
		Where("id NOT IN (SELECT inventory_item_id FROM coc_lot_production_records WHERE inventory_item_id IS NOT NULL)").
		Order("date DESC, created_at DESC").Find(&records)

	return records, result.Error
}

func (r *ComplianceRepository) GetProductionRecordsByPit(userID uint, pitNumber string) ([]*InventoryItem, error) {
	var records []*InventoryItem
	result := r.db.Where("user_id = ? AND pit_number = ? AND type = 'mineral'", userID, pitNumber).
		Where("id NOT IN (SELECT inventory_item_id FROM coc_lot_production_records WHERE inventory_item_id IS NOT NULL)").
		Order("date DESC").Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) LinkProductionRecordsToCoCLot(lotID uint, productionRecordIDs []uint) error {
	// Use GORM's Association to handle the many2many relationship
	var lot CoCLot
	if err := r.db.First(&lot, lotID).Error; err != nil {
		return err
	}

	// Get the production records to link
	var records []InventoryItem
	if err := r.db.Where("id IN ?", productionRecordIDs).Find(&records).Error; err != nil {
		return err
	}

	// Add the association
	return r.db.Model(&lot).Association("ProductionRecords").Append(records)
}

func (r *ComplianceRepository) InsertCoCLot(record *CoCLot) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *ComplianceRepository) UpdateCoCLot(record *CoCLot) error {
	return r.db.Save(record).Error
}

func (r *ComplianceRepository) DeleteCoCLot(id uint, userID uint) error {
	return deleteForUser[CoCLot](r.db, id, userID)
}

func (r *ComplianceRepository) GetExportShipments(userID uint) ([]*ExportShipment, error) {
	var records []*ExportShipment
	result := r.db.Preload("CoCLots").Where("user_id = ?", userID).Order("updated_at DESC").Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) GetExportShipment(id uint, userID uint) (*ExportShipment, error) {
	var record ExportShipment
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *ComplianceRepository) InsertExportShipment(record *ExportShipment) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *ComplianceRepository) UpdateExportShipment(record *ExportShipment) error {
	return r.db.Save(record).Error
}

func (r *ComplianceRepository) DeleteExportShipment(id uint, userID uint) error {
	return deleteForUser[ExportShipment](r.db, id, userID)
}

func (r *ComplianceRepository) GetDueDiligenceReports(userID uint) ([]*DueDiligenceReport, error) {
	var records []*DueDiligenceReport
	result := r.db.Where("user_id = ?", userID).Order("reporting_period_end DESC").Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) GetDueDiligenceReport(id uint, userID uint) (*DueDiligenceReport, error) {
	var record DueDiligenceReport
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *ComplianceRepository) InsertDueDiligenceReport(record *DueDiligenceReport) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *ComplianceRepository) UpdateDueDiligenceReport(record *DueDiligenceReport) error {
	return r.db.Save(record).Error
}

func (r *ComplianceRepository) DeleteDueDiligenceReport(id uint, userID uint) error {
	return deleteForUser[DueDiligenceReport](r.db, id, userID)
}

func (r *ComplianceRepository) GetThirdPartyAudits(userID uint) ([]*ThirdPartyAudit, error) {
	var records []*ThirdPartyAudit
	result := r.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) GetThirdPartyAudit(id uint, userID uint) (*ThirdPartyAudit, error) {
	var record ThirdPartyAudit
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *ComplianceRepository) InsertThirdPartyAudit(record *ThirdPartyAudit) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *ComplianceRepository) UpdateThirdPartyAudit(record *ThirdPartyAudit) error {
	return r.db.Save(record).Error
}

func (r *ComplianceRepository) DeleteThirdPartyAudit(id uint, userID uint) error {
	return deleteForUser[ThirdPartyAudit](r.db, id, userID)
}

func (r *ComplianceRepository) GetComplianceDocuments(userID uint) ([]*ComplianceDocument, error) {
	var records []*ComplianceDocument
	result := r.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&records)
	return records, result.Error
}

func (r *ComplianceRepository) GetComplianceDocument(id uint, userID uint) (*ComplianceDocument, error) {
	var record ComplianceDocument
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *ComplianceRepository) InsertComplianceDocument(record *ComplianceDocument) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *ComplianceRepository) UpdateComplianceDocument(record *ComplianceDocument) error {
	return r.db.Save(record).Error
}

func (r *ComplianceRepository) DeleteComplianceDocument(id uint, userID uint) error {
	return deleteForUser[ComplianceDocument](r.db, id, userID)
}

// InsertCoCLotWithLinks creates a CoC lot and, in the same transaction, links the
// production records it was drawn from and/or the source lots blended into it.
// This is the single-entry point that removes repeated data entry: production data
// is referenced, never re-typed.
func (r *ComplianceRepository) InsertCoCLotWithLinks(record *CoCLot, productionRecordIDs []uint, composition []LotComposition) (uint, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var items []InventoryItem
		if len(productionRecordIDs) > 0 {
			if err := tx.Where("id IN ? AND user_id = ?", productionRecordIDs, record.UserID).Find(&items).Error; err != nil {
				return err
			}
			if len(items) != len(productionRecordIDs) {
				return fmt.Errorf("one or more production records were not found")
			}
			// Auto-fill weight and weighted-average grade from the linked
			// production records so nothing is re-typed. Records can be logged in
			// different units (a pit tally in grams, a stockpile in kg), so they
			// are converted to a common unit before being summed or weighted —
			// never added or weighted by their raw, mismatched quantities.
			if record.Weight <= 0 {
				// Any record in a unit outside UnitToKg is excluded from the sum
				// rather than added in at the wrong scale — matching how the
				// Inventory reconciliation (lib/stock.ts) treats the same case.
				// In practice this is rare: production entry only offers
				// convertible units (kg, g, tonnes, oz).
				totalKg, _ := sumQuantityKg(items)
				record.Weight = totalKg
				record.Unit = "kg"
			}
			if record.GradeValue == nil {
				weightSumKg, gradeSum := 0.0, 0.0
				var unit *string
				consistent := true
				for _, item := range items {
					if item.GradeValue == nil {
						consistent = false
						break
					}
					if unit == nil {
						unit = item.GradeUnit
					} else if item.GradeUnit == nil || *item.GradeUnit != *unit {
						consistent = false
						break
					}
					factor, ok := UnitToKg[strings.ToLower(item.Unit)]
					if !ok {
						consistent = false
						break
					}
					kg := item.Quantity * factor
					weightSumKg += kg
					gradeSum += *item.GradeValue * kg
				}
				if consistent && weightSumKg > 0 && unit != nil {
					avg := gradeSum / weightSumKg
					record.GradeValue = &avg
					record.GradeUnit = unit
				}
			}
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Model(record).Association("ProductionRecords").Append(items); err != nil {
				return err
			}
		}
		if len(composition) > 0 {
			for i := range composition {
				composition[i].MixedLotID = record.ID
				composition[i].UserID = record.UserID
			}
			if err := tx.Create(&composition).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return record.ID, err
}

// ProcessingRunResult is returned after a processing run — the freshly created
// output lot and the processing record linking it to its inputs.
type ProcessingRunResult struct {
	OutputLot        *CoCLot           `json:"output_lot"`
	ProcessingRecord *ProcessingRecord `json:"processing_record"`
}

// CreateProcessingRun consumes inputLotIDs into a new output lot, in one
// transaction. This is the real "merge lots" feature: ProcessingRecord has
// carried InputLots and OutputLotID since the schema was written, but no
// frontend form ever populated them — every processing record created
// through the UI was a free-typed note disconnected from the lots it
// claimed to consume.
//
// outputLot must already have its identity and origin fields set by the
// caller (LotNumber, VerifyCode, MineralType, Weight, Unit, SourceMineSite,
// MineSiteStatus, MineSiteCertificationID, CoCSystem, UserID) — mirroring how
// CreateCoCLot's handler prepares a record before InsertCoCLotWithLinks. This
// method owns only the linking and the input lots' lifecycle.
//
// Input lots are re-fetched inside the transaction (rather than trusting
// whatever the caller validated against a moment earlier) so a lot exported
// by a concurrent request can't slip through.
func (r *ComplianceRepository) CreateProcessingRun(
	userID uint,
	inputLotIDs []uint,
	outputLot *CoCLot,
	rec *ProcessingRecord,
) (*ProcessingRunResult, error) {
	result := &ProcessingRunResult{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var inputLots []CoCLot
		if err := tx.Where("id IN ? AND (user_id = ? OR current_custodian_user_id = ?)", inputLotIDs, userID, userID).
			Find(&inputLots).Error; err != nil {
			return err
		}
		if len(inputLots) != len(inputLotIDs) {
			return fmt.Errorf("one or more input lots were not found, or are not in your custody")
		}
		for _, l := range inputLots {
			if l.IsExported {
				return fmt.Errorf("lot %s has already been exported and cannot be processed", l.LotNumber)
			}
		}

		if err := tx.Create(outputLot).Error; err != nil {
			return err
		}

		rec.OutputLotID = &outputLot.ID
		rec.UserID = userID
		if err := tx.Create(rec).Error; err != nil {
			return err
		}
		if err := tx.Model(rec).Association("InputLots").Append(&inputLots); err != nil {
			return err
		}

		// Consumed — no longer independently at the mine or in transit.
		ids := make([]uint, len(inputLots))
		for i, l := range inputLots {
			ids[i] = l.ID
		}
		if err := tx.Model(&CoCLot{}).Where("id IN ?", ids).
			Update("tracking_state", string(StatusProcessing)).Error; err != nil {
			return err
		}

		result.OutputLot = outputLot
		result.ProcessingRecord = rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AttachLotsToShipment links CoC lots to an export shipment and moves them to
// ready_for_export. The lots' data (weights, grades, lot numbers) is aggregated
// by the handler, so nothing is re-typed.
func (r *ComplianceRepository) AttachLotsToShipment(shipmentID uint, lotIDs []uint, userID uint) error {
	return r.db.Model(&CoCLot{}).
		Where("id IN ? AND user_id = ?", lotIDs, userID).
		Updates(map[string]interface{}{
			"export_shipment_id": shipmentID,
			"tracking_state":     string(StatusReadyForExport),
		}).Error
}

// DetachLotsFromShipment releases all lots currently linked to a shipment.
func (r *ComplianceRepository) DetachLotsFromShipment(shipmentID uint, userID uint) error {
	return r.db.Model(&CoCLot{}).
		Where("export_shipment_id = ? AND user_id = ?", shipmentID, userID).
		Updates(map[string]interface{}{
			"export_shipment_id": nil,
			"tracking_state":     string(StatusStored),
		}).Error
}

// MarkShipmentLotsExported flags every lot on a shipment as exported once the
// shipment ships, closing the chain of custody for those lots.
func (r *ComplianceRepository) MarkShipmentLotsExported(shipmentID uint, userID uint, shippedAt time.Time) error {
	return r.db.Model(&CoCLot{}).
		Where("export_shipment_id = ? AND user_id = ?", shipmentID, userID).
		Updates(map[string]interface{}{
			"is_exported":    true,
			"tracking_state": string(StatusExported),
			"shipped_at":     shippedAt,
		}).Error
}

// GetLatestCertificationForMineSite returns the most recent certification record
// for the user's mine site profile, used to derive lot status automatically.
func (r *ComplianceRepository) GetLatestCertificationForMineSite(userID uint) (*MineSiteCertification, error) {
	var record MineSiteCertification
	result := r.db.Where("user_id = ?", userID).
		Order("inspection_date DESC NULLS LAST, updated_at DESC").First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// PublicLotView is the sanitized, login-free representation of a lot returned by
// the QR verification endpoint. It deliberately omits user IDs, names, prices and
// any account data — only what the public may see under reg. 53.
type PublicLotView struct {
	LotNumber      string           `json:"lot_number"`
	Mineral        MineralType      `json:"mineral"`
	Weight         float64          `json:"weight"`
	Unit           string           `json:"unit"`
	Grade          string           `json:"grade,omitempty"`
	SourceMineSite string           `json:"source_mine_site"`
	SiteStatus     ComplianceStatus `json:"site_status"`
	SealNumber     string           `json:"seal_number,omitempty"`
	Registered     time.Time        `json:"registered"`
	SealedAt       *time.Time       `json:"sealed_at,omitempty"`
	ShippedAt      *time.Time       `json:"shipped_at,omitempty"`
	IsExported     bool             `json:"is_exported"`
	CleanChain     bool             `json:"clean_chain"`
	Steps          []PublicStep     `json:"steps"`
}

// PublicStep is one visible event in a lot's public journey.
type PublicStep struct {
	Date     time.Time `json:"date"`
	Type     string    `json:"type"`
	From     string    `json:"from,omitempty"`
	To       string    `json:"to,omitempty"`
	Location string    `json:"location,omitempty"`
}

// HandoverResult is returned after a successful custody handover.
type HandoverResult struct {
	Lot      *CoCLot `json:"lot"`
	ToUserID uint    `json:"to_user_id"`
	ToName   string  `json:"to_name"`
}

// HandoverLot transfers custody of a lot to another user (by email), records an
// immutable custody transfer, and moves the lot to in-transit. The caller must
// currently hold the lot (owner or current custodian). Multi-party chain of custody.
func (r *ComplianceRepository) HandoverLot(lotID, fromUserID uint, toEmail, note, location string) (*HandoverResult, error) {
	var lot CoCLot
	if err := r.db.First(&lot, lotID).Error; err != nil {
		return nil, fmt.Errorf("lot not found")
	}
	// Caller must hold the lot: either the owner (and no one else holds it) or the current custodian.
	holds := (lot.CurrentCustodianUserID != nil && *lot.CurrentCustodianUserID == fromUserID) ||
		(lot.CurrentCustodianUserID == nil && lot.UserID == fromUserID)
	if !holds {
		return nil, fmt.Errorf("only the current custodian can hand this lot over")
	}
	if lot.IsExported {
		return nil, fmt.Errorf("this lot has already been exported")
	}

	var recipient User
	if err := r.db.Where("email = ?", strings.ToLower(strings.TrimSpace(toEmail))).First(&recipient).Error; err != nil {
		return nil, fmt.Errorf("no registered account with that email")
	}
	if recipient.ID == fromUserID {
		return nil, fmt.Errorf("cannot hand a lot over to yourself")
	}

	fromName := "owner"
	if lot.CurrentCustodian != nil && *lot.CurrentCustodian != "" {
		fromName = *lot.CurrentCustodian
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		recipientID := recipient.ID
		updates := map[string]interface{}{
			"current_custodian_user_id": recipientID,
			"current_custodian":         recipient.Name,
			"tracking_state":            string(StatusInTransit),
		}
		if err := tx.Model(&CoCLot{}).Where("id = ?", lot.ID).Updates(updates).Error; err != nil {
			return err
		}
		reason := "Custody handover"
		if location != "" {
			reason += " at " + location
		}
		var notePtr *string
		if note != "" {
			notePtr = &note
		}
		return tx.Create(&CustodyTransfer{
			CoCLotID:       lot.ID,
			FromCustodian:  fromName,
			ToCustodian:    recipient.Name,
			TransferDate:   time.Now(),
			TransferReason: reason,
			ConditionNotes: notePtr,
			UserID:         fromUserID,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	var updated CoCLot
	r.db.Preload("MineSiteCertification").First(&updated, lot.ID)
	return &HandoverResult{Lot: &updated, ToUserID: recipient.ID, ToName: recipient.Name}, nil
}

// BackfillVerifyCodes assigns a verification code to any existing lot that lacks
// one, so lots created before this feature become QR-verifiable. Additive and
// idempotent — safe to run on every startup against a live database.
func (r *ComplianceRepository) BackfillVerifyCodes() (int, error) {
	var lots []CoCLot
	if err := r.db.Where("verify_code IS NULL OR verify_code = ''").Find(&lots).Error; err != nil {
		return 0, err
	}
	count := 0
	for i := range lots {
		b := make([]byte, 10)
		if _, err := cryptorand.Read(b); err != nil {
			continue
		}
		code := hex.EncodeToString(b)
		if err := r.db.Model(&CoCLot{}).Where("id = ?", lots[i].ID).
			Update("verify_code", code).Error; err == nil {
			count++
		}
	}
	return count, nil
}

// GetPublicLotView looks a lot up by its verification code and builds the public view.
func (r *ComplianceRepository) GetPublicLotView(code string) (*PublicLotView, error) {
	var lot CoCLot
	if err := r.db.Preload("MineSiteCertification").
		Where("verify_code = ?", code).First(&lot).Error; err != nil {
		return nil, err
	}

	status := lot.MineSiteStatus
	if lot.MineSiteCertification != nil {
		status = lot.MineSiteCertification.Status
	}
	grade := ""
	if lot.GradeValue != nil {
		unit := ""
		if lot.GradeUnit != nil {
			unit = *lot.GradeUnit
		}
		grade = fmt.Sprintf("%g %s", *lot.GradeValue, unit)
	} else if lot.Grade != nil {
		grade = *lot.Grade
	}
	seal := ""
	if lot.SealNumber != nil {
		seal = *lot.SealNumber
	}

	view := &PublicLotView{
		LotNumber:      lot.LotNumber,
		Mineral:        lot.MineralType,
		Weight:         lot.Weight,
		Unit:           lot.Unit,
		Grade:          grade,
		SourceMineSite: lot.SourceMineSite,
		SiteStatus:     status,
		SealNumber:     seal,
		Registered:     lot.CreatedAt,
		SealedAt:       lot.SealedAt,
		ShippedAt:      lot.ShippedAt,
		IsExported:     lot.IsExported,
		CleanChain:     status != StatusRed,
		Steps:          []PublicStep{},
	}

	// Registration step.
	view.Steps = append(view.Steps, PublicStep{
		Date: lot.CreatedAt, Type: "registered", Location: lot.SourceMineSite,
	})
	// Custody transfers, chronological, sanitized.
	var transfers []CustodyTransfer
	r.db.Where(&CustodyTransfer{CoCLotID: lot.ID}).Order("transfer_date ASC").Find(&transfers)
	for _, t := range transfers {
		view.Steps = append(view.Steps, PublicStep{
			Date: t.TransferDate, Type: "custody_transfer",
			From: t.FromCustodian, To: t.ToCustodian,
		})
	}
	if lot.ShippedAt != nil {
		view.Steps = append(view.Steps, PublicStep{Date: *lot.ShippedAt, Type: "exported"})
	}
	return view, nil
}

// GetLotPassport assembles the complete traceability chain for one lot:
// production → certification → composition → custody → transport → processing → export.
func (r *ComplianceRepository) GetLotPassport(lotID uint, userID uint) (*LotPassport, error) {
	var lot CoCLot
	// A user can open the passport of a lot they own, plus any lot currently
	// handed over to them — the same rule GetCoCLots uses for its list. This
	// endpoint had fallen out of sync with that rule: the lot would appear in
	// a recipient's list but 404 the moment they opened it.
	if err := r.db.Preload("ProductionRecords").Preload("MineSiteCertification").
		Preload("SourceLots").Preload("SourceLots.SourceLot").
		Where("id = ? AND (user_id = ? OR current_custodian_user_id = ?)", lotID, userID, userID).
		First(&lot).Error; err != nil {
		return nil, err
	}

	passport := &LotPassport{
		Lot:               &lot,
		Certification:     lot.MineSiteCertification,
		ProductionRecords: lot.ProductionRecords,
		Composition:       lot.SourceLots,
		CustodyTransfers:  []CustodyTransfer{},
		TransportRecords:  []TransportRecord{},
		ProcessingRecords: []ProcessingRecord{},
		UsedInLots:        []LotComposition{},
		Documents:         []ComplianceDocument{},
		Alerts:            []TrackingAlert{},
	}

	// Everything below belongs to the lot, not to whoever happens to be asking
	// for it — a custody transfer is created under the sender's user ID
	// (HandoverLot), so once the top query above has cleared userID to view
	// this lot at all, none of these are re-filtered by user_id. Doing so
	// used to hide a lot's own handover record from the party it was handed
	// to: they'd open the passport and see "still with the owner".

	// Mixed lots this lot was blended into (forward traceability).
	r.db.Where("source_lot_id = ?", lotID).Find(&passport.UsedInLots)

	// Custody transfers in chronological order. Struct-based condition so GORM
	// resolves the CoCLotID column name itself.
	r.db.Preload("TransferLocation").
		Where(&CustodyTransfer{CoCLotID: lotID}).
		Order("transfer_date ASC").Find(&passport.CustodyTransfers)

	// Transport records that carried this lot — linked records plus legacy
	// records that stored lot IDs as JSON text.
	var transports []TransportRecord
	r.db.Preload("Origin").Preload("Destination").Preload("Lots").
		Order("departure_time ASC").Find(&transports)
	for _, t := range transports {
		matched := false
		for _, l := range t.Lots {
			if l.ID == lotID {
				matched = true
				break
			}
		}
		if !matched && t.CoCLotIDs != "" {
			var ids []uint
			if json.Unmarshal([]byte(t.CoCLotIDs), &ids) == nil {
				for _, id := range ids {
					if id == lotID {
						matched = true
						break
					}
				}
			}
		}
		if matched {
			t.Lots = nil // avoid bloating the payload
			passport.TransportRecords = append(passport.TransportRecords, t)
		}
	}

	// Processing runs that consumed or produced this lot.
	var processes []ProcessingRecord
	r.db.Preload("InputLots").Find(&processes)
	for _, p := range processes {
		matched := p.OutputLotID != nil && *p.OutputLotID == lotID
		if !matched {
			for _, l := range p.InputLots {
				if l.ID == lotID {
					matched = true
					break
				}
			}
		}
		if matched {
			// InputLots is kept (unlike TransportRecords' Lots above) — it's
			// what lets a lot's own passport show real provenance for an
			// output lot: "produced from LOT-X, LOT-Y", not just a run count.
			passport.ProcessingRecords = append(passport.ProcessingRecords, p)
		}
	}

	// Export shipment, documents, alerts and live tracking.
	if lot.ExportShipmentID != nil {
		var shipment ExportShipment
		if err := r.db.Where("id = ?", *lot.ExportShipmentID).First(&shipment).Error; err == nil {
			passport.ExportShipment = &shipment
		}
	}
	r.db.Where("related_entity = ? AND related_id = ?", "coc_lot", lotID).Find(&passport.Documents)
	r.db.Where("lot_id = ? AND lot_type = ?", lotID, "coc_lot").Find(&passport.Alerts)
	var tracking RealTimeTracking
	if err := r.db.Preload("CurrentLocation").
		Where("lot_id = ? AND lot_type = ?", lotID, "coc_lot").
		First(&tracking).Error; err == nil {
		passport.Tracking = &tracking
	}

	passport.ChainComplete = (len(passport.ProductionRecords) > 0 || len(passport.Composition) > 0) &&
		passport.Certification != nil

	return passport, nil
}

func deleteForUser[T any](db *gorm.DB, id uint, userID uint) error {
	result := db.Where("id = ? AND user_id = ?", id, userID).Delete(new(T))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
