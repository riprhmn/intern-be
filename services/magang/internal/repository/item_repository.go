package repository

import (
	"errors"

	"magang-be/services/magang/internal/models"

	"gorm.io/gorm"
)

type ItemRepository struct {
	DB *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{DB: db}
}

func (r *ItemRepository) GetAll(search string, offset, limit int) ([]models.Item, int64, error) {
	base := r.DB.Model(&models.Item{})

	if search != "" {
		like := "%" + search + "%"
		base = base.Where("name ILIKE ?", like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.Item
	err := base.Order("id ASC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *ItemRepository) GetByID(id uint64) (*models.Item, error) {
	var item models.Item
	err := r.DB.First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

func (r *ItemRepository) Create(item *models.Item) error {
	return r.DB.Create(item).Error
}

func (r *ItemRepository) Update(item *models.Item) error {
	return r.DB.Save(item).Error
}

func (r *ItemRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.Item{}, id).Error
}
