package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"app-intelligence-mcp/internal/astro"
	"app-intelligence-mcp/internal/constants"
	"app-intelligence-mcp/internal/mixpanel"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type stubSavedReportProvider struct {
	report mixpanel.SavedReport
	err    error
}

func (p stubSavedReportProvider) GetSavedReport(_ context.Context, projectID, bookmarkID int) (mixpanel.SavedReport, error) {
	if projectID != 4081780 || bookmarkID != 93903711 {
		return mixpanel.SavedReport{}, errors.New("unexpected project or bookmark ID")
	}
	return p.report, p.err
}

func TestAppIntelligenceTool(t *testing.T) {
	const results = `{"App Opens":{"headers":["date","value"],"rows":[["2026-11-08",1]]}}`
	tests := []struct {
		name      string
		results   string
		astroErr  error
		reportErr error
		wantError string
	}{
		{name: "both sources", results: results},
		{name: "empty report", results: "{}"},
		{name: "Astro error", astroErr: errors.New("unavailable"), wantError: "Astro:"},
		{name: "Mixpanel error", reportErr: errors.New("unavailable"), wantError: "Mixpanel:"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			rank := 131
			rankings := &fakeRankingProvider{
				rankings: []astro.Ranking{{Keyword: "dibujo digital", CurrentRanking: &rank}},
				err:      tc.astroErr,
			}
			reports := stubSavedReportProvider{
				report: mixpanel.SavedReport{Results: json.RawMessage(tc.results)},
				err:    tc.reportErr,
			}
			server := newServer(rankings)
			addMixpanelTools(server, rankings, reports, 4081780)

			serverTransport, clientTransport := mcp.NewInMemoryTransports()
			serverSession, err := server.Connect(ctx, serverTransport, nil)
			if err != nil {
				t.Fatalf("connect server: %v", err)
			}
			defer serverSession.Close()

			client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.1.0"}, nil)
			session, err := client.Connect(ctx, clientTransport, nil)
			if err != nil {
				t.Fatalf("connect client: %v", err)
			}
			defer session.Close()

			result, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name: constants.AppIntelligenceToolName,
				Arguments: map[string]any{
					"keyword":     "dibujo digital",
					"store":       "es",
					"platform":    "iphone",
					"bookmark_id": 93903711,
				},
			})
			if err != nil {
				t.Fatalf("call combined tool: %v", err)
			}
			if tc.wantError != "" {
				if !result.IsError || len(result.Content) == 0 {
					t.Fatalf("expected tool error containing %q, got %+v", tc.wantError, result)
				}
				message, ok := result.Content[0].(*mcp.TextContent)
				if !ok || !strings.Contains(message.Text, tc.wantError) {
					t.Fatalf("tool error = %+v, want %q", result.Content, tc.wantError)
				}
				return
			}
			if result.IsError {
				t.Fatalf("combined tool returned an error: %+v", result.Content)
			}

			data, err := json.Marshal(result.StructuredContent)
			if err != nil {
				t.Fatalf("marshal structured content: %v", err)
			}
			var output AppIntelligenceOutput
			if err := json.Unmarshal(data, &output); err != nil {
				t.Fatalf("decode combined output: %v", err)
			}
			if len(output.Astro.Rankings) != 1 {
				t.Fatalf("Astro rankings = %d, want 1", len(output.Astro.Rankings))
			}
			if got := output.Astro.Rankings[0].CurrentRank; got == nil || *got != 131 {
				t.Errorf("Astro current rank = %v, want 131", got)
			}
			if got := string(output.Mixpanel.Results); got != tc.results {
				t.Errorf("Mixpanel results = %s, want %s", got, tc.results)
			}
		})
	}
}
