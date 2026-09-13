package repository

import (
	"magang-be/services/magang/internal/models"

	"gorm.io/gorm"
)

type ANCRRepository struct {
	DB *gorm.DB
}

func NewANCRRepository(db *gorm.DB) *ANCRRepository {
	return &ANCRRepository{DB: db}
}

// GetAll — semua user yang login bisa lihat semua data ANCR
func (r *ANCRRepository) GetAll(search string, offset, limit int) ([]models.ANCRRequest, int64, error) {
	base := r.DB.Model(&models.ANCRRequest{})
	if search != "" {
		like := "%" + search + "%"
		base = base.Where(
			"registration_number ILIKE ? OR audit_no ILIKE ? OR initiator ILIKE ? OR auditor ILIKE ? OR auditee ILIKE ?",
			like, like, like, like, like,
		)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.ANCRRequest
	err := base.Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ANCRRepository) GetByID(id uint64) (*models.ANCRRequest, error) {
	var row models.ANCRRequest
	if err := r.DB.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ANCRRepository) Create(row *models.ANCRRequest) error {
	return r.DB.Create(row).Error
}

func (r *ANCRRepository) Update(row *models.ANCRRequest) error {
	return r.DB.Save(row).Error
}

func (r *ANCRRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.ANCRRequest{}, id).Error
}
