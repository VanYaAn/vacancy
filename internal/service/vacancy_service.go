package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"mcpPrep/internal/domain"
	hhinfra "mcpPrep/internal/infrastructure/hh"
)

type VacancyService struct {
	hh   domain.VacancyProvider
	repo domain.VacancyRepository
}

func NewVacancyService(hh domain.VacancyProvider, repo domain.VacancyRepository) *VacancyService {
	return &VacancyService{hh: hh, repo: repo}
}

// SearchAndSave ищет вакансии на hh.ru, загружает детали каждой и сохраняет в БД.
func (s *VacancyService) SearchAndSave(ctx context.Context, params domain.SearchParams, maxPages int) (int, error) {
	vacancies, err := s.hh.SearchAll(ctx, params, maxPages)
	if err != nil {
		return 0, fmt.Errorf("search vacancies: %w", err)
	}

	saved := 0
	for _, v := range vacancies {
		detail, err := s.hh.GetVacancy(ctx, v.ID)
		if err != nil {
			if errors.Is(err, hhinfra.ErrNotFound) {
				log.Printf("vacancy %s not found, skipping", v.ID)
				continue
			}
			return saved, fmt.Errorf("get vacancy %s: %w", v.ID, err)
		}

		if err := s.repo.Save(ctx, *detail); err != nil {
			return saved, fmt.Errorf("save vacancy %s: %w", v.ID, err)
		}
		saved++
	}

	return saved, nil
}

// GetAll возвращает все сохранённые вакансии из БД.
func (s *VacancyService) GetAll(ctx context.Context) ([]domain.VacancyDetail, error) {
	return s.repo.GetAll(ctx)
}
