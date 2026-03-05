package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcpPrep/internal/domain"
)

// SheetsClient вызывает sheets_server.py через MCP.
// Реализует domain.SheetsExporter.
type SheetsClient struct {
	pythonBin       string
	serverPath      string
	credentialsFile string
	sheetID         string
}

func NewSheetsClient(pythonBin, serverPath, credentialsFile, sheetID string) *SheetsClient {
	return &SheetsClient{
		pythonBin:       pythonBin,
		serverPath:      serverPath,
		credentialsFile: credentialsFile,
		sheetID:         sheetID,
	}
}

type exportPayload struct {
	Vacancies []domain.VacancyDetail `json:"vacancies"`
	Resumes   []domain.Resume        `json:"resumes"`
}

func (c *SheetsClient) Export(ctx context.Context, vacancies []domain.VacancyDetail, resumes []domain.Resume) (string, error) {
	session, close, err := c.connect(ctx)
	if err != nil {
		return "", fmt.Errorf("connect to sheets server: %w", err)
	}
	defer close()

	payload, err := json.Marshal(exportPayload{
		Vacancies: vacancies,
		Resumes:   resumes,
	})
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "export_to_sheets",
		Arguments: map[string]any{"data": string(payload)},
	})
	if err != nil {
		return "", fmt.Errorf("call export_to_sheets: %w", err)
	}

	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return "", fmt.Errorf("unexpected content type from sheets server")
	}

	return content.Text, nil
}

func (c *SheetsClient) connect(ctx context.Context) (*mcp.ClientSession, func(), error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "sheets-client",
		Version: "v1.0.0",
	}, nil)

	cmd := exec.CommandContext(ctx, c.pythonBin, c.serverPath)
	cmd.Env = append(cmd.Environ(),
		"GOOGLE_CREDENTIALS_FILE="+c.credentialsFile,
		"GOOGLE_SHEET_ID="+c.sheetID,
	)
	transport := &mcp.CommandTransport{Command: cmd}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, nil, err
	}

	return session, func() { session.Close() }, nil
}
