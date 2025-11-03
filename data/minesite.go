package data

import (
	"gorm.io/gorm"
)

// MineSiteInterface defines the methods for mine site information
type MineSiteInterface interface {
	GetByUserID(userID uint) (*MineSiteInfo, error)
	Insert(info *MineSiteInfo) (uint, error)
	Update(info *MineSiteInfo) error
}

// MineSiteRepository implements MineSiteInterface using GORM
type MineSiteRepository struct {
	db *gorm.DB
}

// NewMineSiteRepository creates a new instance of MineSiteRepository
func NewMineSiteRepository(db *gorm.DB) MineSiteInterface {
	return &MineSiteRepository{db: db}
}

// GetByUserID retrieves mine site information for a user
func (r *MineSiteRepository) GetByUserID(userID uint) (*MineSiteInfo, error) {
	var info MineSiteInfo
	result := r.db.Where("user_id = ?", userID).First(&info)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Return nil if not found (not an error)
		}
		return nil, result.Error
	}
	return &info, nil
}

// Insert creates a new mine site information record
func (r *MineSiteRepository) Insert(info *MineSiteInfo) (uint, error) {
	result := r.db.Create(info)
	return info.ID, result.Error
}

// Update updates an existing mine site information record
func (r *MineSiteRepository) Update(info *MineSiteInfo) error {
	result := r.db.Save(info)
	return result.Error
}
