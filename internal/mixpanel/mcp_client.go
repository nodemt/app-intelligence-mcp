package mixpanel

import (
	"app-intelligence-mcp/internal/constants"
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ConnectMCP(ctx context.Context, endpoint string, oauthHandler auth.OAuthHandler) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    constants.AppName,
		Version: constants.AppVersion,
	}, nil)

	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:     endpoint,
		OAuthHandler: oauthHandler,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to Mixpanel MCP: %w", err)

	}

	return session, nil
}
