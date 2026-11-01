package astro

import (
	"app-intelligence-mcp/internal/constants"
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Client struct {
	session *mcp.ClientSession
}

func Connect(ctx context.Context, endpoint string) (*Client, error) {
	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "app-intelligence-probe",
			Version: "0.1.0",
		},
		nil,
	)

	transport := &mcp.StreamableClientTransport{
		Endpoint: endpoint,
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	return &Client{session: session}, nil
}

func (c *Client) Close() error {
	return c.session.Close()
}

func (c *Client) ListTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	return c.session.ListTools(ctx, nil)
}

func (c *Client) GetKeywordPerformance(ctx context.Context, keyword string, store string, platform string, appID string) ([]Ranking, error) {
	arguments := keywordPerformanceArguments(keyword, store, platform, appID)

	rankingResult, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      constants.AstroKeywordPerformanceToolName,
		Arguments: arguments,
	})
	if err != nil {
		return nil, fmt.Errorf("call Astro get_keyword_performance: %w", err)
	}

	if rankingResult.IsError {
		for _, content := range rankingResult.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				return nil, fmt.Errorf("astro get_keyword_performance: %s", textContent.Text)
			}
		}
		return nil, fmt.Errorf("astro get_keyword_performance failed")
	}

	for _, content := range rankingResult.Content {
		textContent, ok := content.(*mcp.TextContent)
		if !ok {
			continue
		}

		return decodeRankings(textContent.Text)

	}
	return nil, fmt.Errorf("Astro keyword performance response contains no text content")
}

func keywordPerformanceArguments(
	keyword, store, platform, appID string,
) map[string]any {
	arguments := map[string]any{
		"keyword":        keyword,
		"store":          store,
		"platform":       platform,
		"includeHistory": true,
		"period":         constants.AstroDefaultPeriod,
	}

	if appID != "" {
		arguments["appId"] = appID
	}

	return arguments
}
