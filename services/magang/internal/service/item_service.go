package service

import (
	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/repository"
)

type ItemService struct {
	repo *repository.ItemRepository
}

func NewItemService(repo *repository.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) GetAll(search string, page, limit int) ([]models.Item, int64, error) {
	offset := (page - 1) * limit
	return s.repo.GetAll(search, offset, limit)
}

func (s *ItemService) GetByID(id uint64) (*models.Item, error) {
	return s.repo.GetByID(id)
}

func (s *ItemService) Create(item *models.Item) error {
	return s.repo.Create(item)
}

func (s *ItemService) Update(item *models.Item) error {
	return s.repo.Update(item)
}

func (s *ItemService) Delete(id uint64) error {
	return s.repo.Delete(id)
}
