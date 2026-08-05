package data

import (
	"time"

	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleStandard UserRole = "standard"
)

// ChainRole is a user's role in the mineral supply chain. It is separate from
// UserRole (standard/admin) so existing accounts and admin logic are unaffected.
// An empty ChainRole means "legacy" — full access, the behaviour before roles existed.
type ChainRole string

const (
	ChainOperator    ChainRole = "operator"    // registers sites, production, lots
	ChainTransporter ChainRole = "transporter" // receives and moves lots
	ChainExporter    ChainRole = "exporter"    // processes and exports lots
	ChainInspector   ChainRole = "inspector"   // certifies sites, read-only across the chain
)

// ValidChainRole reports whether s is one of the four supply-chain roles.
func ValidChainRole(s string) bool {
	switch ChainRole(s) {
	case ChainOperator, ChainTransporter, ChainExporter, ChainInspector:
		return true
	default:
		return false
	}
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"
)

// PaymentStatus represents the payment status
type PaymentStatus string

const (
	PaymentPaid    PaymentStatus = "paid"
	PaymentUnpaid  PaymentStatus = "unpaid"
	PaymentPartial PaymentStatus = "partial"
)

// ComplianceStatus represents ICGLR status colors for mine sites, exporters and CoC systems.
type ComplianceStatus string

const (
	StatusGreen  ComplianceStatus = "green"
	StatusYellow ComplianceStatus = "yellow"
	StatusRed    ComplianceStatus = "red"
	StatusBlue   ComplianceStatus = "blue"
)

// ReportFrequency represents statutory reporting cadence.
type ReportFrequency string

const (
	FrequencyMonthly   ReportFrequency = "monthly"
	FrequencyQuarterly ReportFrequency = "quarterly"
	FrequencyAnnual    ReportFrequency = "annual"
)

// MineralType represents the type of mineral
type MineralType string

const (
	MineralGold              MineralType = "gold"
	MineralCopper            MineralType = "copper"
	MineralCobalt            MineralType = "cobalt"
	MineralDiamond           MineralType = "diamond"
	MineralIronOre           MineralType = "iron_ore"
	MineralLead              MineralType = "lead"
	MineralZinc              MineralType = "zinc"
	MineralLithium           MineralType = "lithium"
	MineralNickel            MineralType = "nickel"
	MineralColtan            MineralType = "coltan"
	MineralTin               MineralType = "tin"
	MineralWolfram           MineralType = "wolfram"
	MineralTitanium          MineralType = "titanium"
	MineralManganese         MineralType = "manganese"
	MineralRareEarthElements MineralType = "rare_earth_elements"
	MineralUranium           MineralType = "uranium"
	MineralBentonite         MineralType = "bentonite"
	MineralDiatomite         MineralType = "diatomite"
	MineralGraphite          MineralType = "graphite"
	MineralGypsum            MineralType = "gypsum"
	MineralFeldspar          MineralType = "feldspar"
	MineralLimestone         MineralType = "limestone"
	MineralMarble            MineralType = "marble"
	MineralKaolin            MineralType = "kaolin"
	MineralPhosphates        MineralType = "phosphates"
	MineralPozzolana         MineralType = "pozzolana"
	MineralSalt              MineralType = "salt"
	MineralSand              MineralType = "sand"
	MineralVermiculite       MineralType = "vermiculite"
	MineralSilver            MineralType = "silver"
	MineralGranite           MineralType = "granite"
	MineralChromite          MineralType = "chromite"
	MineralGemstones         MineralType = "gemstones"
	MineralOther             MineralType = "other"
)

// GemstoneType represents the type of gemstone
type GemstoneType string

const (
	GemstoneApatite    GemstoneType = "apatite"
	GemstoneBeryl      GemstoneType = "beryl"
	GemstoneAquamarine GemstoneType = "aquamarine"
	GemstoneRuby       GemstoneType = "ruby"
	GemstoneSapphire   GemstoneType = "sapphire"
	GemstoneFlourite   GemstoneType = "flourite"
	GemstoneGarnet     GemstoneType = "garnet"
	GemstoneOpal       GemstoneType = "opal"
	GemstoneQuartz     GemstoneType = "quartz"
	GemstoneTopaz      GemstoneType = "topaz"
	GemstoneTourmaline GemstoneType = "tourmaline"
	GemstoneZircon     GemstoneType = "zircon"
	GemstonePearl      GemstoneType = "pearl"
	GemstoneAmber      GemstoneType = "amber"
	GemstoneKyanite    GemstoneType = "kyanite"
	GemstoneCoral      GemstoneType = "coral"
	GemstoneJade       GemstoneType = "jade"
	GemstoneMalachite  GemstoneType = "malachite"
	GemstoneOnyx       GemstoneType = "onyx"
	GemstonePeridot    GemstoneType = "peridot"
	GemstoneTurquoise  GemstoneType = "turquoise"
	GemstoneDiamond    GemstoneType = "diamond"
	GemstoneAmethyst   GemstoneType = "amethyst"
	GemstoneEmerald    GemstoneType = "emerald"
	GemstoneOther      GemstoneType = "other"
)

// SalesType represents the type of sale
type SalesType string

const (
	SalesTypeMineral      SalesType = "mineral"
	SalesTypeSupply       SalesType = "supply"
	SalesTypeConcentrates SalesType = "concentrates"
	SalesTypeTailings     SalesType = "tailings"
)

// ExpenseCategory represents the category of expense
type ExpenseCategory string

const (
	ExpenseEquipment   ExpenseCategory = "equipment"
	ExpenseLabor       ExpenseCategory = "labor"
	ExpenseChemicals   ExpenseCategory = "chemicals"
	ExpenseMaintenance ExpenseCategory = "maintenance"
	ExpenseTransport   ExpenseCategory = "transport"
	ExpenseOther       ExpenseCategory = "other"
)

// User represents a user in the system
type User struct {
	gorm.Model
	Email     string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Phone     *string        `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Location  *string        `gorm:"type:varchar(255)" json:"location,omitempty"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role      UserRole       `gorm:"type:varchar(50);default:'standard'" json:"role"`
	// ChainRole is the supply-chain role. Nullable/empty for existing users,
	// who keep full access. New signups pick one.
	ChainRole ChainRole      `gorm:"type:varchar(20)" json:"chain_role,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// OTP fields for password reset
	OTPCode      string     `gorm:"type:varchar(6)" json:"-"`
	OTPExpiresAt *time.Time `json:"-"`

	// OTP brute-force protection
	OTPAttempts    int        `gorm:"default:0" json:"-"`
	OTPLockedUntil *time.Time `json:"-"`
}

// Income represents an income transaction (Sales)
type Income struct {
	gorm.Model
	Date            time.Time      `gorm:"not null" json:"date"`
	ItemName        *string        `gorm:"type:varchar(255)" json:"item_name,omitempty"`
	MineralType     MineralType    `gorm:"type:varchar(50);not null;default:'other'" json:"mineral_type"`
	GemstoneType    *GemstoneType  `gorm:"type:varchar(50)" json:"gemstone_type,omitempty"`
	SalesType       SalesType      `gorm:"type:varchar(20);default:'mineral'" json:"sales_type"`
	Quantity        float64        `gorm:"not null" json:"quantity"`
	Unit            string         `gorm:"type:varchar(20);not null" json:"unit"`
	PricePerUnit    float64        `gorm:"not null" json:"price_per_unit"`
	TotalAmount     float64        `gorm:"not null" json:"total_amount"`
	CustomerName    string         `gorm:"type:varchar(100);not null" json:"customer_name"`
	CustomerContact string         `gorm:"type:varchar(100)" json:"customer_contact"`
	PaymentStatus   PaymentStatus  `gorm:"type:varchar(20);default:'unpaid'" json:"payment_status"`
	AmountPaid      float64        `gorm:"default:0" json:"amount_paid"`
	AmountDue       float64        `gorm:"default:0" json:"amount_due"`
	Notes           *string        `gorm:"type:text" json:"notes,omitempty"`
	UserID          uint           `gorm:"not null" json:"user_id"`
	User            User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Expense represents an expense transaction
type Expense struct {
	gorm.Model
	Date            time.Time       `gorm:"not null" json:"date"`
	Category        ExpenseCategory `gorm:"type:varchar(50);not null" json:"category"`
	Description     string          `gorm:"type:varchar(255);not null" json:"description"`
	Quantity        *float64        `gorm:"type:decimal(10,2)" json:"quantity,omitempty"`
	Unit            *string         `gorm:"type:varchar(50)" json:"unit,omitempty"`
	Amount          float64         `gorm:"not null" json:"amount"`
	SupplierName    string          `gorm:"type:varchar(100);not null" json:"supplier_name"`
	SupplierContact *string         `gorm:"type:varchar(100)" json:"supplier_contact,omitempty"`
	PaymentStatus   PaymentStatus   `gorm:"type:varchar(20);default:'unpaid'" json:"payment_status"`
	AmountPaid      float64         `gorm:"default:0" json:"amount_paid"`
	AmountDue       float64         `gorm:"default:0" json:"amount_due"`
	Notes           *string         `gorm:"type:text" json:"notes,omitempty"`
	UserID          uint            `gorm:"not null" json:"user_id"`
	User            User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

// ProductionFrom represents the source of production
type ProductionFrom string

const (
	ProductionFromMine       ProductionFrom = "mine"
	ProductionFromProcessing ProductionFrom = "processing"
)

// ProcessingMethod represents the processing method used
type ProcessingMethod string

const (
	ProcessingCrushing    ProcessingMethod = "crushing"
	ProcessingMilling     ProcessingMethod = "milling"
	ProcessingSieving     ProcessingMethod = "sieving"
	ProcessingGrading     ProcessingMethod = "grading"
	ProcessingSorting     ProcessingMethod = "sorting"
	ProcessingCutting     ProcessingMethod = "cutting"
	ProcessingDressing    ProcessingMethod = "dressing"
	ProcessingLeaching    ProcessingMethod = "leaching"
	ProcessingElution     ProcessingMethod = "elution"
	ProcessingRefining    ProcessingMethod = "refining"
	ProcessingFloatation  ProcessingMethod = "floatation"
	ProcessingGrinding    ProcessingMethod = "grinding"
	ProcessingScreening   ProcessingMethod = "screening"
	ProcessingDrying      ProcessingMethod = "drying"
	ProcessingExfoliation ProcessingMethod = "exfoliation"
	ProcessingPolishing   ProcessingMethod = "polishing"
	ProcessingWashing     ProcessingMethod = "washing"
)

// ProductType represents the type of product
type ProductType string

const (
	ProductOre         ProductType = "ore"
	ProductConcentrate ProductType = "concentrate"
	ProductMetal       ProductType = "metal"
	ProductRough       ProductType = "rough"
	ProductCut         ProductType = "cut"
	ProductPolished    ProductType = "polished"
	ProductFaceted     ProductType = "faceted"
	ProductOther       ProductType = "other"
)

// InventoryItem represents an inventory/production item
type InventoryItem struct {
	gorm.Model
	Name              string            `gorm:"type:varchar(100);not null" json:"name"`
	Type              string            `gorm:"type:varchar(20);not null" json:"type"`  // "mineral" or "supply"
	Date              *time.Time        `json:"date,omitempty"`                         // Production date set by user
	From              *ProductionFrom   `gorm:"type:varchar(20)" json:"from,omitempty"` // "mine" or "processing"
	PitNumber         *string           `gorm:"type:varchar(100)" json:"pit_number,omitempty"`
	MinerName         *string           `gorm:"type:varchar(100)" json:"miner_name,omitempty"`
	MinerSerialNumber *string           `gorm:"type:varchar(100);uniqueIndex" json:"miner_serial_number,omitempty"`
	BatchNumber       *string           `gorm:"type:varchar(100)" json:"batch_number,omitempty"`
	ProcessingMethod  *ProcessingMethod `gorm:"type:varchar(50)" json:"processing_method,omitempty"`
	Product           *ProductType      `gorm:"type:varchar(50)" json:"product,omitempty"`
	GemstoneType      *GemstoneType     `gorm:"type:varchar(50)" json:"gemstone_type,omitempty"`
	GradeValue        *float64          `gorm:"type:decimal(10,4)" json:"grade_value,omitempty"`
	GradeUnit         *string           `gorm:"type:varchar(40)" json:"grade_unit,omitempty"`
	GradeNotes        *string           `gorm:"type:text" json:"grade_notes,omitempty"`
	Quantity          float64           `gorm:"not null" json:"quantity"`
	Unit              string            `gorm:"type:varchar(20);not null" json:"unit"`
	MinStockLevel     float64           `gorm:"not null" json:"min_stock_level"`
	CurrentValue      float64           `gorm:"default:0" json:"current_value"`
	LastUpdated       time.Time         `gorm:"not null" json:"last_updated"`
	UserID            uint              `gorm:"not null" json:"user_id"`
	User              User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `gorm:"index" json:"-"`
}

// FinancialSummary represents financial summary data
type FinancialSummary struct {
	TotalIncome      float64 `json:"total_income"`
	TotalExpenses    float64 `json:"total_expenses"`
	NetProfit        float64 `json:"net_profit"`
	TotalReceivables float64 `json:"total_receivables"`
	TotalPayables    float64 `json:"total_payables"`
	ProfitMargin     float64 `json:"profit_margin"`
}

// MonthlyData represents monthly financial data
type MonthlyData struct {
	Month    string  `json:"month"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Profit   float64 `json:"profit"`
}

// CategoryBreakdown represents category breakdown data
type CategoryBreakdown struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

// MineSiteInfo represents mine site information
type MineSiteInfo struct {
	gorm.Model
	Owner           string         `gorm:"type:varchar(255);not null" json:"owner"`
	License         *string        `gorm:"type:varchar(100)" json:"license,omitempty"`
	Location        string         `gorm:"type:varchar(255);not null" json:"location"`
	Size            *float64       `gorm:"type:decimal(10,2)" json:"size,omitempty"` // hectares
	NumberOfPits    *int           `gorm:"type:integer" json:"number_of_pits,omitempty"`
	Commodities     *string        `gorm:"type:text" json:"commodities,omitempty"`
	Equipment       *string        `gorm:"type:text" json:"equipment,omitempty"`
	Employees       *int           `gorm:"type:integer" json:"employees,omitempty"`
	EstablishedYear *int           `gorm:"type:integer" json:"established_year,omitempty"`
	Contact         *string        `gorm:"type:varchar(255)" json:"contact,omitempty"`
	UserID          uint           `gorm:"not null" json:"user_id"`
	User            User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// MineSiteCertification tracks ICGLR inspection status and follow-up obligations for a mine site.
type MineSiteCertification struct {
	gorm.Model
	MineSiteInfoID      *uint            `gorm:"index" json:"mine_site_info_id,omitempty"`
	MineSiteInfo        *MineSiteInfo    `gorm:"foreignKey:MineSiteInfoID" json:"mine_site_info,omitempty"`
	MineSiteName        string           `gorm:"type:varchar(255);not null" json:"mine_site_name"`
	RMDIdentification   *string          `gorm:"type:varchar(120)" json:"rmd_identification,omitempty"`
	InspectionDate      *time.Time       `json:"inspection_date,omitempty"`
	InspectorName       *string          `gorm:"type:varchar(255)" json:"inspector_name,omitempty"`
	Status              ComplianceStatus `gorm:"type:varchar(20);not null;default:'blue';index" json:"status"`
	ReportReference     *string          `gorm:"type:varchar(255)" json:"report_reference,omitempty"`
	Findings            *string          `gorm:"type:text" json:"findings,omitempty"`
	CorrectiveActions   *string          `gorm:"type:text" json:"corrective_actions,omitempty"`
	GracePeriodEndsAt   *time.Time       `json:"grace_period_ends_at,omitempty"`
	FollowUpRequestedAt *time.Time       `json:"follow_up_requested_at,omitempty"`
	FollowUpDueAt       *time.Time       `json:"follow_up_due_at,omitempty"`
	IsSuspended         bool             `gorm:"default:false" json:"is_suspended"`
	SuspensionLiftedAt  *time.Time       `json:"suspension_lifted_at,omitempty"`
	UserID              uint             `gorm:"not null;index" json:"user_id"`
	User                User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	DeletedAt           gorm.DeletedAt   `gorm:"index" json:"-"`
}

// CoCLot captures chain-of-custody data for a designated mineral lot.
type CoCLot struct {
	gorm.Model
	LotNumber              string           `gorm:"type:varchar(120);not null;index" json:"lot_number"`
	// Public verification code (QR target). Nullable so existing rows migrate
	// cleanly; backfilled on startup. Anyone with the code can view the lot's
	// public chain of custody without logging in (reg. 53).
	VerifyCode             *string          `gorm:"type:varchar(32);uniqueIndex" json:"verify_code,omitempty"`
	ProductionRecordIDs    *string          `gorm:"type:text" json:"production_record_ids,omitempty"`
	ProductionRecords      []InventoryItem  `gorm:"many2many:coc_lot_production_records;" json:"production_records,omitempty"`
	ParentLotNumbers       *string          `gorm:"type:text" json:"parent_lot_numbers,omitempty"`
	// Mixed-lot composition per Schedule 5 para 3 — source lots blended into this lot,
	// with the weight and grade each source contributed.
	SourceLots []LotComposition `gorm:"foreignKey:MixedLotID" json:"source_lots,omitempty"`
	// Linked mine site certification — the lot's mine site status is derived from this
	// record instead of being re-typed by the user.
	MineSiteCertificationID *uint                  `gorm:"index" json:"mine_site_certification_id,omitempty"`
	MineSiteCertification   *MineSiteCertification `gorm:"foreignKey:MineSiteCertificationID" json:"mine_site_certification,omitempty"`
	// External purchase details per Schedule 5 para 2.
	PurchaseOrderNumber *string    `gorm:"type:varchar(120)" json:"purchase_order_number,omitempty"`
	PurchaseDate        *time.Time `json:"purchase_date,omitempty"`
	AcceptedBy          *string    `gorm:"type:varchar(255)" json:"accepted_by,omitempty"`
	// Export linkage — set when the lot is included in an export shipment.
	ExportShipmentID *uint `gorm:"index" json:"export_shipment_id,omitempty"`
	// Lifecycle state of the lot in the supply chain.
	TrackingState TrackingStatus `gorm:"type:varchar(30);not null;default:'extracted'" json:"tracking_state"`
	MineralType            MineralType      `gorm:"type:varchar(50);not null;index" json:"mineral_type"`
	OreType                *string          `gorm:"type:varchar(120)" json:"ore_type,omitempty"`
	Weight                 float64          `gorm:"not null" json:"weight"`
	Unit                   string           `gorm:"type:varchar(20);not null" json:"unit"`
	Grade                  *string          `gorm:"type:varchar(80)" json:"grade,omitempty"`
	GradeValue             *float64         `gorm:"type:decimal(10,4)" json:"grade_value,omitempty"`
	GradeUnit              *string          `gorm:"type:varchar(40)" json:"grade_unit,omitempty"`
	NumberOfSacks          *int             `gorm:"type:integer" json:"number_of_sacks,omitempty"`
	SourceMineSite         string           `gorm:"type:varchar(255);not null" json:"source_mine_site"`
	MineSiteStatus         ComplianceStatus `gorm:"type:varchar(20);not null;default:'blue'" json:"mine_site_status"`
	MineOperatorName       *string          `gorm:"type:varchar(255)" json:"mine_operator_name,omitempty"`
	MinerName              *string          `gorm:"type:varchar(255)" json:"miner_name,omitempty"`
	MinerNationalID        *string          `gorm:"type:varchar(120)" json:"miner_national_id,omitempty"`
	ArtisanalLicenseNumber *string          `gorm:"type:varchar(120)" json:"artisanal_license_number,omitempty"`
	CoCSystem              string           `gorm:"type:varchar(255);not null" json:"coc_system"`
	SealNumber             *string          `gorm:"type:varchar(120)" json:"seal_number,omitempty"`
	RegisteredAt           *time.Time       `json:"registered_at,omitempty"`
	SealedAt               *time.Time       `json:"sealed_at,omitempty"`
	ShippedAt              *time.Time       `json:"shipped_at,omitempty"`
	TransporterName        *string          `gorm:"type:varchar(255)" json:"transporter_name,omitempty"`
	TransportRoute         *string          `gorm:"type:text" json:"transport_route,omitempty"`
	CurrentCustodian       *string          `gorm:"type:varchar(255)" json:"current_custodian,omitempty"`
	// The account that currently holds the lot after a multi-party handover.
	// Nullable: null means the owner still holds it (legacy/default).
	CurrentCustodianUserID *uint            `gorm:"index" json:"current_custodian_user_id,omitempty"`
	UpstreamActors         *string          `gorm:"type:text" json:"upstream_actors,omitempty"`
	TaxesFeesRoyalties     *string          `gorm:"type:text" json:"taxes_fees_royalties,omitempty"`
	VerificationOfficer    *string          `gorm:"type:varchar(255)" json:"verification_officer,omitempty"`
	DocumentationReference *string          `gorm:"type:varchar(255)" json:"documentation_reference,omitempty"`
	IsExported             bool             `gorm:"default:false" json:"is_exported"`
	UserID                 uint             `gorm:"not null;index" json:"user_id"`
	User                   User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
	DeletedAt              gorm.DeletedAt   `gorm:"index" json:"-"`
}

// LotComposition records how much weight/grade a source CoC lot contributed to a mixed lot
// (Schedule 5, paragraph 3 of S.I. No. 23 of 2023).
type LotComposition struct {
	gorm.Model
	MixedLotID          uint           `gorm:"not null;index" json:"mixed_lot_id"`
	SourceLotID         uint           `gorm:"not null;index" json:"source_lot_id"`
	SourceLot           *CoCLot        `gorm:"foreignKey:SourceLotID" json:"source_lot,omitempty"`
	WeightContributed   float64        `gorm:"not null" json:"weight_contributed"`
	GradeContributed    *float64       `gorm:"type:decimal(10,4)" json:"grade_contributed,omitempty"`
	PurchaseOrderNumber *string        `gorm:"type:varchar(120)" json:"purchase_order_number,omitempty"`
	UserID              uint           `gorm:"not null;index" json:"user_id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// ExportShipment records per-shipment ICGLR certificate application data.
type ExportShipment struct {
	gorm.Model
	ExporterLotNumber      string           `gorm:"type:varchar(120);not null;index" json:"exporter_lot_number"`
	ExporterName           string           `gorm:"type:varchar(255);not null" json:"exporter_name"`
	ExporterLicenseNumber  *string          `gorm:"type:varchar(120)" json:"exporter_license_number,omitempty"`
	ExporterStatus         ComplianceStatus `gorm:"type:varchar(20);not null;default:'blue'" json:"exporter_status"`
	CustomerName           string           `gorm:"type:varchar(255);not null" json:"customer_name"`
	CustomerAddress        *string          `gorm:"type:text" json:"customer_address,omitempty"`
	DestinationCountry     string           `gorm:"type:varchar(120);not null" json:"destination_country"`
	MaterialDescription    string           `gorm:"type:text;not null" json:"material_description"`
	Weight                 float64          `gorm:"not null" json:"weight"`
	Unit                   string           `gorm:"type:varchar(20);not null" json:"unit"`
	Grade                  *string          `gorm:"type:varchar(80)" json:"grade,omitempty"`
	IncomingLotNumbers     string           `gorm:"type:text;not null" json:"incoming_lot_numbers"`
	IncomingLotWeights     *string          `gorm:"type:text" json:"incoming_lot_weights,omitempty"`
	// CoC lots physically included in this shipment (FK on CoCLot.ExportShipmentID).
	CoCLots []CoCLot `gorm:"foreignKey:ExportShipmentID" json:"coc_lots,omitempty"`
	// Mass-balance check: percentage difference between declared shipment weight
	// and the sum of the linked lots' weights. Values above ~2% raise a compliance alert.
	WeightDiscrepancyPct *float64 `gorm:"type:decimal(8,4)" json:"weight_discrepancy_pct,omitempty"`
	TaxesFeesRoyalties     *string          `gorm:"type:text" json:"taxes_fees_royalties,omitempty"`
	SealedAt               *time.Time       `json:"sealed_at,omitempty"`
	ShippedAt              *time.Time       `json:"shipped_at,omitempty"`
	TransporterName        *string          `gorm:"type:varchar(255)" json:"transporter_name,omitempty"`
	TransportRoute         *string          `gorm:"type:text" json:"transport_route,omitempty"`
	AuthorisedOfficer      *string          `gorm:"type:varchar(255)" json:"authorised_officer,omitempty"`
	ApplicationStatus      string           `gorm:"type:varchar(40);not null;default:'draft'" json:"application_status"`
	ICGLRCertificateNumber *string          `gorm:"type:varchar(120);index" json:"icglr_certificate_number,omitempty"`
	ICGLRCertificateFile   *string          `gorm:"type:varchar(255)" json:"icglr_certificate_file,omitempty"`
	CertificateIssuedAt    *time.Time       `json:"certificate_issued_at,omitempty"`
	CertificateExpiresAt   *time.Time       `json:"certificate_expires_at,omitempty"`
	UserID                 uint             `gorm:"not null;index" json:"user_id"`
	User                   User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
	DeletedAt              gorm.DeletedAt   `gorm:"index" json:"-"`
}

// DueDiligenceReport records OECD/ICGLR due diligence reporting evidence.
type DueDiligenceReport struct {
	gorm.Model
	ReportingPeriodStart   time.Time       `gorm:"not null" json:"reporting_period_start"`
	ReportingPeriodEnd     time.Time       `gorm:"not null" json:"reporting_period_end"`
	Frequency              ReportFrequency `gorm:"type:varchar(20);not null;default:'annual'" json:"frequency"`
	ResponsiblePerson      string          `gorm:"type:varchar(255);not null" json:"responsible_person"`
	MineralChainPolicy     *string         `gorm:"type:text" json:"mineral_chain_policy,omitempty"`
	ManagementSystem       *string         `gorm:"type:text" json:"management_system,omitempty"`
	RiskAssessmentSummary  *string         `gorm:"type:text" json:"risk_assessment_summary,omitempty"`
	RiskMitigationPlan     *string         `gorm:"type:text" json:"risk_mitigation_plan,omitempty"`
	GrievanceMechanism     *string         `gorm:"type:text" json:"grievance_mechanism,omitempty"`
	SupplierCapacity       *string         `gorm:"type:text" json:"supplier_capacity,omitempty"`
	GovernmentPayments     *string         `gorm:"type:text" json:"government_payments,omitempty"`
	BeneficialOwnership    *string         `gorm:"type:text" json:"beneficial_ownership,omitempty"`
	PublishedAt            *time.Time      `json:"published_at,omitempty"`
	SubmittedToDirectorate bool            `gorm:"default:false" json:"submitted_to_directorate"`
	SubmittedAt            *time.Time      `json:"submitted_at,omitempty"`
	AttachmentReference    *string         `gorm:"type:varchar(255)" json:"attachment_reference,omitempty"`
	UserID                 uint            `gorm:"not null;index" json:"user_id"`
	User                   User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	DeletedAt              gorm.DeletedAt  `gorm:"index" json:"-"`
}

// ThirdPartyAudit tracks exporter audit status and corrective obligations.
type ThirdPartyAudit struct {
	gorm.Model
	ExporterName        string           `gorm:"type:varchar(255);not null" json:"exporter_name"`
	AuditorName         *string          `gorm:"type:varchar(255)" json:"auditor_name,omitempty"`
	AuditRequestedAt    *time.Time       `json:"audit_requested_at,omitempty"`
	AuditStartedAt      *time.Time       `json:"audit_started_at,omitempty"`
	AuditCompletedAt    *time.Time       `json:"audit_completed_at,omitempty"`
	Status              ComplianceStatus `gorm:"type:varchar(20);not null;default:'blue'" json:"status"`
	StatusExpiresAt     *time.Time       `json:"status_expires_at,omitempty"`
	ReportReference     *string          `gorm:"type:varchar(255)" json:"report_reference,omitempty"`
	Findings            *string          `gorm:"type:text" json:"findings,omitempty"`
	CorrectiveActions   *string          `gorm:"type:text" json:"corrective_actions,omitempty"`
	FollowUpRequestedAt *time.Time       `json:"follow_up_requested_at,omitempty"`
	FollowUpDueAt       *time.Time       `json:"follow_up_due_at,omitempty"`
	UserID              uint             `gorm:"not null;index" json:"user_id"`
	User                User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	DeletedAt           gorm.DeletedAt   `gorm:"index" json:"-"`
}

// ComplianceDocument stores references to scanned documents and evidence retained for compliance.
type ComplianceDocument struct {
	gorm.Model
	DocumentType  string         `gorm:"type:varchar(80);not null;index" json:"document_type"`
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Reference     string         `gorm:"type:varchar(255);not null" json:"reference"`
	RelatedEntity string         `gorm:"type:varchar(80)" json:"related_entity,omitempty"`
	RelatedID     *uint          `gorm:"index" json:"related_id,omitempty"`
	IssuedAt      *time.Time     `json:"issued_at,omitempty"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty"`
	Notes         *string        `gorm:"type:text" json:"notes,omitempty"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// Enhanced traceability models for comprehensive chain of custody

// TransportType represents the mode of transport
type TransportType string

const (
	TransportRoad TransportType = "road"
	TransportRail TransportType = "rail"
	TransportAir  TransportType = "air"
	TransportSea  TransportType = "sea"
)

// TrackingStatus represents the current status in the supply chain
type TrackingStatus string

const (
	StatusExtracted      TrackingStatus = "extracted"
	StatusStored         TrackingStatus = "stored"
	StatusInTransit      TrackingStatus = "in_transit"
	StatusProcessing     TrackingStatus = "processing"
	StatusQualityControl TrackingStatus = "quality_control"
	StatusReadyForExport TrackingStatus = "ready_for_export"
	StatusExported       TrackingStatus = "exported"
)

// AlertType represents different types of tracking alerts
type AlertType string

const (
	AlertDelay           AlertType = "delay"
	AlertRouteDeviation  AlertType = "route_deviation"
	AlertTemperature     AlertType = "temperature"
	AlertSecurity        AlertType = "security"
	AlertCompliance      AlertType = "compliance"
	AlertQuality         AlertType = "quality"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"
	SeverityMedium   AlertSeverity = "medium"
	SeverityHigh     AlertSeverity = "high"
	SeverityCritical AlertSeverity = "critical"
)

// GPSLocation stores geographical coordinates with address and timestamp
type GPSLocation struct {
	gorm.Model
	Latitude  float64   `gorm:"not null" json:"latitude"`
	Longitude float64   `gorm:"not null" json:"longitude"`
	Address   *string   `gorm:"type:varchar(255)" json:"address,omitempty"`
	Timestamp time.Time `gorm:"not null" json:"timestamp"`
}

// TransportRecord tracks vehicle movements and cargo details
type TransportRecord struct {
	gorm.Model
	TransportType       TransportType `gorm:"type:varchar(20);not null" json:"transport_type"`
	LicensePlate        string        `gorm:"type:varchar(50);not null" json:"license_plate"`
	DriverName          string        `gorm:"type:varchar(255);not null" json:"driver_name"`
	DriverLicense       string        `gorm:"type:varchar(100);not null" json:"driver_license"`
	VehicleType         string        `gorm:"type:varchar(100);not null" json:"vehicle_type"`
	VehicleCapacity     float64       `gorm:"not null" json:"vehicle_capacity"`
	
	// Route information
	OriginID            uint          `gorm:"not null" json:"origin_id"`
	Origin              GPSLocation   `gorm:"foreignKey:OriginID" json:"origin"`
	DestinationID       uint          `gorm:"not null" json:"destination_id"`
	Destination         GPSLocation   `gorm:"foreignKey:DestinationID" json:"destination"`
	PlannedDistance     float64       `gorm:"not null" json:"planned_distance"`
	ActualDistance      *float64      `json:"actual_distance,omitempty"`
	
	// Timing
	DepartureTime       time.Time     `gorm:"not null" json:"departure_time"`
	ArrivalTime         *time.Time    `json:"arrival_time,omitempty"`
	EstimatedDuration   int           `gorm:"not null" json:"estimated_duration"` // minutes
	ActualDuration      *int          `json:"actual_duration,omitempty"`          // minutes
	
	// Cargo details
	CoCLotIDs           string        `gorm:"type:text;not null" json:"coc_lot_ids"`      // JSON array of lot IDs (legacy)
	Lots                []CoCLot      `gorm:"many2many:transport_record_lots;" json:"lots,omitempty"`
	SealNumbers         string        `gorm:"type:text;not null" json:"seal_numbers"`     // JSON array of seals
	PackagingType       string        `gorm:"type:varchar(100);not null" json:"packaging_type"`
	GrossWeight         float64       `gorm:"not null" json:"gross_weight"`
	NetWeight           float64       `gorm:"not null" json:"net_weight"`
	
	// Custody handover
	HandoverFromName    string        `gorm:"type:varchar(255);not null" json:"handover_from_name"`
	HandoverFromID      string        `gorm:"type:varchar(100);not null" json:"handover_from_id"`
	HandoverToName      string        `gorm:"type:varchar(255);not null" json:"handover_to_name"`
	HandoverToID        string        `gorm:"type:varchar(100);not null" json:"handover_to_id"`
	WitnessedBy         *string       `gorm:"type:text" json:"witnessed_by,omitempty"`     // JSON array
	HandoverNotes       *string       `gorm:"type:text" json:"handover_notes,omitempty"`
	
	// Security and compliance
	SecurityEscort      bool          `gorm:"default:false" json:"security_escort"`
	EscortDetails       *string       `gorm:"type:text" json:"escort_details,omitempty"`
	RouteSecurityLevel  string        `gorm:"type:varchar(20);default:'low'" json:"route_security_level"` // low, medium, high
	TransportPermit     *string       `gorm:"type:varchar(255)" json:"transport_permit,omitempty"`
	CustomsDeclaration  *string       `gorm:"type:varchar(255)" json:"customs_declaration,omitempty"`
	InsurancePolicy     *string       `gorm:"type:varchar(255)" json:"insurance_policy,omitempty"`
	
	Status              TrackingStatus `gorm:"type:varchar(30);not null;default:'scheduled'" json:"status"`
	UserID              uint          `gorm:"not null;index" json:"user_id"`
	User                User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProcessingRecord tracks all processing operations and quality control
type ProcessingRecord struct {
	gorm.Model
	FacilityID          string        `gorm:"type:varchar(100);not null" json:"facility_id"`
	FacilityName        string        `gorm:"type:varchar(255);not null" json:"facility_name"`
	ProcessType         string        `gorm:"type:text;not null" json:"process_type"` // JSON array of ProcessingMethod
	
	// Input batches (JSON array of input details, legacy)
	InputBatches        string        `gorm:"type:text;not null" json:"input_batches"`
	// CoC lots consumed as input to this processing run.
	InputLots           []CoCLot      `gorm:"many2many:processing_record_input_lots;" json:"input_lots,omitempty"`
	// CoC lot produced as output of this processing run, if registered.
	OutputLotID         *uint         `gorm:"index" json:"output_lot_id,omitempty"`
	
	// Processing details
	Equipment           string        `gorm:"type:text;not null" json:"equipment"`
	Parameters          string        `gorm:"type:text;not null" json:"parameters"`
	Duration            int           `gorm:"not null" json:"duration"`        // minutes
	Operator            string        `gorm:"type:varchar(255);not null" json:"operator"`
	Supervisor          string        `gorm:"type:varchar(255);not null" json:"supervisor"`
	StartTime           time.Time     `gorm:"not null" json:"start_time"`
	EndTime             *time.Time    `json:"end_time,omitempty"`
	
	// Output details (JSON structure)
	OutputItems         string        `gorm:"type:text;not null" json:"output_items"`
	Yield               float64       `gorm:"not null" json:"yield"`              // percentage
	Recovery            *float64      `json:"recovery,omitempty"`                 // percentage
	WasteGenerated      float64       `gorm:"not null" json:"waste_generated"`    // kg
	
	// Quality control
	SamplesCollected    int           `gorm:"not null;default:0" json:"samples_collected"`
	AssayResults        *string       `gorm:"type:text" json:"assay_results,omitempty"`
	QualityNotes        *string       `gorm:"type:text" json:"quality_notes,omitempty"`
	Certificates        *string       `gorm:"type:text" json:"certificates,omitempty"` // JSON array
	
	Status              TrackingStatus `gorm:"type:varchar(30);not null;default:'scheduled'" json:"status"`
	UserID              uint          `gorm:"not null;index" json:"user_id"`
	User                User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TrackingAlert stores alerts and notifications for traceability events
type TrackingAlert struct {
	gorm.Model
	LotID               uint          `gorm:"not null;index" json:"lot_id"`
	LotType             string        `gorm:"type:varchar(20);not null" json:"lot_type"` // inventory, coc_lot
	AlertType           AlertType     `gorm:"type:varchar(30);not null" json:"alert_type"`
	Severity            AlertSeverity `gorm:"type:varchar(20);not null" json:"severity"`
	Message             string        `gorm:"type:text;not null" json:"message"`
	Resolved            bool          `gorm:"default:false" json:"resolved"`
	ResolvedAt          *time.Time    `json:"resolved_at,omitempty"`
	ResolvedBy          *string       `gorm:"type:varchar(255)" json:"resolved_by,omitempty"`
	UserID              uint          `gorm:"not null;index" json:"user_id"`
	User                User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// RealTimeTracking stores current status and location of lots
type RealTimeTracking struct {
	gorm.Model
	LotID                uint           `gorm:"not null;index" json:"lot_id"`
	LotType              string         `gorm:"type:varchar(20);not null" json:"lot_type"`   // inventory, coc_lot
	CurrentLocationID    *uint          `json:"current_location_id,omitempty"`
	CurrentLocation      *GPSLocation   `gorm:"foreignKey:CurrentLocationID" json:"current_location,omitempty"`
	CurrentCustodian     string         `gorm:"type:varchar(255);not null" json:"current_custodian"`
	Status               TrackingStatus `gorm:"type:varchar(30);not null" json:"status"`
	LastUpdated          time.Time      `gorm:"not null" json:"last_updated"`
	NextDestination      *string        `gorm:"type:varchar(255)" json:"next_destination,omitempty"`
	EstimatedArrival     *time.Time     `json:"estimated_arrival,omitempty"`
	TransportRecordID    *uint          `gorm:"index" json:"transport_record_id,omitempty"`
	TransportRecord      *TransportRecord `gorm:"foreignKey:TransportRecordID" json:"transport_record,omitempty"`
	ProcessingRecordID   *uint          `gorm:"index" json:"processing_record_id,omitempty"`
	ProcessingRecord     *ProcessingRecord `gorm:"foreignKey:ProcessingRecordID" json:"processing_record,omitempty"`
	UserID               uint           `gorm:"not null;index" json:"user_id"`
	User                 User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

// CustodyTransfer records each handover in the chain of custody
type CustodyTransfer struct {
	gorm.Model
	CoCLotID            uint         `gorm:"not null;index" json:"coc_lot_id"`
	CoCLot              CoCLot       `gorm:"foreignKey:CoCLotID" json:"coc_lot,omitempty"`
	FromCustodian       string       `gorm:"type:varchar(255);not null" json:"from_custodian"`
	ToCustodian         string       `gorm:"type:varchar(255);not null" json:"to_custodian"`
	TransferDate        time.Time    `gorm:"not null" json:"transfer_date"`
	TransferLocationID  *uint        `json:"transfer_location_id,omitempty"`
	TransferLocation    *GPSLocation `gorm:"foreignKey:TransferLocationID" json:"transfer_location,omitempty"`
	Witness             *string      `gorm:"type:varchar(255)" json:"witness,omitempty"`
	TransferReason      string       `gorm:"type:text;not null" json:"transfer_reason"`
	ConditionNotes      *string      `gorm:"type:text" json:"condition_notes,omitempty"`
	Photos              *string      `gorm:"type:text" json:"photos,omitempty"`        // JSON array of photo URLs
	DigitalSignature    *string      `gorm:"type:text" json:"digital_signature,omitempty"`
	UserID              uint         `gorm:"not null;index" json:"user_id"`
	User                User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

// PhotoRecord stores photos taken during traceability processes
type PhotoRecord struct {
	gorm.Model
	PhotoURL            string         `gorm:"type:varchar(500);not null" json:"photo_url"`
	PhotoType           string         `gorm:"type:varchar(50);not null" json:"photo_type"` // extraction, transport, processing, quality_control, custody_transfer, compliance
	Description         *string        `gorm:"type:text" json:"description,omitempty"`
	LocationID          *uint          `json:"location_id,omitempty"`
	Location            *GPSLocation   `gorm:"foreignKey:LocationID" json:"location,omitempty"`
	TakenBy             string         `gorm:"type:varchar(255);not null" json:"taken_by"`
	TakenAt             time.Time      `gorm:"not null" json:"taken_at"`
	RelatedEntity       string         `gorm:"type:varchar(50);not null" json:"related_entity"` // coc_lot, transport_record, processing_record
	RelatedID           uint           `gorm:"not null;index" json:"related_id"`
	UserID              uint           `gorm:"not null;index" json:"user_id"`
	User                User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// Stakeholder manages all parties in the supply chain
type Stakeholder struct {
	gorm.Model
	Name                string         `gorm:"type:varchar(255);not null" json:"name"`
	Type                string         `gorm:"type:varchar(50);not null" json:"type"` // miner, processor, transporter, trader, exporter, customs, inspector
	Email               *string        `gorm:"type:varchar(255)" json:"email,omitempty"`
	Phone               *string        `gorm:"type:varchar(50)" json:"phone,omitempty"`
	Address             *string        `gorm:"type:text" json:"address,omitempty"`
	Licenses            string         `gorm:"type:text" json:"licenses"`             // JSON array of license objects
	ComplianceStatus    ComplianceStatus `gorm:"type:varchar(20);not null;default:'blue'" json:"compliance_status"`
	VerifiedBy          *string        `gorm:"type:varchar(255)" json:"verified_by,omitempty"`
	VerifiedAt          *time.Time     `json:"verified_at,omitempty"`
	UserID              uint           `gorm:"not null;index" json:"user_id"`
	User                User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// LotPassport is a read-only assembly of the complete traceability chain for one CoC lot:
// production → certification → custody → transport → processing → export.
// It is built on demand and is not persisted.
type LotPassport struct {
	Lot               *CoCLot                `json:"lot"`
	Certification     *MineSiteCertification `json:"certification,omitempty"`
	ProductionRecords []InventoryItem        `json:"production_records"`
	Composition       []LotComposition       `json:"composition"`
	UsedInLots        []LotComposition       `json:"used_in_lots"`
	CustodyTransfers  []CustodyTransfer      `json:"custody_transfers"`
	TransportRecords  []TransportRecord      `json:"transport_records"`
	ProcessingRecords []ProcessingRecord     `json:"processing_records"`
	ExportShipment    *ExportShipment        `json:"export_shipment,omitempty"`
	Documents         []ComplianceDocument   `json:"documents"`
	Alerts            []TrackingAlert        `json:"alerts"`
	Tracking          *RealTimeTracking      `json:"tracking,omitempty"`
	// ChainComplete is true when the lot is linked to production records (or source lots)
	// and a mine site certification, i.e. origin is fully demonstrated.
	ChainComplete bool `json:"chain_complete"`
}
