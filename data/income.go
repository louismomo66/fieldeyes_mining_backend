package data

import (
	"gorm.io/gorm"
)

// IncomeRepository implements IncomeInterface using GORM
type IncomeRepository struct {
	db *gorm.DB
}

// NewIncomeRepository creates a new instance of IncomeRepository
func NewIncomeRepository(db *gorm.DB) IncomeInterface {
	return &IncomeRepository{db: db}
}

// GetAll retrieves all income records for a user
func (r *IncomeRepository) GetAll(userID uint) ([]*Income, error) {
	var incomes []*Income
	result := r.db.Where("user_id = ?", userID).Order("date DESC").Find(&incomes)
	return incomes, result.Error
}

// GetOne retrieves a specific income record by ID for a user
func (r *IncomeRepository) GetOne(id uint, userID uint) (*Income, error) {
	var income Income
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&income)
	if result.Error != nil {
		return nil, result.Error
	}
	return &income, nil
}

// Insert creates a new income record
func (r *IncomeRepository) Insert(income *Income) (uint, error) {
	// Calculate total amount
	income.TotalAmount = income.Quantity * income.PricePerUnit

	// Calculate amount due
	income.AmountDue = income.TotalAmount - income.AmountPaid

	result := r.db.Create(income)
	return income.ID, result.Error
}

// Update updates an existing income record
func (r *IncomeRepository) Update(income *Income) error {
	// Recalculate total amount and amount due
	income.TotalAmount = income.Quantity * income.PricePerUnit
	income.AmountDue = income.TotalAmount - income.AmountPaid

	result := r.db.Save(income)
	return result.Error
}

// Delete soft deletes an income record
func (r *IncomeRepository) Delete(id uint, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Income{})
	return result.Error
}

// GetByDateRange retrieves income records within a date range
func (r *IncomeRepository) GetByDateRange(userID uint, startDate, endDate string) ([]*Income, error) {
	var incomes []*Income
	result := r.db.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("date DESC").Find(&incomes)
	return incomes, result.Error
}

// GetFinancialSummary calculates financial summary for a user
func (r *IncomeRepository) GetFinancialSummary(userID uint) (*FinancialSummary, error) {
	var summary FinancialSummary

	// Get total income
	var totalIncome float64
	result := r.db.Model(&Income{}).Where("user_id = ? AND deleted_at IS NULL", userID).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalIncome)
	if result.Error != nil {
		return nil, result.Error
	}
	summary.TotalIncome = totalIncome

	// Get total receivables (unpaid amounts)
	var totalReceivables float64
	result = r.db.Model(&Income{}).Where("user_id = ? AND deleted_at IS NULL AND payment_status IN (?, ?)", userID, PaymentUnpaid, PaymentPartial).
		Select("COALESCE(SUM(amount_due), 0)").Scan(&totalReceivables)
	if result.Error != nil {
		return nil, result.Error
	}
	summary.TotalReceivables = totalReceivables

	return &summary, nil
}

// GetMonthlyData retrieves monthly income data for a year
func (r *IncomeRepository) GetMonthlyData(userID uint, year int) ([]*MonthlyData, error) {
	var monthlyData []*MonthlyData

	query := `
		SELECT 
			TO_CHAR(date, 'YYYY-MM') as month,
			COALESCE(SUM(total_amount), 0) as income
		FROM incomes 
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
