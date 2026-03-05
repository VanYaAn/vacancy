package service

import (
	"context"
	"fmt"

	"mcpPrep/internal/domain"
)

type ExportService struct {
	vacancyRepo domain.VacancyRepository
	resumeRepo  domain.ResumeRepository
	exporter    domain.SheetsExporter
}

func NewExportService(
	vacancyRepo domain.VacancyRepository,
	resumeRepo domain.ResumeRepository,
	exporter domain.SheetsExporter,
) *ExportService {
	return &ExportService{
		vacancyRepo: vacancyRepo,
		resumeRepo:  resumeRepo,
		exporter:    exporter,
	}
}

// ExportToSheets загружает все вакансии и резюме из БД и экспортирует в Google Sheets.
func (s *ExportService) ExportToSheets(ctx context.Context) (string, error) {
	vacancies, err := s.vacancyRepo.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("get vacancies: %w", err)
	}

	resumes, err := s.resumeRepo.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("get resumes: %w", err)
	}

	url, err := s.exporter.Export(ctx, vacancies, resumes)
	if err != nil {
		return "", fmt.Errorf("export to sheets: %w", err)
	}

	return url, nil
}
