package service

import (
	"context"
	"fmt"

	"mcpPrep/internal/domain"
)

type ResumeService struct {
	vacancyRepo domain.VacancyRepository
	resumeRepo  domain.ResumeRepository
	generator   domain.ResumeGenerator
	template    domain.ResumeTemplate
}

func NewResumeService(
	vacancyRepo domain.VacancyRepository,
	resumeRepo domain.ResumeRepository,
	generator domain.ResumeGenerator,
	template domain.ResumeTemplate,
) *ResumeService {
	return &ResumeService{
		vacancyRepo: vacancyRepo,
		resumeRepo:  resumeRepo,
		generator:   generator,
		template:    template,
	}
}

// GenerateForVacancy генерирует резюме под одну вакансию и сохраняет в БД.
func (s *ResumeService) GenerateForVacancy(ctx context.Context, vacancyID string) (*domain.Resume, error) {
	vacancy, err := s.vacancyRepo.GetByID(ctx, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("get vacancy %s: %w", vacancyID, err)
	}

	req := domain.ResumeRequest{
		VacancyID:   vacancy.ID,
		Title:       vacancy.Name,
		Description: vacancy.Description,
		KeySkills:   vacancy.KeySkills,
		Experience:  vacancy.Experience,
		WorkFormats: vacancy.WorkFormats,
		Employer:    vacancy.Employer.Name,
		Salary:      vacancy.Salary,
		Template:    s.template,
	}

	resume, err := s.generator.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("generate resume for vacancy %s: %w", vacancyID, err)
	}

	if err := s.resumeRepo.Save(ctx, resume); err != nil {
		return nil, fmt.Errorf("save resume for vacancy %s: %w", vacancyID, err)
	}

	return &resume, nil
}

// GenerateForAll генерирует резюме для всех вакансий без резюме.
func (s *ResumeService) GenerateForAll(ctx context.Context) (int, error) {
	vacancies, err := s.vacancyRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("get all vacancies: %w", err)
	}

	generated := 0
	for _, v := range vacancies {
		// пропускаем если резюме уже есть
		existing, err := s.resumeRepo.GetByVacancyID(ctx, v.ID)
		if err != nil {
			return generated, fmt.Errorf("check existing resume for vacancy %s: %w", v.ID, err)
		}
		if existing != nil {
			continue
		}

		if _, err := s.GenerateForVacancy(ctx, v.ID); err != nil {
			return generated, err
		}
		generated++
	}

	return generated, nil
}
