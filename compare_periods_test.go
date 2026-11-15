package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"app-intelligence-mcp/internal/astro"
	"app-intelligence-mcp/internal/constants"
	"app-intelligence-mcp/internal/mixpanel"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCompareAppPeriodsTool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	beforeRank, afterRank := 18, 11
	rankings := &fakeRankingProvider{rankings: []astro.Ranking{{
		Keyword: "dibujo digital",
		History: []astro.RankingPoint{
			{Date: time.Date(2026, time.November, 7, 18, 0, 0, 0, time.UTC), Ranking: &beforeRank},
			{Date: time.Date(2026, time.November, 9, 18, 0, 0, 0, time.UTC), Ranking: &afterRank},
		},
	}}}
	const results = `{"App Opens":{"headers":["date","value"],"rows":[["2026-11-07",40],["2026-11-08",20],["2026-11-09",32]]}}`
	reports := stubSavedReportProvider{
		report: mixpanel.SavedReport{Results: json.RawMessage(results)},
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
		Name: constants.CompareAppPeriodsToolName,
		Arguments: map[string]any{
			"keyword":     "dibujo digital",
			"store":       "es",
			"platform":    "iphone",
			"bookmark_id": 93903711,
			"metric":      "App Opens",
			"before":      map[string]string{"from": "2026-11-07", "to": "2026-11-07"},
			"after":       map[string]string{"from": "2026-11-08", "to": "2026-11-09"},
		},
	})
	if err != nil {
		t.Fatalf("call comparison tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("comparison tool returned an error: %+v", result.Content)
	}
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var output CompareAppPeriodsOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("decode comparison output: %v", err)
	}
	if len(output.Rankings) != 1 {
		t.Fatalf("rankings count = %d, want 1", len(output.Rankings))
	}
	if change := output.Rankings[0].RankChange; change == nil {
		t.Error("rank_change is missing, want 7")
	} else if *change != 7 {
		t.Errorf("rank_change = %d, want 7", *change)
	}
	if change := output.Mixpanel.TotalChange; change == nil {
		t.Error("total_change is missing, want 12")
	} else if *change != 12 {
		t.Errorf("total_change = %d, want 12", *change)
	}
}

func TestSummarizePeriodRanking(t *testing.T) {
	from := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.November, 7, 0, 0, 0, 0, time.UTC)
	latest := to.Add(18 * time.Hour)
	olderRank, latestRank, nextDayRank := 18, 11, 9

	history := []astro.RankingPoint{
		{Date: from.AddDate(0, 0, 2), Ranking: &olderRank},
		{Date: latest, Ranking: &latestRank},
		{Date: to.AddDate(0, 0, 1), Ranking: &nextDayRank},
	}

	got := summarizePeriodRanking(history, from, to)
	if got.Rank == nil {
		t.Fatal("rank is missing")
	}
	if *got.Rank != latestRank {
		t.Errorf("rank = %d, want %d", *got.Rank, latestRank)
	}
	if got.ObservedAt == nil || !got.ObservedAt.Equal(latest) {
		t.Errorf("observed_at = %v, want %v", got.ObservedAt, latest)
	}
	if got.Status != astro.RankStatusRanked {
		t.Errorf("status = %q, want %q", got.Status, astro.RankStatusRanked)
	}
}
