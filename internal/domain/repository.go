package domain

import "context"

type VacancyRepository interface {
	Save(ctx context.Context, vacancy VacancyDetail) error
	SaveBatch(ctx context.Context, vacancies []VacancyDetail) error
	GetByID(ctx context.Context, id string) (*VacancyDetail, error)
	GetAll(ctx context.Context) ([]VacancyDetail, error)
	MarkDeleted(ctx context.Context, id string) error
}

type ResumeRepository interface {
	Save(ctx context.Context, resume Resume) error
	GetByVacancyID(ctx context.Context, vacancyID string) (*Resume, error)
	GetAll(ctx context.Context) ([]Resume, error)
}
