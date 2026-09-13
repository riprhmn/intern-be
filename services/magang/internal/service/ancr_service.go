package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/repository"
)

type ANCRService struct {
	repo *repository.ANCRRepository
}

func NewANCRService(repo *repository.ANCRRepository) *ANCRService {
	return &ANCRService{repo: repo}
}

type ANCRCreateRequest struct {
	TypeANCR          string   `json:"typeAncr"`
	AuditNo           string   `json:"auditNo"`
	Initiator         string   `json:"initiator"`
	DepartmentInit    string   `json:"departmentInitiator"`
	Auditor           string   `json:"auditor"`
	Auditee           string   `json:"auditee"`
	DepartmentAuditee string   `json:"departmentAuditee"`
	RelatedTo         []string `json:"relatedTo"`
	FindingCriteria   string   `json:"findingCriteria"`
	Problem           string   `json:"problem"`
	ObjectiveEvidence string   `json:"objectiveEvidence"`
	Location          string   `json:"location"`
	Reference         string   `json:"reference"`
}

type ANCRUpdateRequest struct {
	Process      string `json:"process"`
	Stage        int    `json:"stage"`
	WorkflowData string `json:"workflowData"`
}

func (s *ANCRService) GetAll(search string, page, limit int) ([]models.ANCRRequest, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.repo.GetAll(search, offset, limit)
}

func (s *ANCRService) GetByID(id uint64) (*models.ANCRRequest, error) {
	return s.repo.GetByID(id)
}

func (s *ANCRService) Create(userID uint64, req ANCRCreateRequest) (*models.ANCRRequest, error) {
	if strings.TrimSpace(req.TypeANCR) == "" {
		return nil, errors.New("typeAncr wajib diisi")
	}

	relatedJSON, _ := json.Marshal(req.RelatedTo)

	regNo := fmt.Sprintf("REG-%d", 100000+rand.Intn(900000))

	row := &models.ANCRRequest{
		UserID:             userID,
		RegistrationNumber: regNo,
		TypeANCR:           req.TypeANCR,
		AuditNo:            req.AuditNo,
		Initiator:          req.Initiator,
		DepartmentInit:     req.DepartmentInit,
		Auditor:            req.Auditor,
		Auditee:            req.Auditee,
		DepartmentAuditee:  req.DepartmentAuditee,
		RelatedTo:          string(relatedJSON),
		FindingCriteria:    req.FindingCriteria,
		Problem:            req.Problem,
		ObjectiveEvidence:  req.ObjectiveEvidence,
		Location:           req.Location,
		Reference:          req.Reference,
		Process:            "Outstanding",
		Stage:              1,
		SubmittedAt:        time.Now(),
	}
	if err := s.repo.Create(row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ANCRService) Update(id uint64, req ANCRUpdateRequest) (*models.ANCRRequest, error) {
	row, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("ANCR tidak ditemukan")
	}
	if req.Process != "" {
		row.Process = req.Process
	}
	if req.Stage > 0 {
		row.Stage = req.Stage
	}
	if req.WorkflowData != "" {
		row.WorkflowData = req.WorkflowData
	}
	if err := s.repo.Update(row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ANCRService) Delete(id uint64) error {
	return s.repo.Delete(id)
}
