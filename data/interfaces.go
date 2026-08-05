package data

import "time"

// UserInterface defines the methods that must be implemented by a User repository
type UserInterface interface {
	GetAll() ([]*User, error)
	GetByEmail(email string) (*User, error)
	GetOne(id uint) (*User, error)
	Insert(user *User) (uint, error)
	Update(user *User) error
	Delete(user *User) error
	DeleteByID(id uint) error
	ResetPassword(userID uint, newPassword string) error
	PasswordMatches(user *User, plainText string) (bool, error)
	// OTP Related methods
	GenerateAndSaveOTP(email string) (string, error)
	VerifyOTP(email, otp string) (bool, error)
	ResetPasswordWithOTP(email, otp, newPassword string) error
}

// IncomeInterface defines the methods for income transactions
type IncomeInterface interface {
	GetAll(userID uint) ([]*Income, error)
	GetOne(id uint, userID uint) (*Income, error)
	Insert(income *Income) (uint, error)
	Update(income *Income) error
	Delete(id uint, userID uint) error
	GetByDateRange(userID uint, startDate, endDate string) ([]*Income, error)
	GetFinancialSummary(userID uint) (*FinancialSummary, error)
	GetMonthlyData(userID uint, year int) ([]*MonthlyData, error)
}

// ExpenseInterface defines the methods for expense transactions
type ExpenseInterface interface {
	GetAll(userID uint) ([]*Expense, error)
	GetOne(id uint, userID uint) (*Expense, error)
	Insert(expense *Expense) (uint, error)
	Update(expense *Expense) error
	Delete(id uint, userID uint) error
	GetByDateRange(userID uint, startDate, endDate string) ([]*Expense, error)
	GetCategoryBreakdown(userID uint) ([]*CategoryBreakdown, error)
	GetMonthlyData(userID uint, year int) ([]*MonthlyData, error)
	GetFinancialSummary(userID uint) (*FinancialSummary, error)
}

// InventoryInterface defines the methods for inventory management
type InventoryInterface interface {
	GetAll(userID uint) ([]*InventoryItem, error)
	GetOne(id uint, userID uint) (*InventoryItem, error)
	Insert(item *InventoryItem) (uint, error)
	Update(item *InventoryItem) error
	Delete(id uint, userID uint) error
	GetLowStockItems(userID uint) ([]*InventoryItem, error)
	UpdateQuantity(id uint, userID uint, quantity float64) error
}

// ComplianceInterface defines the methods for ICGLR compliance records.
type ComplianceInterface interface {
	GetMineSiteCertifications(userID uint) ([]*MineSiteCertification, error)
	GetMineSiteCertification(id uint, userID uint) (*MineSiteCertification, error)
	InsertMineSiteCertification(record *MineSiteCertification) (uint, error)
	UpdateMineSiteCertification(record *MineSiteCertification) error
	DeleteMineSiteCertification(id uint, userID uint) error

	GetCoCLots(userID uint) ([]*CoCLot, error)
	GetCoCLot(id uint, userID uint) (*CoCLot, error)
	GetCoCLotsByIDs(ids []uint, userID uint) ([]*CoCLot, error)
	InsertCoCLot(record *CoCLot) (uint, error)
	InsertCoCLotWithLinks(record *CoCLot, productionRecordIDs []uint, composition []LotComposition) (uint, error)
	UpdateCoCLot(record *CoCLot) error
	DeleteCoCLot(id uint, userID uint) error
	GetAvailableProductionRecords(userID uint, mineralType MineralType) ([]*InventoryItem, error)
	GetProductionRecordsByPit(userID uint, pitNumber string) ([]*InventoryItem, error)
	LinkProductionRecordsToCoCLot(lotID uint, productionRecordIDs []uint) error
	GetLatestCertificationForMineSite(userID uint) (*MineSiteCertification, error)
	GetLotPassport(lotID uint, userID uint) (*LotPassport, error)
	GetPublicLotView(code string) (*PublicLotView, error)
	BackfillVerifyCodes() (int, error)
	HandoverLot(lotID, fromUserID uint, toEmail, note, location string) (*HandoverResult, error)
	AttachLotsToShipment(shipmentID uint, lotIDs []uint, userID uint) error
	DetachLotsFromShipment(shipmentID uint, userID uint) error
	MarkShipmentLotsExported(shipmentID uint, userID uint, shippedAt time.Time) error

	GetExportShipments(userID uint) ([]*ExportShipment, error)
	GetExportShipment(id uint, userID uint) (*ExportShipment, error)
	InsertExportShipment(record *ExportShipment) (uint, error)
	UpdateExportShipment(record *ExportShipment) error
	DeleteExportShipment(id uint, userID uint) error

	GetDueDiligenceReports(userID uint) ([]*DueDiligenceReport, error)
	GetDueDiligenceReport(id uint, userID uint) (*DueDiligenceReport, error)
	InsertDueDiligenceReport(record *DueDiligenceReport) (uint, error)
	UpdateDueDiligenceReport(record *DueDiligenceReport) error
	DeleteDueDiligenceReport(id uint, userID uint) error

	GetThirdPartyAudits(userID uint) ([]*ThirdPartyAudit, error)
	GetThirdPartyAudit(id uint, userID uint) (*ThirdPartyAudit, error)
	InsertThirdPartyAudit(record *ThirdPartyAudit) (uint, error)
	UpdateThirdPartyAudit(record *ThirdPartyAudit) error
	DeleteThirdPartyAudit(id uint, userID uint) error

	GetComplianceDocuments(userID uint) ([]*ComplianceDocument, error)
	GetComplianceDocument(id uint, userID uint) (*ComplianceDocument, error)
	InsertComplianceDocument(record *ComplianceDocument) (uint, error)
	UpdateComplianceDocument(record *ComplianceDocument) error
	DeleteComplianceDocument(id uint, userID uint) error
}

// TraceabilityInterface defines methods for enhanced traceability operations.
type TraceabilityInterface interface {
	// Transport records
	GetTransportRecords(userID uint) ([]*TransportRecord, error)
	GetTransportRecord(id uint, userID uint) (*TransportRecord, error)
	InsertTransportRecord(record *TransportRecord) (uint, error)
	InsertTransportRecordWithLots(record *TransportRecord, lotIDs []uint) (uint, error)
	CompleteTransportDelivery(record *TransportRecord) error
	UpdateTransportRecord(record *TransportRecord) error
	DeleteTransportRecord(id uint, userID uint) error
	
	// Processing records
	GetProcessingRecords(userID uint) ([]*ProcessingRecord, error)
	GetProcessingRecord(id uint, userID uint) (*ProcessingRecord, error)
	InsertProcessingRecord(record *ProcessingRecord) (uint, error)
	UpdateProcessingRecord(record *ProcessingRecord) error
	DeleteProcessingRecord(id uint, userID uint) error
	
	// Real-time tracking
	GetRealTimeTracking(userID uint) ([]*RealTimeTracking, error)
	GetRealTimeTrackingByLot(lotID uint, lotType string, userID uint) (*RealTimeTracking, error)
	InsertRealTimeTracking(record *RealTimeTracking) (uint, error)
	UpdateRealTimeTracking(record *RealTimeTracking) error
	
	// Custody transfers
	GetCustodyTransfers(userID uint) ([]*CustodyTransfer, error)
	GetCustodyTransfersByCoCLot(cocLotID uint, userID uint) ([]*CustodyTransfer, error)
	InsertCustodyTransfer(record *CustodyTransfer) (uint, error)
	UpdateCustodyTransfer(record *CustodyTransfer) error
	
	// Tracking alerts
	GetTrackingAlerts(userID uint) ([]*TrackingAlert, error)
	GetUnresolvedAlerts(userID uint) ([]*TrackingAlert, error)
	InsertTrackingAlert(record *TrackingAlert) (uint, error)
	UpdateTrackingAlert(record *TrackingAlert) error
	ResolveTrackingAlert(id uint, userID uint, resolvedBy string) error
	
	// Photo records
	GetPhotoRecords(userID uint) ([]*PhotoRecord, error)
	GetPhotoRecordsByEntity(entityType string, entityID uint, userID uint) ([]*PhotoRecord, error)
	InsertPhotoRecord(record *PhotoRecord) (uint, error)
	DeletePhotoRecord(id uint, userID uint) error
	
	// Stakeholders
	GetStakeholders(userID uint) ([]*Stakeholder, error)
	GetStakeholder(id uint, userID uint) (*Stakeholder, error)
	InsertStakeholder(record *Stakeholder) (uint, error)
	UpdateStakeholder(record *Stakeholder) error
	DeleteStakeholder(id uint, userID uint) error
	
	// GPS locations
	InsertGPSLocation(location *GPSLocation) (uint, error)
	GetGPSLocation(id uint) (*GPSLocation, error)
}

// Models wraps all repository interfaces
type Models struct {
	User         UserInterface
	Income       IncomeInterface
	Expense      ExpenseInterface
	Inventory    InventoryInterface
	MineSite     MineSiteInterface
	Compliance   ComplianceInterface
	Traceability TraceabilityInterface
}
