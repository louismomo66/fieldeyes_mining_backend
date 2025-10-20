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

// MineralType represents the type of mineral
type MineralType string

const (
	MineralGold    MineralType = "gold"
	MineralCopper  MineralType = "copper"
	MineralCobalt  MineralType = "cobalt"
	MineralDiamond MineralType = "diamond"
	MineralOther   MineralType = "other"
)

// ExpenseCategory represents the category of expense
type ExpenseCategory string

const (
	ExpenseEquipment   ExpenseCategory = "equipment"
	ExpenseLabor       ExpenseCategory = "labor"
	ExpenseChemicals   ExpenseCategory = "chemicals"
	ExpenseFuel        ExpenseCategory = "fuel"
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
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role      UserRole       `gorm:"type:varchar(50);default:'standard'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// OTP fields for password reset
	OTPCode      string     `gorm:"type:varchar(6)" json:"-"`
	OTPExpiresAt *time.Time `json:"-"`
}

// Income represents an income transaction
type Income struct {
	gorm.Model
	Date            time.Time      `gorm:"not null" json:"date"`
	MineralType     MineralType    `gorm:"type:varchar(50);not null" json:"mineral_type"`
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

// InventoryItem represents an inventory item
type InventoryItem struct {
	gorm.Model
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	Type          string         `gorm:"type:varchar(20);not null" json:"type"` // "mineral" or "supply"
	Quantity      float64        `gorm:"not null" json:"quantity"`
	Unit          string         `gorm:"type:varchar(20);not null" json:"unit"`
	MinStockLevel float64        `gorm:"not null" json:"min_stock_level"`
	CurrentValue  float64        `gorm:"not null" json:"current_value"`
	LastUpdated   time.Time      `gorm:"not null" json:"last_updated"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
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
