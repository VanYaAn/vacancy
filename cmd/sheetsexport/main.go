package main

import (
	"context"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"

	"mcpPrep/internal/config"
	mcpclient "mcpPrep/internal/infrastructure/mcp"
	pgrepo "mcpPrep/internal/infrastructure/postgres"
	"mcpPrep/internal/service"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err := pgxpool.New(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}

	exportSvc := service.NewExportService(
		pgrepo.NewVacancyRepo(pool),
		pgrepo.NewResumeRepo(pool),
		mcpclient.NewSheetsClient(cfg.MCP.PythonBin, cfg.MCP.SheetsServer, cfg.Google.CredentialsFile, cfg.Google.SheetID),
	)

	log.Println("exporting to google sheets...")
	sheetURL, err := exportSvc.ExportToSheets(ctx)
	if err != nil {
		log.Fatalf("export: %v", err)
	}
	log.Printf("done: %s", sheetURL)
}
