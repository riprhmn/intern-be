package repository

import (
	"errors"

	"magang-be/services/magang/internal/models"

	"gorm.io/gorm"
)

type ChangeNoteRepository struct {
	DB *gorm.DB
}

func NewChangeNoteRepository(db *gorm.DB) *ChangeNoteRepository {
	return &ChangeNoteRepository{DB: db}
}

func (r *ChangeNoteRepository) GetAll(search string, userID uint64, role string, offset, limit int) ([]models.ChangeNote, int64, error) {
	base := r.DB.Model(&models.ChangeNote{})

	if role != "admin" && userID > 0 {
		base = base.Where("user_id = ?", userID)
	}

	if search != "" {
		like := "%" + search + "%"
		base = base.Where("initiator ILIKE ? OR emp_id_initiator ILIKE ? OR department ILIKE ? OR change_pertains_to ILIKE ?", like, like, like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var changeNotes []models.ChangeNote
	err := base.Order("id DESC").Offset(offset).Limit(limit).Find(&changeNotes).Error
	return changeNotes, total, err
}

func (r *ChangeNoteRepository) GetByID(id uint64) (*models.ChangeNote, error) {
	var cn models.ChangeNote
	err := r.DB.First(&cn, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &cn, err
}

func (r *ChangeNoteRepository) Create(cn *models.ChangeNote) error {
	return r.DB.Create(cn).Error
}

func (r *ChangeNoteRepository) Update(cn *models.ChangeNote) error {
	return r.DB.Save(cn).Error
}

func (r *ChangeNoteRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.ChangeNote{}, id).Error
}
