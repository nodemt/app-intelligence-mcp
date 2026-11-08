package mixpanel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPReport struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Results json.RawMessage `json:"results"`
}

func DecodeMCPReport(content []mcp.Content, bookmarkID int) (MCPReport, error) {
	for _, item := range content {
		text, ok := item.(*mcp.TextContent)
		if !ok {
			continue
		}

		var report MCPReport
		if err := json.Unmarshal([]byte(text.Text), &report); err != nil {
			continue
		}

		if report.ID == bookmarkID {
			return report, nil
		}
	}

	return MCPReport{}, fmt.Errorf(
		"Mixpanel MCP response did not contain report %d", bookmarkID,
	)
}

func GetMCPReport(ctx context.Context, session *mcp.ClientSession, projectID, bookmarkID int) (MCPReport, error) {
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "Get-Report",
		Arguments: map[string]any{
			"project_id":   projectID,
			"bookmark_id":  bookmarkID,
			"skip_results": false,
		},
	})
	if err != nil {
		return MCPReport{}, fmt.Errorf("call Mixpanel Get-Report: %w", err)
	}

	if result.IsError {
		return MCPReport{}, fmt.Errorf("Mixpanel Get-Report failed for report %d", bookmarkID)
	}

	return DecodeMCPReport(result.Content, bookmarkID)
}
