package data

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

// Models wraps all repository interfaces
type Models struct {
	User      UserInterface
	Income    IncomeInterface
	Expense   ExpenseInterface
	Inventory InventoryInterface
}
