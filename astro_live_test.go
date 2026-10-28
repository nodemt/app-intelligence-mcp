package main

import (
	"app-intelligence-mcp/internal/astro"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestAstroConnectionLive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	provider, err := astro.Connect(ctx, "http://127.0.0.1:8089/mcp")
	if err != nil {
		t.Fatalf("connect Astro: %v", err)
	}
	defer provider.Close()

	server := newServer(provider)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "live-test-client",
			Version: "0.1.0",
		},
		nil,
	)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect test client: %v", err)
	}
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_keyword_rankings",
		Arguments: map[string]any{
			"keyword":  "dibujo digital",
			"store":    "es",
			"platform": "iphone",
		},
	})
	if err != nil {
		t.Fatalf("call keyword rankings tool: %v", err)
	}

	if result.IsError {
		for _, content := range result.Content {
			if text, ok := content.(*mcp.TextContent); ok {
				t.Logf("tool error: %s", text.Text)
			}
		}
		t.Fatal("keyword rankings tool returned an error")
	}

	t.Logf("structured content: %+v", result.StructuredContent)

	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal output: %v", err)
	}

	var output KeywordRankingsOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if len(output.Rankings) == 0 {
		t.Fatal("expected at least one ranking for the tracked keyword")
	}
}
