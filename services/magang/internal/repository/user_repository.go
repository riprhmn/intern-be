package repository

import (
	"errors"
	"strings"
	"time"

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
	NoEmp      string
	IdEmp      string
	CodeName   string
	FullName   string
	Username   string
	Email      string
	Title      string
	Company    string
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

	q = applyUserFilters(q, filter)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 1000 {
		limit = 1000
	}

	offset := (page - 1) * limit
	if err := q.Order("id ASC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Column names are fixed here; only values come from the request.
func applyUserFilters(q *gorm.DB, filter UserFilter) *gorm.DB {
	fields := []struct {
		column, value string
		exact         bool
	}{
		{"no_emp", filter.NoEmp, false}, {"id_emp", filter.IdEmp, false},
		{"code_name", filter.CodeName, false},
		{"full_name", filter.FullName, false}, {"username", filter.Username, false},
		{"email", filter.Email, false}, {"title", filter.Title, true},
		{"company", filter.Company, true}, {"section", filter.Section, true},
		{"department", filter.Department, true}, {"division", filter.Division, true},
	}
	if value := strings.TrimSpace(filter.Search); value != "" {
		columns := []string{"no_emp", "id_emp", "code_name", "full_name", "username", "email", "title", "company", "section", "department", "division", "role"}
		parts := []string{}
		args := []interface{}{}
		for _, column := range columns {
			parts = append(parts, column+" ILIKE ?")
			args = append(args, "%"+value+"%")
		}
		q = q.Where("("+strings.Join(parts, " OR ")+")", args...)
	}
	for _, field := range fields {
		value := strings.TrimSpace(field.value)
		if value == "" {
			continue
		}
		if field.exact {
			q = q.Where("LOWER(BTRIM("+field.column+")) = LOWER(?)", value)
		} else {
			q = q.Where(field.column+" ILIKE ?", "%"+value+"%")
		}
	}
	if role := strings.TrimSpace(filter.Role); role != "" && role != "all" {
		q = q.Where("role = ?", role)
	}
	return q
}

func (r *UserRepository) FilterOptions() (map[string][]string, error) {
	options := map[string][]string{}
	for _, column := range []string{"department", "section", "division", "title", "company"} {
		values := []string{}
		if err := r.DB.Model(&models.User{}).Where("is_active = ? AND BTRIM("+column+") <> ''", true).
			Distinct("BTRIM("+column+")").Order("BTRIM("+column+")").Pluck("BTRIM("+column+")", &values).Error; err != nil {
			return nil, err
		}
		options[column] = values
	}
	return options, nil
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

func (r *UserRepository) GetByIdentity(identity string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("(username = ? OR email = ?) AND is_active = ?", identity, identity, true).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) CodeNameExists(codeName string, excludeID uint64) (bool, error) {
	var count int64
	query := r.DB.Model(&models.User{}).Where("LOWER(BTRIM(code_name)) = LOWER(?)", strings.TrimSpace(codeName))
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
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

func (r *UserRepository) SaveResetToken(userID uint64, token string, expiry time.Time) error {
	return r.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"reset_token":        token,
		"reset_token_expiry": expiry,
	}).Error
}

func (r *UserRepository) GetByResetToken(token string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("reset_token = ? AND reset_token_expiry > ? AND is_active = ?", token, time.Now(), true).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) UpdatePassword(userID uint64, password string) error {
	return r.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":           password,
		"reset_token":        "",
		"reset_token_expiry": nil,
	}).Error
}
