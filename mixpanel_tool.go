package main

import (
	"app-intelligence-mcp/internal/constants"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"app-intelligence-mcp/internal/mixpanel"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SavedProductReportInput struct {
	BookmarkID int `json:"bookmark_id" jsonschema:"ID of a saved Mixpanel report"`
}

type SavedReportProvider interface {
	GetSavedReport(ctx context.Context, projectID, bookmarkID int) (mixpanel.SavedReport, error)
}

type MixpanelHandler struct {
	client    SavedReportProvider
	projectID int
}

func (h *MixpanelHandler) GetSavedProductReport(ctx context.Context, _ *mcp.CallToolRequest, input SavedProductReportInput) (*mcp.CallToolResult, mixpanel.SavedReport, error) {
	if input.BookmarkID <= 0 {
		return nil, mixpanel.SavedReport{}, fmt.Errorf("bookmark_id must be positive")
	}

	report, err := h.client.GetSavedReport(ctx, h.projectID, input.BookmarkID)
	if err != nil {
		return nil, mixpanel.SavedReport{}, err
	}

	return nil, report, nil
}

func registerMixpanelTool(server *mcp.Server, rankings RankingProvider) error {
	endpoint := os.Getenv("MIXPANEL_QUERY_ENDPOINT")
	username := os.Getenv("MIXPANEL_SERVICE_ACCOUNT_USERNAME")
	secret := os.Getenv("MIXPANEL_SERVICE_ACCOUNT_SECRET")
	project := os.Getenv("MIXPANEL_PROJECT_ID")

	if endpoint == "" && username == "" && secret == "" && project == "" {
		return nil
	}
	if endpoint == "" || username == "" || secret == "" || project == "" {
		return fmt.Errorf("Mixpanel configuration is incomplete")
	}

	projectID, err := strconv.Atoi(project)
	if err != nil || projectID <= 0 {
		return fmt.Errorf("MIXPANEL_PROJECT_ID must be a positive integer")
	}

	addMixpanelTools(server, rankings, mixpanel.NewClient(endpoint, username, secret), projectID)
	return nil
}

func addMixpanelTools(server *mcp.Server, rankings RankingProvider, provider SavedReportProvider, projectID int) {
	handler := &MixpanelHandler{
		client:    provider,
		projectID: projectID,
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:         constants.SavedProductReportToolName,
		Description:  "Read a saved Mixpanel report by bookmark ID.",
		OutputSchema: savedReportOutputSchema(),
	}, handler.GetSavedProductReport)

	addAppIntelligenceTool(server, &KeywordHandler{provider: rankings}, handler)
}

func savedReportOutputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"computed_at": map[string]any{"type": "string"},
			"date_range":  map[string]any{"type": "object"},
			"headers":     map[string]any{"type": []string{"array", "null"}},
			"series":      map[string]any{},
			"results":     map[string]any{},
		},
	}
}

func registerMixpanelOAuthTool(ctx context.Context, server *mcp.Server, rankings RankingProvider, endpoint string) (func(), error) {
	projectID, err := strconv.Atoi(os.Getenv("MIXPANEL_PROJECT_ID"))
	if err != nil || projectID <= 0 {
		return nil, fmt.Errorf("MIXPANEL_PROJECT_ID must be a positive integer")
	}

	oauthHandler, closeOAuth, err := mixpanel.NewLoopbackOAuth()
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	session, err := mixpanel.ConnectMCP(connectCtx, endpoint, oauthHandler)
	if err != nil {
		closeOAuth()
		return nil, err
	}

	addMixpanelTools(server, rankings, &mixpanel.MCPReportProvider{Session: session}, projectID)

	return func() {
		_ = session.Close()
		closeOAuth()
	}, nil
}
