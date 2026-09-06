package service

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

var allowedUserRoles = map[string]struct{}{
	"admin":          {},
	"user":           {},
	"approval":       {},
	"external_audit": {},
}

func isAllowedUserRole(role string) bool {
	_, ok := allowedUserRoles[role]
	return ok
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAll(filter repository.UserFilter) ([]*models.User, int64, error) {
	return s.repo.GetAll(filter)
}

func (s *UserService) GetByID(id uint64) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) GetByUsername(username string) (*models.User, error) {
	return s.repo.GetByUsername(username)
}

func (s *UserService) VerifyPassword(plain, hashed string) bool {
	// Support bcrypt accounts and legacy plaintext seed accounts.
	if strings.HasPrefix(hashed, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
	}
	return plain == hashed
}

func (s *UserService) Create(req CreateUserRequest) error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	if strings.TrimSpace(req.FullName) == "" {
		return errors.New("full_name is required")
	}

	existing, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("username already exists")
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "user"
	}
	if !isAllowedUserRole(role) {
		return errors.New("invalid role")
	}

	user := &models.User{
		NoEmp:      req.NoEmp,
		IdEmp:      req.IdEmp,
		Username:   strings.TrimSpace(req.Username),
		Password:   req.Password,
		FullName:   strings.TrimSpace(req.FullName),
		Email:      strings.TrimSpace(req.Email),
		Title:      strings.TrimSpace(req.Title),
		Company:    strings.TrimSpace(req.Company),
		Role:       role,
		Section:    req.Section,
		Department: req.Department,
		Division:   req.Division,
		IsActive:   true,
	}
	return s.repo.Create(user)
}

func (s *UserService) Update(id uint64, req UpdateUserRequest) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if req.FullName != "" {
		user.FullName = strings.TrimSpace(req.FullName)
	}
	if req.Role != "" {
		role := strings.TrimSpace(req.Role)
		if !isAllowedUserRole(role) {
			return errors.New("invalid role")
		}
		user.Role = role
	}
	if req.Password != "" {
		user.Password = req.Password
	}
	user.NoEmp = req.NoEmp
	user.IdEmp = req.IdEmp
	user.Email = strings.TrimSpace(req.Email)
	user.Title = strings.TrimSpace(req.Title)
	user.Company = strings.TrimSpace(req.Company)
	user.Section = req.Section
	user.Department = req.Department
	user.Division = req.Division

	return s.repo.Update(user)
}

func (s *UserService) Delete(id uint64) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	return s.repo.Delete(id)
}

// --- Request Types ---

type CreateUserRequest struct {
	NoEmp      string `json:"no_emp"`
	IdEmp      string `json:"id_emp"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Title      string `json:"title"`
	Company    string `json:"company"`
	Role       string `json:"role"`
	Section    string `json:"section"`
	Department string `json:"department"`
	Division   string `json:"division"`
}

type UpdateUserRequest struct {
	NoEmp      string `json:"no_emp"`
	IdEmp      string `json:"id_emp"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Title      string `json:"title"`
	Company    string `json:"company"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	Section    string `json:"section"`
	Department string `json:"department"`
	Division   string `json:"division"`
}
