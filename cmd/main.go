package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"mcpPrep/internal/config"
	"mcpPrep/internal/domain"
	hhclient "mcpPrep/internal/infrastructure/hh"
	mcpclient "mcpPrep/internal/infrastructure/mcp"
	pgrepo "mcpPrep/internal/infrastructure/postgres"
	"mcpPrep/internal/service"
)

func main() {
	ctx := context.Background()

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	pool, err := pgxpool.New(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}
	log.Println("postgres: connected")

	// ── Migrations ────────────────────────────────────────────────────────────
	db, err := sql.Open("pgx", cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("open db for migrations: %v", err)
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	log.Println("migrations: up to date")

	// ── Infrastructure ────────────────────────────────────────────────────────
	hhClient := hhclient.NewClient(cfg.HH.Token)
	vacancyRepo := pgrepo.NewVacancyRepo(pool)
	resumeRepo := pgrepo.NewResumeRepo(pool)
	resumeGenerator := mcpclient.NewResumeClient(cfg.MCP.PythonBin, cfg.MCP.ResumeServer)
	sheetsExporter := mcpclient.NewSheetsClient(cfg.MCP.PythonBin, cfg.MCP.SheetsServer, cfg.Google.CredentialsFile, cfg.Google.SheetID)

	// ── Template ──────────────────────────────────────────────────────────────
	// TODO: загружать из файла
	template := domain.ResumeTemplate{
		Blocks: []domain.TemplateBlock{
			{Name: "summary", Content: ""},
			{Name: "experience", Content: ""},
			{Name: "skills", Content: ""},
		},
	}

	// ── Services ──────────────────────────────────────────────────────────────
	vacancySvc := service.NewVacancyService(hhClient, vacancyRepo)
	resumeSvc := service.NewResumeService(vacancyRepo, resumeRepo, resumeGenerator, template)
	exportSvc := service.NewExportService(vacancyRepo, resumeRepo, sheetsExporter)

	// ── Pipeline ──────────────────────────────────────────────────────────────

	// Шаг 1: поиск и сохранение вакансий
	log.Println("step 1: searching vacancies...")
	params := domain.SearchParams{
		Text:           cfg.Search.Text,
		Area:           cfg.Search.Area,
		Experience:     cfg.Search.Experience,
		WorkFormat:     cfg.Search.WorkFormat,
		Salary:         cfg.Search.Salary,
		OnlyWithSalary: cfg.Search.OnlyWithSalary,
		PerPage:        cfg.HH.DefaultPerPage,
	}
	saved, err := vacancySvc.SearchAndSave(ctx, params, cfg.HH.MaxPages)
	if err != nil {
		log.Fatalf("search and save: %v", err)
	}
	log.Printf("step 1: saved %d vacancies", saved)

	// Шаг 2: генерация резюме под каждую вакансию
	log.Println("step 2: generating resumes...")
	generated, err := resumeSvc.GenerateForAll(ctx)
	if err != nil {
		log.Fatalf("generate resumes: %v", err)
	}
	log.Printf("step 2: generated %d resumes", generated)

	// Шаг 3: экспорт в Google Sheets
	log.Println("step 3: exporting to google sheets...")
	sheetURL, err := exportSvc.ExportToSheets(ctx)
	if err != nil {
		log.Fatalf("export to sheets: %v", err)
	}
	log.Printf("step 3: done — %s", sheetURL)
}
