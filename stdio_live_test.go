//go:build integration

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStdioConnectionLive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	binaryPath, err := filepath.Abs("app-intelligence-mcp")
	if err != nil {
		t.Fatalf("resolve binary path: %v", err)
	}

	command := exec.CommandContext(ctx, binaryPath)
	command.Stderr = os.Stderr

	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "stdio-test-client",
			Version: "0.1.0",
		},
		nil,
	)

	session, err := client.Connect(ctx, &mcp.CommandTransport{
		Command: command,
	}, nil)
	if err != nil {
		t.Fatalf("connect over stdio: %v", err)
	}
	defer session.Close()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools over stdio: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "get_keyword_rankings" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("get_keyword_rankings not found over stdio")
	}

	callResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_keyword_rankings",
		Arguments: map[string]any{
			"keyword":  "dibujo digital",
			"store":    "es",
			"platform": "iphone",
		},
	})
	if err != nil {
		t.Fatalf("call tool over stdio: %v", err)
	}

	if callResult.IsError {
		for _, content := range callResult.Content {
			if text, ok := content.(*mcp.TextContent); ok {
				t.Logf("tool error: %s", text.Text)
			}
		}
		t.Fatal("tool returned an error over stdio")
	}

	t.Logf("structured content: %+v", callResult.StructuredContent)
}
