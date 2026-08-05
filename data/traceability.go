package data

import (
	"time"
	"gorm.io/gorm"
)

// TraceabilityRepository implements TraceabilityInterface using GORM.
type TraceabilityRepository struct {
	db *gorm.DB
}

// NewTraceabilityRepository creates a traceability repository.
func NewTraceabilityRepository(db *gorm.DB) TraceabilityInterface {
	return &TraceabilityRepository{db: db}
}

// Transport Records
func (r *TraceabilityRepository) GetTransportRecords(userID uint) ([]*TransportRecord, error) {
	var records []*TransportRecord
	result := r.db.Preload("Origin").Preload("Destination").Preload("Lots").Preload("User").
		Where("user_id = ?", userID).Order("created_at DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) GetTransportRecord(id uint, userID uint) (*TransportRecord, error) {
	var record TransportRecord
	result := r.db.Preload("Origin").Preload("Destination").Preload("Lots").Preload("User").
		Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *TraceabilityRepository) InsertTransportRecord(record *TransportRecord) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

// InsertTransportRecordWithLots creates a transport record and, in one transaction,
// links the CoC lots it carries, writes a custody transfer for each lot, moves the
// lots to in_transit with the new custodian, and upserts real-time tracking.
// This removes every duplicated entry step between the CoC and transport modules.
func (r *TraceabilityRepository) InsertTransportRecordWithLots(record *TransportRecord, lotIDs []uint) (uint, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		if len(lotIDs) == 0 {
			return nil
		}

		var lots []CoCLot
		if err := tx.Where("id IN ? AND user_id = ?", lotIDs, record.UserID).Find(&lots).Error; err != nil {
			return err
		}
		if len(lots) != len(lotIDs) {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Model(record).Association("Lots").Append(lots); err != nil {
			return err
		}

		// Auto-fill cargo weight from the linked lots so it is never re-typed.
		if record.NetWeight <= 0 {
			total := 0.0
			for _, lot := range lots {
				total += lot.Weight
			}
			record.NetWeight = total
			if record.GrossWeight <= 0 {
				record.GrossWeight = total
			}
			if err := tx.Model(&TransportRecord{}).Where("id = ?", record.ID).
				Updates(map[string]interface{}{
					"net_weight":   record.NetWeight,
					"gross_weight": record.GrossWeight,
				}).Error; err != nil {
				return err
			}
		}

		now := record.DepartureTime
		if now.IsZero() {
			now = time.Now()
		}
		for _, lot := range lots {
			transfer := CustodyTransfer{
				CoCLotID:           lot.ID,
				FromCustodian:      record.HandoverFromName,
				ToCustodian:        record.HandoverToName,
				TransferDate:       now,
				TransferLocationID: &record.OriginID,
				TransferReason:     "Transport departure — " + record.LicensePlate,
				ConditionNotes:     record.HandoverNotes,
				UserID:             record.UserID,
			}
			if err := tx.Create(&transfer).Error; err != nil {
				return err
			}

			updates := map[string]interface{}{
				"current_custodian": record.HandoverToName,
				"transporter_name":  record.DriverName,
				"tracking_state":    string(StatusInTransit),
			}
			if err := tx.Model(&CoCLot{}).Where("id = ?", lot.ID).Updates(updates).Error; err != nil {
				return err
			}

			// Upsert real-time tracking for the lot.
			var tracking RealTimeTracking
			err := tx.Where("lot_id = ? AND lot_type = ? AND user_id = ?", lot.ID, "coc_lot", record.UserID).
				First(&tracking).Error
			if err == gorm.ErrRecordNotFound {
				tracking = RealTimeTracking{
					LotID:            lot.ID,
					LotType:          "coc_lot",
					CurrentCustodian: record.HandoverToName,
					Status:           StatusInTransit,
					LastUpdated:      time.Now(),
					TransportRecordID: &record.ID,
					UserID:           record.UserID,
				}
				if err := tx.Create(&tracking).Error; err != nil {
					return err
				}
			} else if err == nil {
				tracking.CurrentCustodian = record.HandoverToName
				tracking.Status = StatusInTransit
				tracking.LastUpdated = time.Now()
				tracking.TransportRecordID = &record.ID
				if err := tx.Save(&tracking).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
	return record.ID, err
}

// CompleteTransportDelivery marks a transport as delivered and propagates the
// arrival to every carried lot: custody passes to the receiver, lots become
// stored, and a delivery custody transfer is recorded.
func (r *TraceabilityRepository) CompleteTransportDelivery(record *TransportRecord) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if record.ArrivalTime == nil {
			record.ArrivalTime = &now
		}
		duration := int(record.ArrivalTime.Sub(record.DepartureTime).Minutes())
		record.ActualDuration = &duration
		if err := tx.Save(record).Error; err != nil {
			return err
		}

		var lots []CoCLot
		if err := tx.Model(record).Association("Lots").Find(&lots); err != nil {
			return err
		}
		for _, lot := range lots {
			transfer := CustodyTransfer{
				CoCLotID:           lot.ID,
				FromCustodian:      record.DriverName,
				ToCustodian:        record.HandoverToName,
				TransferDate:       *record.ArrivalTime,
				TransferLocationID: &record.DestinationID,
				TransferReason:     "Transport delivery — " + record.LicensePlate,
				UserID:             record.UserID,
			}
			if err := tx.Create(&transfer).Error; err != nil {
				return err
			}
			updates := map[string]interface{}{
				"tracking_state": string(StatusStored),
			}
			if err := tx.Model(&CoCLot{}).Where("id = ?", lot.ID).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Model(&RealTimeTracking{}).
				Where("lot_id = ? AND lot_type = ? AND user_id = ?", lot.ID, "coc_lot", record.UserID).
				Updates(map[string]interface{}{
					"status":              string(StatusStored),
					"current_location_id": record.DestinationID,
					"last_updated":        now,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *TraceabilityRepository) UpdateTransportRecord(record *TransportRecord) error {
	return r.db.Save(record).Error
}

func (r *TraceabilityRepository) DeleteTransportRecord(id uint, userID uint) error {
	return deleteForUser[TransportRecord](r.db, id, userID)
}

// Processing Records
func (r *TraceabilityRepository) GetProcessingRecords(userID uint) ([]*ProcessingRecord, error) {
	var records []*ProcessingRecord
	result := r.db.Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) GetProcessingRecord(id uint, userID uint) (*ProcessingRecord, error) {
	var record ProcessingRecord
	result := r.db.Preload("User").Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *TraceabilityRepository) InsertProcessingRecord(record *ProcessingRecord) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *TraceabilityRepository) UpdateProcessingRecord(record *ProcessingRecord) error {
	return r.db.Save(record).Error
}

func (r *TraceabilityRepository) DeleteProcessingRecord(id uint, userID uint) error {
	return deleteForUser[ProcessingRecord](r.db, id, userID)
}

// Real-time Tracking
func (r *TraceabilityRepository) GetRealTimeTracking(userID uint) ([]*RealTimeTracking, error) {
	var records []*RealTimeTracking
	result := r.db.Preload("CurrentLocation").Preload("TransportRecord").
		Preload("ProcessingRecord").Preload("User").
		Where("user_id = ?", userID).Order("last_updated DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) GetRealTimeTrackingByLot(lotID uint, lotType string, userID uint) (*RealTimeTracking, error) {
	var record RealTimeTracking
	result := r.db.Preload("CurrentLocation").Preload("TransportRecord").
		Preload("ProcessingRecord").Preload("User").
		Where("lot_id = ? AND lot_type = ? AND user_id = ?", lotID, lotType, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *TraceabilityRepository) InsertRealTimeTracking(record *RealTimeTracking) (uint, error) {
	record.LastUpdated = time.Now()
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *TraceabilityRepository) UpdateRealTimeTracking(record *RealTimeTracking) error {
	record.LastUpdated = time.Now()
	return r.db.Save(record).Error
}

// Custody Transfers
func (r *TraceabilityRepository) GetCustodyTransfers(userID uint) ([]*CustodyTransfer, error) {
	var records []*CustodyTransfer
	result := r.db.Preload("CoCLot").Preload("TransferLocation").Preload("User").
		Where("user_id = ?", userID).Order("transfer_date DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) GetCustodyTransfersByCoCLot(cocLotID uint, userID uint) ([]*CustodyTransfer, error) {
	var records []*CustodyTransfer
	// Struct-based condition so GORM resolves the CoCLotID column name itself.
	result := r.db.Preload("CoCLot").Preload("TransferLocation").Preload("User").
		Where(&CustodyTransfer{CoCLotID: cocLotID}).
		Where("user_id = ?", userID).
		Order("transfer_date DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) InsertCustodyTransfer(record *CustodyTransfer) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *TraceabilityRepository) UpdateCustodyTransfer(record *CustodyTransfer) error {
	return r.db.Save(record).Error
}

// Tracking Alerts
func (r *TraceabilityRepository) GetTrackingAlerts(userID uint) ([]*TrackingAlert, error) {
	var records []*TrackingAlert
	result := r.db.Preload("User").Where("user_id = ?", userID).Order("created_at DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) GetUnresolvedAlerts(userID uint) ([]*TrackingAlert, error) {
	var records []*TrackingAlert
	result := r.db.Preload("User").Where("user_id = ? AND resolved = ?", userID, false).
		Order("created_at DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) InsertTrackingAlert(record *TrackingAlert) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *TraceabilityRepository) UpdateTrackingAlert(record *TrackingAlert) error {
	return r.db.Save(record).Error
}

func (r *TraceabilityRepository) ResolveTrackingAlert(id uint, userID uint, resolvedBy string) error {
	now := time.Now()
	result := r.db.Model(&TrackingAlert{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"resolved":    true,
			"resolved_at": &now,
			"resolved_by": resolvedBy,
		})
	return result.Error
}

// Photo Records
func (r *TraceabilityRepository) GetPhotoRecords(userID uint) ([]*PhotoRecord, error) {
	var records []*PhotoRecord
	result := r.db.Preload("Location").Preload("User").
		Where("user_id = ?", userID).Order("taken_at DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) GetPhotoRecordsByEntity(entityType string, entityID uint, userID uint) ([]*PhotoRecord, error) {
	var records []*PhotoRecord
	result := r.db.Preload("Location").Preload("User").
		Where("related_entity = ? AND related_id = ? AND user_id = ?", entityType, entityID, userID).
		Order("taken_at DESC").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) InsertPhotoRecord(record *PhotoRecord) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *TraceabilityRepository) DeletePhotoRecord(id uint, userID uint) error {
	return deleteForUser[PhotoRecord](r.db, id, userID)
}

// Stakeholders
func (r *TraceabilityRepository) GetStakeholders(userID uint) ([]*Stakeholder, error) {
	var records []*Stakeholder
	result := r.db.Preload("User").Where("user_id = ?", userID).Order("name").Find(&records)
	return records, result.Error
}

func (r *TraceabilityRepository) GetStakeholder(id uint, userID uint) (*Stakeholder, error) {
	var record Stakeholder
	result := r.db.Preload("User").Where("id = ? AND user_id = ?", id, userID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *TraceabilityRepository) InsertStakeholder(record *Stakeholder) (uint, error) {
	result := r.db.Create(record)
	return record.ID, result.Error
}

func (r *TraceabilityRepository) UpdateStakeholder(record *Stakeholder) error {
	return r.db.Save(record).Error
}

func (r *TraceabilityRepository) DeleteStakeholder(id uint, userID uint) error {
	return deleteForUser[Stakeholder](r.db, id, userID)
}

// GPS Locations
func (r *TraceabilityRepository) InsertGPSLocation(location *GPSLocation) (uint, error) {
	result := r.db.Create(location)
	return location.ID, result.Error
}

func (r *TraceabilityRepository) GetGPSLocation(id uint) (*GPSLocation, error) {
	var location GPSLocation
	result := r.db.First(&location, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &location, nil
}