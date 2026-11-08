package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSavedProductReportTool(t *testing.T) {
	const reportJSON = `{
		"computed_at": "2026-11-08T12:00:00Z",
		"date_range": {
			"from_date": "2026-11-01T00:00:00Z",
			"to_date": "2026-11-07T23:59:59Z"
		},
		"headers": ["$event"],
		"series": {"First Image Created": {"2026-11-01T00:00:00Z": 42}}
	}`

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("project_id") != "123" || query.Get("bookmark_id") != "456" {
			t.Errorf("unexpected report query: %v", query)
			http.Error(w, "unexpected report query", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, reportJSON)
	}))
	defer upstream.Close()

	t.Setenv("MIXPANEL_QUERY_ENDPOINT", upstream.URL)
	t.Setenv("MIXPANEL_SERVICE_ACCOUNT_USERNAME", "test-user")
	t.Setenv("MIXPANEL_SERVICE_ACCOUNT_SECRET", "test-secret")
	t.Setenv("MIXPANEL_PROJECT_ID", "123")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server := newServer(&fakeRankingProvider{})
	if err := registerMixpanelTool(server); err != nil {
		t.Fatalf("register Mixpanel tool: %v", err)
	}

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
		Name:      "get_saved_product_report",
		Arguments: map[string]any{"bookmark_id": 456},
	})
	if err != nil {
		t.Fatalf("call Mixpanel tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("Mixpanel tool returned an error: %+v", result.Content)
	}

	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}

	var report struct {
		ComputedAt string                    `json:"computed_at"`
		Series     map[string]map[string]int `json:"series"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode structured report with object-valued series: %v", err)
	}
	if report.ComputedAt != "2026-11-08T12:00:00Z" {
		t.Errorf("computed_at = %q, want 2026-11-08T12:00:00Z", report.ComputedAt)
	}
	if count := report.Series["First Image Created"]["2026-11-01T00:00:00Z"]; count != 42 {
		t.Errorf("first image count = %d, want 42", count)
	}
}
