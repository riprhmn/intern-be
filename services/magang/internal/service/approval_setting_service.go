package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"magang-be/services/magang/internal/models"

	"gorm.io/gorm"
)

type ApprovalSettingService struct {
	db *gorm.DB
}

func NewApprovalSettingService(db *gorm.DB) *ApprovalSettingService {
	return &ApprovalSettingService{db: db}
}

type CodeMappingRequest struct {
	Seq1 string `json:"seq_1"`
	Seq2 string `json:"seq_2"`
	Seq3 string `json:"seq_3"`
	Seq4 string `json:"seq_4"`
}

type DelegationRequest struct {
	FromUserID   uint64     `json:"from_user_id"`
	ToUserID     uint64     `json:"to_user_id"`
	ApprovalType string     `json:"approval_type"`
	SectionCode  string     `json:"section_code"`
	StartsAt     *time.Time `json:"starts_at"`
	EndsAt       *time.Time `json:"ends_at"`
}

func normalizeApprovalScope(approvalType, sectionCode string, allowAll bool) (string, string, error) {
	approvalType = strings.ToUpper(strings.TrimSpace(approvalType))
	sectionCode = strings.ToUpper(strings.TrimSpace(sectionCode))
	validType := approvalType == "CN" || approvalType == "ANCR" || (allowAll && approvalType == "ALL")
	if !validType {
		return "", "", errors.New("type approval harus CN, ANCR, atau ALL")
	}
	if sectionCode == "" {
		if allowAll {
			sectionCode = "ALL"
		} else {
			return "", "", errors.New("ID seksi wajib diisi")
		}
	}
	return approvalType, sectionCode, nil
}

func (s *ApprovalSettingService) Options() (map[string]interface{}, error) {
	users := []models.User{}
	if err := s.db.Where("is_active = ?", true).
		Order("level_rank, full_name").Find(&users).Error; err != nil {
		return nil, err
	}
	sections := []string{}
	if err := s.db.Model(&models.User{}).
		Where("is_active = ? AND BTRIM(section) <> ''", true).
		Distinct("UPPER(BTRIM(section))").Order("UPPER(BTRIM(section))").
		Pluck("UPPER(BTRIM(section))", &sections).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"users": users, "sections": sections}, nil
}

func (s *ApprovalSettingService) Mapping(approvalType, sectionCode string) ([]models.ApprovalMapping, error) {
	approvalType, sectionCode, err := normalizeApprovalScope(approvalType, sectionCode, false)
	if err != nil {
		return nil, err
	}
	rows := []models.ApprovalMapping{}
	err = s.db.Where(
		"approval_type = ? AND section_code = ? AND is_active = ?",
		approvalType, sectionCode, true,
	).Order("id").Find(&rows).Error
	return rows, err
}

func (s *ApprovalSettingService) ReplaceMapping(approvalType, sectionCode string, actorID uint64, req CodeMappingRequest) (*models.ApprovalMapping, error) {
	approvalType, sectionCode, err := normalizeApprovalScope(approvalType, sectionCode, false)
	if err != nil {
		return nil, err
	}
	var existing models.ApprovalMapping
	err = s.db.Where("approval_type = ? AND section_code = ?", approvalType, sectionCode).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		existing = models.ApprovalMapping{
			ApprovalType: approvalType,
			SectionCode:  sectionCode,
			Seq1:         strings.ToUpper(strings.TrimSpace(req.Seq1)),
			Seq2:         strings.ToUpper(strings.TrimSpace(req.Seq2)),
			Seq3:         strings.ToUpper(strings.TrimSpace(req.Seq3)),
			Seq4:         strings.ToUpper(strings.TrimSpace(req.Seq4)),
			IsActive:     true,
			CreatedBy:    actorID,
			UpdatedBy:    actorID,
		}
		if err := s.db.Create(&existing).Error; err != nil {
			return nil, err
		}
	} else {
		existing.Seq1 = strings.ToUpper(strings.TrimSpace(req.Seq1))
		existing.Seq2 = strings.ToUpper(strings.TrimSpace(req.Seq2))
		existing.Seq3 = strings.ToUpper(strings.TrimSpace(req.Seq3))
		existing.Seq4 = strings.ToUpper(strings.TrimSpace(req.Seq4))
		existing.UpdatedBy = actorID
		if err := s.db.Save(&existing).Error; err != nil {
			return nil, err
		}
	}
	return &existing, nil
}

func (s *ApprovalSettingService) DeleteMapping(approvalType, sectionCode string) error {
	approvalType, sectionCode, err := normalizeApprovalScope(approvalType, sectionCode, false)
	if err != nil {
		return err
	}
	result := s.db.Where("approval_type = ? AND section_code = ?", approvalType, sectionCode).
		Delete(&models.ApprovalMapping{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *ApprovalSettingService) Delegations(includeInactive bool) ([]models.ApprovalDelegation, error) {
	rows := []models.ApprovalDelegation{}
	query := s.db.Preload("FromUser").Preload("ToUser")
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("is_active DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

func (s *ApprovalSettingService) validateDelegation(req DelegationRequest, excludeID uint64) (DelegationRequest, error) {
	if req.FromUserID == 0 || req.ToUserID == 0 {
		return req, errors.New("user FROM dan TO wajib dipilih")
	}
	if req.FromUserID == req.ToUserID {
		return req, errors.New("user FROM dan TO tidak boleh sama")
	}
	approvalType, sectionCode, err := normalizeApprovalScope(req.ApprovalType, req.SectionCode, true)
	if err != nil {
		return req, err
	}
	req.ApprovalType, req.SectionCode = approvalType, sectionCode
	if req.StartsAt != nil && req.EndsAt != nil && !req.EndsAt.After(*req.StartsAt) {
		return req, errors.New("waktu berakhir harus setelah waktu mulai")
	}

	var fromUser, toUser models.User
	if err := s.db.First(&fromUser, req.FromUserID).Error; err != nil {
		return req, errors.New("user FROM tidak ditemukan")
	}
	if err := s.db.First(&toUser, req.ToUserID).Error; err != nil {
		return req, errors.New("user TO tidak ditemukan")
	}

	// Hybrid Delegation Rules Check:
	// Section Head (level 3): Can delegate to Sec Head (3) or Staff (4)
	// Dept/Div Head (level 1-2): Can delegate to equal level or 1 level below (Level <= FromLevel+1, cannot jump to Staff)
	if fromUser.LevelRank <= 2 {
		if toUser.LevelRank > fromUser.LevelRank+1 {
			return req, fmt.Errorf("pejabat level %d (%s) hanya dapat mendelegasikan ke pejabat setingkat atau 1 tingkat di bawahnya", fromUser.LevelRank, fromUser.FullName)
		}
	}

	var duplicate int64
	query := s.db.Model(&models.ApprovalDelegation{}).Where(
		"from_user_id = ? AND approval_type = ? AND section_code = ? AND is_active = ?",
		req.FromUserID, approvalType, sectionCode, true,
	)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&duplicate).Error; err != nil {
		return req, err
	}
	if duplicate > 0 {
		return req, errors.New("user FROM sudah memiliki delegasi aktif pada scope tersebut")
	}
	return req, nil
}

func (s *ApprovalSettingService) CreateDelegation(actorID uint64, req DelegationRequest) (*models.ApprovalDelegation, error) {
	req, err := s.validateDelegation(req, 0)
	if err != nil {
		return nil, err
	}
	row := &models.ApprovalDelegation{
		FromUserID: req.FromUserID, ToUserID: req.ToUserID,
		ApprovalType: req.ApprovalType, SectionCode: req.SectionCode,
		StartsAt: req.StartsAt, EndsAt: req.EndsAt, IsActive: true,
		CreatedBy: actorID, UpdatedBy: actorID,
	}
	if err := s.db.Create(row).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("FromUser").Preload("ToUser").First(row, row.ID).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ApprovalSettingService) UpdateDelegation(id, actorID uint64, req DelegationRequest) (*models.ApprovalDelegation, error) {
	var row models.ApprovalDelegation
	if err := s.db.First(&row, id).Error; err != nil {
		return nil, err
	}
	req, err := s.validateDelegation(req, id)
	if err != nil {
		return nil, err
	}
	row.FromUserID, row.ToUserID = req.FromUserID, req.ToUserID
	row.ApprovalType, row.SectionCode = req.ApprovalType, req.SectionCode
	row.StartsAt, row.EndsAt, row.UpdatedBy = req.StartsAt, req.EndsAt, actorID
	row.IsActive = true
	if err := s.db.Save(&row).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("FromUser").Preload("ToUser").First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *ApprovalSettingService) DeactivateDelegation(id, actorID uint64) error {
	result := s.db.Model(&models.ApprovalDelegation{}).Where("id = ? AND is_active = ?", id, true).
		Updates(map[string]interface{}{"is_active": false, "updated_by": actorID})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
