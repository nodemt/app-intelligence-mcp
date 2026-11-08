package mixpanel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSavedReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}

		query := r.URL.Query()
		if got := query.Get("project_id"); got != "123" {
			t.Errorf("project_id = %q, want 123", got)
		}
		if got := query.Get("bookmark_id"); got != "456" {
			t.Errorf("bookmark_id = %q, want 456", got)
		}

		username, secret, ok := r.BasicAuth()
		if !ok || username != "test-user" || secret != "test-secret" {
			t.Error("request is missing the expected Basic Auth credentials")
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"computed_at": "2026-11-08T12:00:00Z",
			"date_range": {
				"from_date": "2026-11-01T00:00:00Z",
				"to_date": "2026-11-07T23:59:59Z"
			},
			"headers": ["$event"],
			"series": {
				"First Image Created": {"2026-11-01T00:00:00Z": 42}
			}
		}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-user", "test-secret")
	report, err := client.GetSavedReport(context.Background(), 123, 456)
	if err != nil {
		t.Fatalf("GetSavedReport: %v", err)
	}

	if report.ComputedAt != "2026-11-08T12:00:00Z" ||
		report.DateRange.FromDate != "2026-11-01T00:00:00Z" ||
		report.DateRange.ToDate != "2026-11-07T23:59:59Z" {
		t.Errorf("unexpected report dates: computed_at=%q, range=%+v",
			report.ComputedAt, report.DateRange)
	}

	var series map[string]map[string]int
	if err := json.Unmarshal(report.Series, &series); err != nil {
		t.Fatalf("decode report series: %v", err)
	}
	if got := series["First Image Created"]["2026-11-01T00:00:00Z"]; got != 42 {
		t.Errorf("first image count = %d, want 42", got)
	}
}
