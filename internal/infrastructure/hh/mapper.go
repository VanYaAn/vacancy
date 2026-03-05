package hh

import (
	"time"

	"mcpPrep/internal/domain"
)

// ─── API response structs (hh.ru shapes) ──────────────────────────────────────

type apiSalary struct {
	From     *int   `json:"from"`
	To       *int   `json:"to"`
	Currency string `json:"currency"`
	Gross    bool   `json:"gross"`
}

type apiArea struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiEmployer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Trusted bool   `json:"trusted"`
}

type apiNamed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiVacancy struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Area         apiArea     `json:"area"`
	Employer     apiEmployer `json:"employer"`
	Salary       *apiSalary  `json:"salary"`
	Experience   *apiNamed   `json:"experience"`
	WorkFormat   []apiNamed  `json:"work_format"`
	PublishedAt  string      `json:"published_at"`
	AlternateURL string      `json:"alternate_url"`
}

type apiVacancyDetail struct {
	apiVacancy
	Description       string     `json:"description"`
	KeySkills         []apiNamed `json:"key_skills"`
	ProfessionalRoles []apiNamed `json:"professional_roles"`
}

type apiSearchResponse struct {
	Items   []apiVacancy `json:"items"`
	Found   int          `json:"found"`
	Page    int          `json:"page"`
	Pages   int          `json:"pages"`
	PerPage int          `json:"per_page"`
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func mapVacancy(a apiVacancy) domain.Vacancy {
	v := domain.Vacancy{
		ID:           a.ID,
		Name:         a.Name,
		Area:         domain.Area{ID: a.Area.ID, Name: a.Area.Name},
		Employer:     domain.Employer{ID: a.Employer.ID, Name: a.Employer.Name, Trusted: a.Employer.Trusted},
		PublishedAt:  parseTime(a.PublishedAt),
		AlternateURL: a.AlternateURL,
	}
	if a.Salary != nil {
		v.Salary = &domain.Salary{
			From:     a.Salary.From,
			To:       a.Salary.To,
			Currency: a.Salary.Currency,
			Gross:    a.Salary.Gross,
		}
	}
	if a.Experience != nil {
		v.Experience = a.Experience.ID
	}
	for _, wf := range a.WorkFormat {
		v.WorkFormats = append(v.WorkFormats, wf.ID)
	}
	return v
}

func mapVacancyDetail(a apiVacancyDetail) *domain.VacancyDetail {
	d := &domain.VacancyDetail{
		Vacancy:     mapVacancy(a.apiVacancy),
		Description: a.Description,
	}
	for _, s := range a.KeySkills {
		d.KeySkills = append(d.KeySkills, s.Name)
	}
	for _, r := range a.ProfessionalRoles {
		d.ProfessionalRoles = append(d.ProfessionalRoles, r.Name)
	}
	return d
}

func mapSearchResult(a apiSearchResponse) *domain.SearchResult {
	result := &domain.SearchResult{
		Found:   a.Found,
		Page:    a.Page,
		Pages:   a.Pages,
		PerPage: a.PerPage,
	}
	for _, item := range a.Items {
		result.Items = append(result.Items, mapVacancy(item))
	}
	return result
}

func parseTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05-0700"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
