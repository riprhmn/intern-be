package repository

import (
	"errors"
	"strings"

	"magang-be/services/magang/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

type UserFilter struct {
	Search     string
	Role       string
	Section    string
	Department string
	Division   string
	Page       int
	Limit      int
}

func (r *UserRepository) GetAll(filter UserFilter) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	q := r.DB.Model(&models.User{}).Where("is_active = ?", true)

	if strings.TrimSpace(filter.Search) != "" {
		like := "%" + strings.TrimSpace(filter.Search) + "%"
		q = q.Where("full_name ILIKE ? OR username ILIKE ? OR no_emp ILIKE ? OR id_emp ILIKE ? OR email ILIKE ? OR title ILIKE ? OR company ILIKE ?", like, like, like, like, like, like, like)
	}

	if strings.TrimSpace(filter.Role) != "" && filter.Role != "all" {
		q = q.Where("role = ?", filter.Role)
	}

	if strings.TrimSpace(filter.Section) != "" {
		like := "%" + strings.TrimSpace(filter.Section) + "%"
		q = q.Where("section ILIKE ?", like)
	}

	if strings.TrimSpace(filter.Department) != "" {
		like := "%" + strings.TrimSpace(filter.Department) + "%"
		q = q.Where("department ILIKE ?", like)
	}

	if strings.TrimSpace(filter.Division) != "" {
		like := "%" + strings.TrimSpace(filter.Division) + "%"
		q = q.Where("division ILIKE ?", like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	if err := q.Order("id ASC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) GetByID(id uint64) (*models.User, error) {
	var user models.User
	err := r.DB.Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("username = ? AND is_active = ?", username, true).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) Update(user *models.User) error {
	return r.DB.Save(user).Error
}

func (r *UserRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.User{}, id).Error
}
