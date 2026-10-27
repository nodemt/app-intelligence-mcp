package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestKeywordRankingsTool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Server
	server := newServer()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	defer serverSession.Close()

	// Client
	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "test-client",
			Version: "0.1.0",
		},
		nil,
	)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	defer session.Close()

	// Tool List
	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	found := false

	for _, tool := range result.Tools {
		if tool.Name == "get_keyword_rankings" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("get_keyword_rankings not found in tools/list")
	}

	// Use from MCP
	callResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_keyword_rankings",
		Arguments: map[string]any{
			"keyword": "photo editor",
		},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}

	if callResult.IsError {
		t.Fatalf("tool returned an error: %+v", callResult.Content)
	}

	t.Logf("structured content: %+v", callResult.StructuredContent)

	data, err := json.Marshal(callResult.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}

	var output KeywordRankingsOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("decode structured content: %v", err)
	}

	if output.Keyword != "photo editor" {
		t.Errorf("Keyword %q, want %q", output.Keyword, "photo editor")
	}

	if output.CurrentRank != 11 {
		t.Errorf("CurrentRank = %d, want 11", output.CurrentRank)
	}

	if output.PreviousRank != 18 {
		t.Errorf("PreviousRank = %d, want 18", output.PreviousRank)
	}

	if output.RankChange != 7 {
		t.Errorf("RankChange = %d, want 7", output.RankChange)
	}

	emptyResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_keyword_rankings",
		Arguments: map[string]any{
			"keyword": "	",
		},
	})
	if err != nil {
		t.Fatalf("call tool with blank keyword: %v", err)
	}
	if !emptyResult.IsError {
		t.Error("expected tool error for blank keyword")
	}
}

func TestRankingSnapshotChangeTable(t *testing.T) {
	previous := 18
	current := 17

	tests := []struct {
		name         string
		previousRank *int
		currentRank  *int
		wantChange   int
		wantOK       bool
	}{
		{
			name:         "improvement",
			previousRank: &previous,
			currentRank:  &current,
			wantChange:   1,
			wantOK:       true,
		},
		{
			name:         "missing_previous_rank",
			previousRank: nil,
			currentRank:  &current,
			wantChange:   0,
			wantOK:       false,
		},
		{
			name:         "decline",
			previousRank: &current,
			currentRank:  &previous,
			wantChange:   -1,
			wantOK:       true,
		},
		{
			name:         "unchanged",
			previousRank: &previous,
			currentRank:  &previous,
			wantChange:   0,
			wantOK:       true,
		},
		{
			name:         "missing_current_rank",
			previousRank: &previous,
			currentRank:  nil,
			wantChange:   0,
			wantOK:       false,
		},
		{
			name:         "missing_both_ranks",
			previousRank: nil,
			currentRank:  nil,
			wantChange:   0,
			wantOK:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := RankingSnapshot{
				PreviousRank: tc.previousRank,
				CurrentRank:  tc.currentRank,
			}

			change, ok := snapshot.Change()

			if change != tc.wantChange || ok != tc.wantOK {
				t.Errorf(
					"Change = (%d, %t), want (%d, %t)",
					change, ok, tc.wantChange, tc.wantOK,
				)
			}
		})
	}
}
