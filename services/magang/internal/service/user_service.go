package service

import (
	"errors"
	"fmt"
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
	// Existing accounts may use bcrypt while development seed accounts are plaintext.
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
	codeName := strings.ToUpper(strings.TrimSpace(req.CodeName))
	if role == "approval" && codeName == "" {
		return errors.New("code_name is required for approval user")
	}
	if codeName != "" {
		exists, err := s.repo.CodeNameExists(codeName, 0)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("code_name already exists")
		}
	}

	user := &models.User{
		NoEmp:      req.NoEmp,
		IdEmp:      req.IdEmp,
		CodeName:   codeName,
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
	codeName := strings.ToUpper(strings.TrimSpace(req.CodeName))
	if user.Role == "approval" && codeName == "" {
		return errors.New("code_name is required for approval user")
	}
	if codeName != "" {
		exists, err := s.repo.CodeNameExists(codeName, id)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("code_name already exists")
		}
	}
	user.CodeName = codeName
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
	CodeName   string `json:"code_name"`
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
	CodeName   string `json:"code_name"`
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

func (s *UserService) FilterOptions() (map[string][]string, error) {
	return s.repo.FilterOptions()
}

type OrganizationSelection struct {
	EndManager string `json:"endManager"`
	MQDManager string `json:"mqdManager"`
	SHOHead    string `json:"shoHead"`
	TDIVHead   string `json:"tdivHead"`
	GMM        string `json:"gmm"`
}

type organizationRule struct {
	Key        string
	Label      string
	Title      string
	Department string
	Section    string
	Division   string
}

var organizationRules = []organizationRule{
	{Key: "endManager", Label: "END Manager", Title: "Manager", Department: "END"},
	{Key: "mqdManager", Label: "MQD Manager", Title: "Manager", Department: "MQD"},
	{Key: "shoHead", Label: "SHO Head", Title: "Section Head", Department: "SHO"},
	{Key: "tdivHead", Label: "TDIV Head", Title: "Division Head", Division: "TDIV"},
	{Key: "gmm", Label: "GMM", Title: "GMM"},
}

func normalizedOrganizationField(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func matchesOrganizationRule(user *models.User, rule organizationRule) bool {
	if user == nil || !user.IsActive || user.Role != "approval" || normalizedOrganizationField(user.Title) != normalizedOrganizationField(rule.Title) {
		return false
	}
	if rule.Department != "" && normalizedOrganizationField(user.Department) != normalizedOrganizationField(rule.Department) {
		return false
	}
	if rule.Section != "" && normalizedOrganizationField(user.Section) != normalizedOrganizationField(rule.Section) {
		return false
	}
	return rule.Division == "" || normalizedOrganizationField(user.Division) == normalizedOrganizationField(rule.Division)
}

func (s *UserService) OrganizationOptions() (map[string][]*models.User, error) {
	options := make(map[string][]*models.User, len(organizationRules))
	users := []*models.User{}
	for page := 1; ; page++ {
		rows, total, err := s.repo.GetAll(repository.UserFilter{Role: "approval", Page: page, Limit: 100})
		if err != nil {
			return nil, err
		}
		users = append(users, rows...)
		if int64(page*100) >= total {
			break
		}
	}
	for _, rule := range organizationRules {
		options[rule.Key] = []*models.User{}
		for _, user := range users {
			if matchesOrganizationRule(user, rule) {
				options[rule.Key] = append(options[rule.Key], user)
			}
		}
	}
	return options, nil
}

func (s *UserService) ValidateOrganizationSelection(selection OrganizationSelection) error {
	values := map[string]string{
		"endManager": selection.EndManager,
		"mqdManager": selection.MQDManager,
		"shoHead":    selection.SHOHead,
		"tdivHead":   selection.TDIVHead,
		"gmm":        selection.GMM,
	}
	for _, rule := range organizationRules {
		username := strings.TrimSpace(values[rule.Key])
		if username == "" {
			continue
		}
		user, err := s.repo.GetByUsername(username)
		if err != nil {
			return err
		}
		if !matchesOrganizationRule(user, rule) {
			return fmt.Errorf("akun untuk %s tidak sesuai data Title, Department, Section, atau Division di Manage Account", rule.Label)
		}
	}
	return nil
}
