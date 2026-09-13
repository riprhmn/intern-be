package service

import (
	"errors"
	"fmt"
	"sort"
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

type MappingEntryInput struct {
	Sequence int    `json:"sequence"`
	UserID   uint64 `json:"user_id"`
	CodeName string `json:"code_name"`
}

type ReplaceMappingRequest struct {
	Entries []MappingEntryInput `json:"entries"`
	Seq1    string              `json:"seq_1"`
	Seq2    string              `json:"seq_2"`
	Seq3    string              `json:"seq_3"`
	Seq4    string              `json:"seq_4"`
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
	if err := s.db.Where("is_active = ? AND role = ? AND BTRIM(code_name) <> ''", true, "approval").
		Order("code_name, full_name").Find(&users).Error; err != nil {
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
	approvalType, sectionCode, err := normalizeApprovalScope(approvalType, sectionCode, true)
	if err != nil {
		return nil, err
	}
	rows := []models.ApprovalMapping{}
	query := s.db.Preload("User").Where("approval_type = ? AND is_active = ?", approvalType, true)
	if sectionCode != "" && sectionCode != "ALL" {
		query = query.Where("section_code = ?", sectionCode)
	}
	err = query.Order("section_code, sequence, id").Find(&rows).Error
	return rows, err
}

func validateMappingEntries(approvalType string, entries []MappingEntryInput) error {
	if len(entries) == 0 {
		return errors.New("master mapping tidak boleh kosong")
	}
	sequences := make(map[int]int)
	users := make(map[uint64]struct{})
	for _, entry := range entries {
		if entry.Sequence < 1 {
			return errors.New("setiap mapping wajib memiliki sequence positif")
		}
		if entry.UserID == 0 {
			continue
		}
		sequences[entry.Sequence]++
		users[entry.UserID] = struct{}{}
	}
	ordered := make([]int, 0, len(sequences))
	for sequence := range sequences {
		ordered = append(ordered, sequence)
	}
	sort.Ints(ordered)
	for index, sequence := range ordered {
		if sequence != index+1 {
			return errors.New("sequence harus berurutan mulai dari 1 tanpa lompatan")
		}
	}
	if approvalType == "CN" {
		for _, count := range sequences {
			if count > 1 {
				return errors.New("CN hanya boleh memiliki satu approver pada setiap sequence")
			}
		}
	}
	return nil
}

func (s *ApprovalSettingService) validateApprovalUsers(entries []MappingEntryInput) error {
	for _, entry := range entries {
		var count int64
		if err := s.db.Model(&models.User{}).Where(
			"id = ? AND is_active = ? AND BTRIM(COALESCE(code_name, '')) <> ''",
			entry.UserID, true,
		).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("user %d tidak ditemukan atau belum memiliki CODE_NAME", entry.UserID)
		}
	}
	return nil
}

func (s *ApprovalSettingService) ReplaceMapping(approvalType, sectionCode string, actorID uint64, req ReplaceMappingRequest) ([]models.ApprovalMapping, error) {
	approvalType, sectionCode, err := normalizeApprovalScope(approvalType, sectionCode, false)
	if err != nil {
		return nil, err
	}

	var finalEntries []MappingEntryInput

	if len(req.Entries) > 0 {
		for _, entry := range req.Entries {
			code := strings.TrimSpace(entry.CodeName)
			if strings.Contains(code, "—") {
				code = strings.TrimSpace(strings.Split(code, "—")[0])
			} else if strings.Contains(code, " - ") {
				code = strings.TrimSpace(strings.Split(code, " - ")[0])
			}
			code = strings.ToUpper(code)

			if code != "" {
				var u models.User
				if err := s.db.Where("is_active = ? AND (LOWER(BTRIM(code_name)) = LOWER(?) OR LOWER(BTRIM(username)) = LOWER(?))", true, code, code).First(&u).Error; err == nil {
					finalEntries = append(finalEntries, MappingEntryInput{
						Sequence: entry.Sequence,
						UserID:   u.ID,
						CodeName: u.CodeName,
					})
				}
			}
		}
	}

	if len(finalEntries) == 0 {
		codes := []string{req.Seq1, req.Seq2, req.Seq3, req.Seq4}
		for i, code := range codes {
			code = strings.TrimSpace(code)
			if strings.Contains(code, "—") {
				code = strings.TrimSpace(strings.Split(code, "—")[0])
			} else if strings.Contains(code, " - ") {
				code = strings.TrimSpace(strings.Split(code, " - ")[0])
			}
			code = strings.ToUpper(code)

			if code != "" {
				var u models.User
				if err := s.db.Where("is_active = ? AND (LOWER(BTRIM(code_name)) = LOWER(?) OR LOWER(BTRIM(username)) = LOWER(?))", true, code, code).First(&u).Error; err == nil {
					finalEntries = append(finalEntries, MappingEntryInput{
						Sequence: i + 1,
						UserID:   u.ID,
						CodeName: u.CodeName,
					})
				}
			}
		}
	}

	req.Entries = finalEntries

	if err := validateMappingEntries(approvalType, req.Entries); err != nil {
		return nil, err
	}
	if err := s.validateApprovalUsers(req.Entries); err != nil {
		return nil, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("approval_type = ? AND section_code = ?", approvalType, sectionCode).
			Delete(&models.ApprovalMapping{}).Error; err != nil {
			return err
		}
		if len(req.Entries) == 0 {
			return nil
		}
		for _, entry := range req.Entries {
			now := time.Now()
			err := tx.Exec(`
				INSERT INTO magang.approval_mappings 
				(approval_type, section_code, sequence, user_id, is_active, created_by, updated_by, created_at, updated_at) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, approvalType, sectionCode, entry.Sequence, entry.UserID, true, actorID, actorID, now, now).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Mapping(approvalType, sectionCode)
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
