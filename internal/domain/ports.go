package domain

import "context"

// VacancyProvider — получение вакансий из hh.ru (реализуется через infrastructure/hh)
type VacancyProvider interface {
	SearchVacancies(ctx context.Context, params SearchParams) (*SearchResult, error)
	GetVacancy(ctx context.Context, id string) (*VacancyDetail, error)
	SearchAll(ctx context.Context, params SearchParams, maxPages int) ([]Vacancy, error)
}

// ResumeGenerator — генерация резюме под вакансию (реализуется через MCP resume_server.py)
type ResumeGenerator interface {
	Generate(ctx context.Context, req ResumeRequest) (Resume, error)
}

// SheetsExporter — экспорт данных в Google Sheets (реализуется через MCP sheets_server.py)
type SheetsExporter interface {
	Export(ctx context.Context, vacancies []VacancyDetail, resumes []Resume) (sheetURL string, err error)
}
