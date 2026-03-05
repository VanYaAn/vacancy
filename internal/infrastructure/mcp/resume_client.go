package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcpPrep/internal/domain"
)

// ResumeClient вызывает resume_server.py через MCP.
// Реализует domain.ResumeGenerator.
type ResumeClient struct {
	pythonBin  string
	serverPath string
	groqAPIKey string
}

func NewResumeClient(pythonBin, serverPath, groqAPIKey string) *ResumeClient {
	return &ResumeClient{
		pythonBin:  pythonBin,
		serverPath: serverPath,
		groqAPIKey: groqAPIKey,
	}
}

// GenerateRaw вызывает generate_resume и возвращает сырой JSON ответ от сервера.
func (c *ResumeClient) GenerateRaw(ctx context.Context, data string) (string, error) {
	session, close, err := c.connect(ctx)
	if err != nil {
		return "", fmt.Errorf("connect to resume server: %w", err)
	}
	defer close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "generate_resume",
		Arguments: map[string]any{"data": data},
	})
	if err != nil {
		return "", fmt.Errorf("call generate_resume: %w", err)
	}

	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return "", fmt.Errorf("unexpected content type from resume server")
	}
	return content.Text, nil
}

func (c *ResumeClient) Generate(ctx context.Context, req domain.ResumeRequest) (domain.Resume, error) {
	session, close, err := c.connect(ctx)
	if err != nil {
		return domain.Resume{}, fmt.Errorf("connect to resume server: %w", err)
	}
	defer close()

	args, err := json.Marshal(req)
	if err != nil {
		return domain.Resume{}, fmt.Errorf("marshal request: %w", err)
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "generate_resume",
		Arguments: map[string]any{"data": string(args)},
	})
	if err != nil {
		return domain.Resume{}, fmt.Errorf("call generate_resume: %w", err)
	}

	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return domain.Resume{}, fmt.Errorf("unexpected content type from resume server")
	}

	return domain.Resume{
		VacancyID: req.VacancyID,
		Content:   content.Text,
		CreatedAt: time.Now(),
	}, nil
}

func (c *ResumeClient) connect(ctx context.Context) (*mcp.ClientSession, func(), error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "resume-client",
		Version: "v1.0.0",
	}, nil)

	cmd := exec.CommandContext(ctx, c.pythonBin, c.serverPath)
	cmd.Env = append(cmd.Environ(), "GROQ_API_KEY="+c.groqAPIKey)
	transport := &mcp.CommandTransport{Command: cmd}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, nil, err
	}

	return session, func() { session.Close() }, nil
}
