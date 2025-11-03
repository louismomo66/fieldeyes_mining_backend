package data

import (
	"time"

	"gorm.io/gorm"
)

// InventoryRepository implements InventoryInterface using GORM
type InventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository creates a new instance of InventoryRepository
func NewInventoryRepository(db *gorm.DB) InventoryInterface {
	return &InventoryRepository{db: db}
}

// GetAll retrieves all inventory items for a user
func (r *InventoryRepository) GetAll(userID uint) ([]*InventoryItem, error) {
	var items []*InventoryItem
	result := r.db.Where("user_id = ?", userID).Order("name ASC").Find(&items)
	return items, result.Error
}

// GetOne retrieves a specific inventory item by ID for a user
func (r *InventoryRepository) GetOne(id uint, userID uint) (*InventoryItem, error) {
	var item InventoryItem
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

// Insert creates a new inventory item
func (r *InventoryRepository) Insert(item *InventoryItem) (uint, error) {
	item.LastUpdated = time.Now()
	result := r.db.Create(item)
	return item.ID, result.Error
}

// Update updates an existing inventory item
func (r *InventoryRepository) Update(item *InventoryItem) error {
	item.LastUpdated = time.Now()
	result := r.db.Save(item)
	return result.Error
}

// Delete soft deletes an inventory item
func (r *InventoryRepository) Delete(id uint, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&InventoryItem{})
	return result.Error
}

// GetLowStockItems retrieves items that are below minimum stock level
func (r *InventoryRepository) GetLowStockItems(userID uint) ([]*InventoryItem, error) {
	var items []*InventoryItem
	result := r.db.Where("user_id = ? AND quantity <= min_stock_level", userID).
		Order("quantity ASC").Find(&items)
	return items, result.Error
}

// UpdateQuantity updates the quantity of an inventory item
func (r *InventoryRepository) UpdateQuantity(id uint, userID uint, quantity float64) error {
	result := r.db.Model(&InventoryItem{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"quantity":     quantity,
			"last_updated": time.Now(),
		})
	return result.Error
}



