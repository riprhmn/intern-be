package service

import (
	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/repository"
)

type ChangeNoteService struct {
	repo *repository.ChangeNoteRepository
}

func NewChangeNoteService(repo *repository.ChangeNoteRepository) *ChangeNoteService {
	return &ChangeNoteService{repo: repo}
}

func (s *ChangeNoteService) GetAll(search string, userID uint64, role string, page, limit int) ([]models.ChangeNote, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.repo.GetAll(search, userID, role, offset, limit)
}

func (s *ChangeNoteService) GetByID(id uint64) (*models.ChangeNote, error) {
	return s.repo.GetByID(id)
}

func (s *ChangeNoteService) Create(cn *models.ChangeNote) error {
	if cn.Status == "" {
		cn.Status = "SUBMITTED"
	}
	return s.repo.Create(cn)
}

func (s *ChangeNoteService) Update(cn *models.ChangeNote) error {
	return s.repo.Update(cn)
}

func (s *ChangeNoteService) Delete(id uint64) error {
	return s.repo.Delete(id)
}
