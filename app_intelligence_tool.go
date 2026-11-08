package main

import (
	"context"
	"fmt"

	"app-intelligence-mcp/internal/constants"
	"app-intelligence-mcp/internal/mixpanel"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AppIntelligenceInput struct {
	KeywordRankingsInput
	SavedProductReportInput
}

type AppIntelligenceOutput struct {
	Astro    KeywordRankingsOutput `json:"astro"`
	Mixpanel mixpanel.SavedReport  `json:"mixpanel"`
}

type appIntelligenceHandler struct {
	astro    *KeywordHandler
	mixpanel *MixpanelHandler
}

func addAppIntelligenceTool(server *mcp.Server, rankings *KeywordHandler, reports *MixpanelHandler) {
	handler := &appIntelligenceHandler{astro: rankings, mixpanel: reports}
	mcp.AddTool(server, &mcp.Tool{
		Name: constants.AppIntelligenceToolName,
		Description: "Read Astro keyword rankings and a saved Mixpanel report in one call. " +
			"Choose a bookmark for the same app and check each source's dates.",
		OutputSchema: map[string]any{
			"type":     "object",
			"required": []string{"astro", "mixpanel"},
			"properties": map[string]any{
				"astro": map[string]any{
					"type":     "object",
					"required": []string{"rankings"},
					"properties": map[string]any{
						"rankings": map[string]any{"type": "array"},
					},
				},
				"mixpanel": savedReportOutputSchema(),
			},
		},
	}, handler.getAppIntelligence)
}

func (h *appIntelligenceHandler) getAppIntelligence(ctx context.Context, req *mcp.CallToolRequest, input AppIntelligenceInput) (*mcp.CallToolResult, AppIntelligenceOutput, error) {
	if input.BookmarkID <= 0 {
		return nil, AppIntelligenceOutput{}, fmt.Errorf("bookmark_id must be positive")
	}

	_, rankings, err := h.astro.GetKeywordRankings(ctx, req, input.KeywordRankingsInput)
	if err != nil {
		return nil, AppIntelligenceOutput{}, fmt.Errorf("Astro: %w", err)
	}

	_, report, err := h.mixpanel.GetSavedProductReport(ctx, req, input.SavedProductReportInput)
	if err != nil {
		return nil, AppIntelligenceOutput{}, fmt.Errorf("Mixpanel: %w", err)
	}

	return nil, AppIntelligenceOutput{Astro: rankings, Mixpanel: report}, nil
}
