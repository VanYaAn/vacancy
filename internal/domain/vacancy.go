package domain

import "time"

type Salary struct {
	From     *int   `json:"from"`
	To       *int   `json:"to"`
	Currency string `json:"currency"`
	Gross    bool   `json:"gross"`
}

type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Employer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Trusted bool   `json:"trusted"`
}

type Vacancy struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Area         Area      `json:"area"`
	Employer     Employer  `json:"employer"`
	Salary       *Salary   `json:"salary"`
	Experience   string    `json:"experience"`    // noExperience, between1And3, ...
	WorkFormats  []string  `json:"work_formats"`  // REMOTE, ON_SITE, HYBRID
	PublishedAt  time.Time `json:"published_at"`
	AlternateURL string    `json:"alternate_url"`
}

type VacancyDetail struct {
	Vacancy
	Description       string   `json:"description"`
	KeySkills         []string `json:"key_skills"`
	ProfessionalRoles []string `json:"professional_roles"`
}

type SearchParams struct {
	Text           string
	Area           string
	Experience     string
	WorkFormat     string
	Salary         int
	Currency       string
	OnlyWithSalary bool
	Page           int
	PerPage        int
}

type SearchResult struct {
	Items   []Vacancy
	Found   int
	Page    int
	Pages   int
	PerPage int
}
