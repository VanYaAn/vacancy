package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"mcpPrep/internal/domain"
)

type VacancyRepo struct {
	pool *pgxpool.Pool
}

func NewVacancyRepo(pool *pgxpool.Pool) *VacancyRepo {
	return &VacancyRepo{pool: pool}
}

func (r *VacancyRepo) Save(ctx context.Context, v domain.VacancyDetail) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO vacancies (
			id, name, description, employer_id, employer_name, employer_trusted,
			area_id, area_name, salary_from, salary_to, salary_currency, salary_gross,
			experience, work_formats, key_skills, professional_roles,
			published_at, alternate_url
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
		)
		ON CONFLICT (id) DO UPDATE SET
			name               = EXCLUDED.name,
			description        = EXCLUDED.description,
			employer_name      = EXCLUDED.employer_name,
			employer_trusted   = EXCLUDED.employer_trusted,
			salary_from        = EXCLUDED.salary_from,
			salary_to          = EXCLUDED.salary_to,
			salary_currency    = EXCLUDED.salary_currency,
			salary_gross       = EXCLUDED.salary_gross,
			experience         = EXCLUDED.experience,
			work_formats       = EXCLUDED.work_formats,
			key_skills         = EXCLUDED.key_skills,
			professional_roles = EXCLUDED.professional_roles,
			published_at       = EXCLUDED.published_at,
			alternate_url      = EXCLUDED.alternate_url,
			deleted            = false`,
		v.ID, v.Name, v.Description,
		v.Employer.ID, v.Employer.Name, v.Employer.Trusted,
		v.Area.ID, v.Area.Name,
		salaryFrom(v.Salary), salaryTo(v.Salary), salaryCurrency(v.Salary), salaryGross(v.Salary),
		v.Experience, v.WorkFormats, v.KeySkills, v.ProfessionalRoles,
		v.PublishedAt, v.AlternateURL,
	)
	return err
}

func (r *VacancyRepo) SaveBatch(ctx context.Context, vacancies []domain.VacancyDetail) error {
	for _, v := range vacancies {
		if err := r.Save(ctx, v); err != nil {
			return fmt.Errorf("save vacancy %s: %w", v.ID, err)
		}
	}
	return nil
}

func (r *VacancyRepo) GetByID(ctx context.Context, id string) (*domain.VacancyDetail, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description,
		       employer_id, employer_name, employer_trusted,
		       area_id, area_name,
		       salary_from, salary_to, salary_currency, salary_gross,
		       experience, work_formats, key_skills, professional_roles,
		       published_at, alternate_url
		FROM vacancies
		WHERE id = $1 AND deleted = false`, id)

	return scanVacancy(row)
}

func (r *VacancyRepo) GetAll(ctx context.Context) ([]domain.VacancyDetail, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description,
		       employer_id, employer_name, employer_trusted,
		       area_id, area_name,
		       salary_from, salary_to, salary_currency, salary_gross,
		       experience, work_formats, key_skills, professional_roles,
		       published_at, alternate_url
		FROM vacancies
		WHERE deleted = false
		ORDER BY published_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.VacancyDetail
	for rows.Next() {
		v, err := scanVacancy(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *v)
	}
	return result, rows.Err()
}

func (r *VacancyRepo) MarkDeleted(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE vacancies SET deleted = true WHERE id = $1`, id)
	return err
}

// ─── helpers ──────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanVacancy(s scanner) (*domain.VacancyDetail, error) {
	var (
		v                                        domain.VacancyDetail
		salaryFrom, salaryTo                     *int
		salaryCurrency                           *string
		salaryGross                              *bool
	)

	err := s.Scan(
		&v.ID, &v.Name, &v.Description,
		&v.Employer.ID, &v.Employer.Name, &v.Employer.Trusted,
		&v.Area.ID, &v.Area.Name,
		&salaryFrom, &salaryTo, &salaryCurrency, &salaryGross,
		&v.Experience, &v.WorkFormats, &v.KeySkills, &v.ProfessionalRoles,
		&v.PublishedAt, &v.AlternateURL,
	)
	if err != nil {
		return nil, err
	}

	if salaryFrom != nil || salaryTo != nil {
		v.Salary = &domain.Salary{}
		v.Salary.From = salaryFrom
		v.Salary.To = salaryTo
		if salaryCurrency != nil {
			v.Salary.Currency = *salaryCurrency
		}
		if salaryGross != nil {
			v.Salary.Gross = *salaryGross
		}
	}

	return &v, nil
}

func salaryFrom(s *domain.Salary) *int {
	if s == nil {
		return nil
	}
	return s.From
}

func salaryTo(s *domain.Salary) *int {
	if s == nil {
		return nil
	}
	return s.To
}

func salaryCurrency(s *domain.Salary) *string {
	if s == nil {
		return nil
	}
	return &s.Currency
}

func salaryGross(s *domain.Salary) *bool {
	if s == nil {
		return nil
	}
	return &s.Gross
}
