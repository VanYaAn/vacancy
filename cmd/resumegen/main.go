package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"mcpPrep/internal/config"
	mcpclient "mcpPrep/internal/infrastructure/mcp"
	pgrepo "mcpPrep/internal/infrastructure/postgres"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.MCP.GroqAPIKey == "" {
		log.Fatal("GROQ_API_KEY is not set")
	}

	pool, err := pgxpool.New(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	db, err := sql.Open("pgx", cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// берём первую вакансию из БД для теста
	vacancyRepo := pgrepo.NewVacancyRepo(pool)
	vacancies, err := vacancyRepo.GetAll(ctx)
	if err != nil || len(vacancies) == 0 {
		log.Fatal("no vacancies in db")
	}
	v := vacancies[0]
	log.Printf("generating resume for: %s — %s", v.Name, v.Employer.Name)

	// вызов MCP resume_server.py
	client := mcpclient.NewResumeClient(cfg.MCP.PythonBin, cfg.MCP.ResumeServer, cfg.MCP.GroqAPIKey)

	reqData, _ := json.Marshal(map[string]any{
		"vacancy_id":    v.ID,
		"title":         v.Name,
		"description":   v.Description,
		"key_skills":    v.KeySkills,
		"experience":    v.Experience,
		"work_formats":  v.WorkFormats,
		"employer_name": v.Employer.Name,
	})

	result, err := client.GenerateRaw(ctx, string(reqData))
	if err != nil {
		log.Fatalf("generate resume: %v", err)
	}

	var resp struct {
		VacancyID string `json:"vacancy_id"`
		LatexSrc  string `json:"latex_src"`
		PDFBase64 string `json:"pdf_base64"`
	}
	n := 300
	if len(result) < n {
		n = len(result)
	}
	log.Printf("raw response (first 300): %s", result[:n])
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		log.Fatalf("parse response: %v\nfull result: %s", err, result)
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(resp.PDFBase64)
	if err != nil {
		log.Fatalf("decode pdf: %v", err)
	}

	// сохраняем PDF в файл для проверки
	outFile := "resume_test.pdf"
	if err := os.WriteFile(outFile, pdfBytes, 0644); err != nil {
		log.Fatalf("write pdf: %v", err)
	}
	log.Printf("PDF saved to %s (%d bytes)", outFile, len(pdfBytes))

	// сохраняем в postgres
	id := v.ID + "_" + time.Now().Format("20060102150405")
	_, err = pool.Exec(ctx, `
		INSERT INTO resume_pdfs (id, vacancy_id, latex_src, pdf_data)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET latex_src = EXCLUDED.latex_src, pdf_data = EXCLUDED.pdf_data`,
		id, v.ID, resp.LatexSrc, pdfBytes,
	)
	if err != nil {
		log.Fatalf("save to postgres: %v", err)
	}
	log.Printf("saved to postgres: resume_pdfs id=%s", id)
}
