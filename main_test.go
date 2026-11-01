package main

import (
	"app-intelligence-mcp/internal/astro"
	"app-intelligence-mcp/internal/constants"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeRankingProvider struct {
	rankings    []astro.Ranking
	calls       int
	err         error
	deadline    time.Time
	hasDeadline bool
	appID       string
}

type waitingRankingProvider struct{}

func (f *fakeRankingProvider) GetKeywordPerformance(ctx context.Context, keyword string, store string, platform string, appID string) ([]astro.Ranking, error) {
	f.calls++
	f.appID = appID
	f.deadline, f.hasDeadline = ctx.Deadline()
	return f.rankings, f.err
}

func (p *waitingRankingProvider) GetKeywordPerformance(ctx context.Context, keyword string, store string, platform string, appID string) ([]astro.Ranking, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestKeywordRankingsTool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	previous := 18
	current := 11

	provider := &fakeRankingProvider{
		rankings: []astro.Ranking{
			{
				App:             "Test App",
				Keyword:         "photo editor",
				Store:           "es",
				Platform:        "iphone",
				PreviousRanking: &previous,
				CurrentRanking:  &current,
				LastUpdate:      time.Date(2026, time.October, 29, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	// Server
	server := newServer(provider)

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
			"keyword":  "photo editor",
			"store":    "es",
			"platform": "iphone",
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

	if len(output.Rankings) != 1 {
		t.Fatalf("got %d rankings, want 1", len(output.Rankings))
	}

	item := output.Rankings[0]
	if item.Keyword != "photo editor" {
		t.Errorf("Keyword = %q, want %q", item.Keyword, "photo editor")
	}

	if item.CurrentRank == nil {
		t.Fatal("CurrentRank is nil, want a rank value")
	}

	if *item.CurrentRank != 11 {
		t.Errorf("CurrentRank = %d, want 11", *item.CurrentRank)
	}

	if item.PreviousRank == nil {
		t.Fatal("PreviousRank is nil, want a rank value")
	}

	if *item.PreviousRank != 18 {
		t.Errorf("PreviousRank = %d, want 18", *item.PreviousRank)
	}

	if item.CurrentRankStatus != astro.RankStatusRanked {
		t.Errorf(
			"CurrentRankStatus = %q, want %q",
			item.CurrentRankStatus, astro.RankStatusRanked,
		)
	}

	if item.PreviousRankStatus != astro.RankStatusRanked {
		t.Errorf(
			"PreviousRankStatus = %q, want %q",
			item.PreviousRankStatus, astro.RankStatusRanked,
		)
	}

	if item.Store != "es" {
		t.Errorf("Store = %q, want %q", item.Store, "es")
	}

	if item.Platform != "iphone" {
		t.Errorf("Platform = %q, want %q", item.Platform, "iphone")
	}

	wantUpdate := time.Date(2026, time.October, 29, 12, 0, 0, 0, time.UTC)

	if !item.LastUpdate.Equal(wantUpdate) {
		t.Errorf("LastUpdate = %s, want %s", item.LastUpdate, wantUpdate)
	}

	emptyResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_keyword_rankings",
		Arguments: map[string]any{
			"keyword":  "	",
			"store":    "es",
			"platform": "iphone",
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

func TestKeywordHandlerRejectsEmptyStore(t *testing.T) {
	provider := &fakeRankingProvider{}
	handler := &KeywordHandler{provider: provider}

	_, _, err := handler.GetKeywordRankings(
		context.Background(),
		nil,
		KeywordRankingsInput{
			Keyword:  "photo editor",
			Store:    " ",
			Platform: constants.AstroPlatformIPhone,
		},
	)

	if err == nil {
		t.Fatal("expected an error for blank store")
	}

	if provider.calls != 0 {
		t.Errorf("provider calls = %d, want 0", provider.calls)
	}
}

func TestKeywordHandlerRejectsUnsupportedPlatform(t *testing.T) {
	provider := &fakeRankingProvider{}
	handler := &KeywordHandler{provider: provider}

	_, _, err := handler.GetKeywordRankings(
		context.Background(),
		nil,
		KeywordRankingsInput{
			Keyword:  "photo editor",
			Store:    "es",
			Platform: "android",
		},
	)

	if err == nil {
		t.Fatal("expected an error for unsupported platform")
	}

	if provider.calls != 0 {
		t.Errorf("provider calls = %d, want 0", provider.calls)
	}
}

func TestKeywordHandlerProviderFailure(t *testing.T) {
	providerErr := errors.New("provider unavailable")

	provider := &fakeRankingProvider{
		err: providerErr,
	}
	handler := &KeywordHandler{provider: provider}

	_, _, err := handler.GetKeywordRankings(
		context.Background(),
		nil,
		KeywordRankingsInput{
			Keyword:  "photo editor",
			Store:    "es",
			Platform: constants.AstroPlatformIPhone,
		},
	)

	if !errors.Is(err, providerErr) {
		t.Errorf("got error %v, want wrapped provider error %v", err, providerErr)
	}

	if provider.calls != 1 {
		t.Errorf("provider calls = %d, want 1", provider.calls)
	}
}

func TestKeywordHandlerRespectsDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	handler := &KeywordHandler{
		provider: &waitingRankingProvider{},
	}

	_, _, err := handler.GetKeywordRankings(
		ctx,
		nil,
		KeywordRankingsInput{
			Keyword:  "photo editor",
			Store:    "es",
			Platform: constants.AstroPlatformIPhone,
		},
	)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got error %v, want context deadline exceeded", err)
	}
}

func TestKeywordHandlerAddsDeadline(t *testing.T) {
	provider := &fakeRankingProvider{}
	handler := &KeywordHandler{provider: provider}

	before := time.Now()

	_, _, err := handler.GetKeywordRankings(
		context.Background(),
		nil,
		KeywordRankingsInput{
			Keyword:  "photo editor",
			Store:    "es",
			Platform: constants.AstroPlatformIPhone,
		},
	)

	after := time.Now()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !provider.hasDeadline {
		t.Fatal("provider context has no deadline")
	}

	earliest := before.Add(constants.AstroRequestTimeout)
	latest := after.Add(constants.AstroRequestTimeout)

	if provider.deadline.Before(earliest) || provider.deadline.After(latest) {
		t.Errorf(
			"deadline = %s, want between %s and %s",
			provider.deadline, earliest, latest,
		)
	}
}

func TestKeywordHandlerPassesAppID(t *testing.T) {
	provider := &fakeRankingProvider{}
	handler := &KeywordHandler{provider: provider}

	_, _, err := handler.GetKeywordRankings(
		context.Background(),
		nil,
		KeywordRankingsInput{
			Keyword:  "photo editor",
			Store:    "es",
			Platform: constants.AstroPlatformIPhone,
			AppID:    "123456789",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.appID != "123456789" {
		t.Errorf("provider appID = %q, want %q", provider.appID, "123456789")
	}
}
