package domain

import "time"

type Resume struct {
	ID        string    `json:"id"`
	VacancyID string    `json:"vacancy_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ResumeTemplate struct {
	Blocks []TemplateBlock `json:"blocks"`
}

type TemplateBlock struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ResumeRequest — данные, которые Go передаёт в MCP resume_server.py
type ResumeRequest struct {
	VacancyID   string         `json:"vacancy_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	KeySkills   []string       `json:"key_skills"`
	Experience  string         `json:"experience"`
	WorkFormats []string       `json:"work_formats"`
	Employer    string         `json:"employer"`
	Salary      *Salary        `json:"salary,omitempty"`
	Template    ResumeTemplate `json:"template"`
}
