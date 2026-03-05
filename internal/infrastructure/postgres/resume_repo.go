package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"mcpPrep/internal/domain"
)

type ResumeRepo struct {
	pool *pgxpool.Pool
}

func NewResumeRepo(pool *pgxpool.Pool) *ResumeRepo {
	return &ResumeRepo{pool: pool}
}

func (r *ResumeRepo) Save(ctx context.Context, resume domain.Resume) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO resumes (id, vacancy_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			content    = EXCLUDED.content,
			created_at = EXCLUDED.created_at`,
		resume.ID, resume.VacancyID, resume.Content, resume.CreatedAt,
	)
	return err
}

func (r *ResumeRepo) GetByVacancyID(ctx context.Context, vacancyID string) (*domain.Resume, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, vacancy_id, content, created_at
		FROM resumes
		WHERE vacancy_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, vacancyID)

	var resume domain.Resume
	if err := row.Scan(&resume.ID, &resume.VacancyID, &resume.Content, &resume.CreatedAt); err != nil {
		return nil, err
	}
	return &resume, nil
}

func (r *ResumeRepo) GetAll(ctx context.Context) ([]domain.Resume, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, vacancy_id, content, created_at
		FROM resumes
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Resume
	for rows.Next() {
		var resume domain.Resume
		if err := rows.Scan(&resume.ID, &resume.VacancyID, &resume.Content, &resume.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, resume)
	}
	return result, rows.Err()
}
