package data

import (
	"gorm.io/gorm"
)

// ExpenseRepository implements ExpenseInterface using GORM
type ExpenseRepository struct {
	db *gorm.DB
}

// NewExpenseRepository creates a new instance of ExpenseRepository
func NewExpenseRepository(db *gorm.DB) ExpenseInterface {
	return &ExpenseRepository{db: db}
}

// GetAll retrieves all expense records for a user
func (r *ExpenseRepository) GetAll(userID uint) ([]*Expense, error) {
	var expenses []*Expense
	result := r.db.Where("user_id = ?", userID).Order("date DESC").Find(&expenses)
	return expenses, result.Error
}

// GetOne retrieves a specific expense record by ID for a user
func (r *ExpenseRepository) GetOne(id uint, userID uint) (*Expense, error) {
	var expense Expense
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&expense)
	if result.Error != nil {
		return nil, result.Error
	}
	return &expense, nil
}

// Insert creates a new expense record
func (r *ExpenseRepository) Insert(expense *Expense) (uint, error) {
	// Calculate amount due
	expense.AmountDue = expense.Amount - expense.AmountPaid

	result := r.db.Create(expense)
	return expense.ID, result.Error
}

// Update updates an existing expense record
func (r *ExpenseRepository) Update(expense *Expense) error {
	// Recalculate amount due
	expense.AmountDue = expense.Amount - expense.AmountPaid

	result := r.db.Save(expense)
	return result.Error
}

// Delete soft deletes an expense record
func (r *ExpenseRepository) Delete(id uint, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Expense{})
	return result.Error
}

// GetByDateRange retrieves expense records within a date range
func (r *ExpenseRepository) GetByDateRange(userID uint, startDate, endDate string) ([]*Expense, error) {
	var expenses []*Expense
	result := r.db.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("date DESC").Find(&expenses)
	return expenses, result.Error
}

// GetCategoryBreakdown retrieves expense breakdown by category
func (r *ExpenseRepository) GetCategoryBreakdown(userID uint) ([]*CategoryBreakdown, error) {
	var breakdown []*CategoryBreakdown

	query := `
		SELECT 
			category,
			COALESCE(SUM(amount), 0) as amount
		FROM expenses 
		WHERE user_id = ? AND deleted_at IS NULL
		GROUP BY category
		ORDER BY amount DESC
	`

	result := r.db.Raw(query, userID).Scan(&breakdown)
	if result.Error != nil {
		return nil, result.Error
	}

	// Calculate total and percentages
	var totalAmount float64
	for _, item := range breakdown {
		totalAmount += item.Amount
	}

	for _, item := range breakdown {
		if totalAmount > 0 {
			item.Percentage = (item.Amount / totalAmount) * 100
		}
	}

	return breakdown, nil
}

// GetMonthlyData retrieves monthly expense data for a year
func (r *ExpenseRepository) GetMonthlyData(userID uint, year int) ([]*MonthlyData, error) {
	var monthlyData []*MonthlyData

	query := `
		SELECT 
			TO_CHAR(date, 'YYYY-MM') as month,
			COALESCE(SUM(amount), 0) as expenses
		FROM expenses 
		WHERE user_id = ? AND EXTRACT(YEAR FROM date) = ?
		GROUP BY TO_CHAR(date, 'YYYY-MM')
		ORDER BY month
	`

	result := r.db.Raw(query, userID, year).Scan(&monthlyData)
	if result.Error != nil {
		return nil, result.Error
	}

	return monthlyData, nil
}

// GetFinancialSummary calculates financial summary for expenses
func (r *ExpenseRepository) GetFinancialSummary(userID uint) (*FinancialSummary, error) {
	var summary FinancialSummary

	// Get total expenses
	var totalExpenses float64
	result := r.db.Model(&Expense{}).Where("user_id = ?", userID).Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)
	if result.Error != nil {
		return nil, result.Error
	}
	summary.TotalExpenses = totalExpenses

	// Get total payables (unpaid amounts)
	var totalPayables float64
	result = r.db.Model(&Expense{}).Where("user_id = ? AND payment_status IN (?, ?)", userID, PaymentUnpaid, PaymentPartial).
		Select("COALESCE(SUM(amount_due), 0)").Scan(&totalPayables)
	if result.Error != nil {
		return nil, result.Error
	}
	summary.TotalReceivables = totalPayables // Using TotalReceivables field for payables

	return &summary, nil
}
