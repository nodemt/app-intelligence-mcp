package main

import (
	"app-intelligence-mcp/internal/constants"
	"context"
	"fmt"
	"os"
	"strconv"

	"app-intelligence-mcp/internal/mixpanel"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SavedProductReportInput struct {
	BookmarkID int `json:"bookmark_id" jsonschema:"ID of a saved Mixpanel report"`
}

type MixpanelHandler struct {
	client    *mixpanel.Client
	projectID int
}

func (h *MixpanelHandler) GetSavedProductReport(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input SavedProductReportInput,
) (*mcp.CallToolResult, mixpanel.SavedReport, error) {
	if input.BookmarkID <= 0 {
		return nil, mixpanel.SavedReport{}, fmt.Errorf("bookmark_id must be positive")
	}

	report, err := h.client.GetSavedReport(ctx, h.projectID, input.BookmarkID)
	if err != nil {
		return nil, mixpanel.SavedReport{}, err
	}

	return nil, report, nil
}

func registerMixpanelTool(server *mcp.Server) error {
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

	handler := &MixpanelHandler{
		client:    mixpanel.NewClient(endpoint, username, secret),
		projectID: projectID,
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        constants.SavedProductReportToolName,
		Description: "Read a saved Mixpanel report by bookmark ID.",
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"computed_at": map[string]any{"type": "string"},
				"date_range":  map[string]any{"type": "object"},
				"headers":     map[string]any{"type": []string{"array", "null"}},
				"series":      map[string]any{},
			},
		},
	}, handler.GetSavedProductReport)

	return nil
}
