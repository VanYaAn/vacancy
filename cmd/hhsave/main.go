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
	pgrepo "mcpPrep/internal/infrastructure/postgres"
	"mcpPrep/internal/service"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// ── PostgreSQL ─────────────────────────────────────────────────────────────
	pool, err := pgxpool.New(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}

	// ── Миграции ───────────────────────────────────────────────────────────────
	db, err := sql.Open("pgx", cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("open db for migrations: %v", err)
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	log.Println("migrations: ok")

	// ── Search & Save ──────────────────────────────────────────────────────────
	vacancySvc := service.NewVacancyService(
		hhclient.NewClient(cfg.HH.Token),
		pgrepo.NewVacancyRepo(pool),
	)

	params := domain.SearchParams{
		Text:           cfg.Search.Text,
		Area:           cfg.Search.Area,
		Experience:     cfg.Search.Experience,
		WorkFormat:     cfg.Search.WorkFormat,
		Salary:         cfg.Search.Salary,
		OnlyWithSalary: cfg.Search.OnlyWithSalary,
		PerPage:        cfg.HH.DefaultPerPage,
	}

	log.Printf("searching: text=%q area=%s max_pages=%d", params.Text, params.Area, cfg.HH.MaxPages)

	saved, err := vacancySvc.SearchAndSave(ctx, params, cfg.HH.MaxPages)
	if err != nil {
		log.Fatalf("search and save: %v", err)
	}

	log.Printf("done: saved %d vacancies to postgres", saved)
}
