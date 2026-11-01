package main

import (
	"app-intelligence-mcp/internal/astro"
	"app-intelligence-mcp/internal/constants"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RankingSnapshot struct {
	Keyword      string
	CurrentRank  *int
	PreviousRank *int
}

type KeywordRankingsInput struct {
	Keyword  string `json:"keyword" jsonschema:"Keyword to search for"`
	Store    string `json:"store" jsonschema:"App Store country code, for example es or us"`
	Platform string `json:"platform" jsonschema:"Platform, for example iphone or ipad"`
}

type KeywordRankingItem struct {
	App                string           `json:"app"`
	Keyword            string           `json:"keyword"`
	Store              string           `json:"store"`
	Platform           string           `json:"platform"`
	CurrentRank        *int             `json:"current_rank"`
	PreviousRank       *int             `json:"previous_rank"`
	LastUpdate         time.Time        `json:"last_update"`
	CurrentRankStatus  astro.RankStatus `json:"current_rank_status"`
	PreviousRankStatus astro.RankStatus `json:"previous_rank_status"`
}

type KeywordRankingsOutput struct {
	Rankings []KeywordRankingItem `json:"rankings"`
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	endpoint := os.Getenv("ASTRO_MCP_ENDPOINT")
	if endpoint == "" {
		endpoint = constants.AstroDefaultEndpoint
	}

	provider, err := astro.Connect(ctx, endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()

	server := newServer(provider)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Print(err)
		}
	}
}

func newServer(provider RankingProvider) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    constants.AppName,
			Version: constants.AppVersion,
		},
		nil,
	)

	handler := &KeywordHandler{
		provider: provider,
	}

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        constants.KeywordRankingsToolName,
			Description: constants.KeywordRankingsToolDescription,
		},
		handler.GetKeywordRankings,
	)

	return server
}

func (h *KeywordHandler) GetKeywordRankings(ctx context.Context, req *mcp.CallToolRequest, input KeywordRankingsInput) (*mcp.CallToolResult, KeywordRankingsOutput, error) {
	if strings.TrimSpace(input.Keyword) == "" {
		return nil, KeywordRankingsOutput{}, fmt.Errorf("keyword must not be empty")
	}

	if strings.TrimSpace(input.Store) == "" {
		return nil, KeywordRankingsOutput{}, fmt.Errorf("store must not be empty")
	}

	switch input.Platform {
	case constants.AstroPlatformIPhone,
		constants.AstroPlatformIPad,
		constants.AstroPlatformMac,
		constants.AstroPlatformAppleTV,
		constants.AstroPlatformRealityDevice:

	default:
		return nil, KeywordRankingsOutput{}, fmt.Errorf(
			"unsupported platform: %q", input.Platform,
		)
	}

	providerCtx, cancel := context.WithTimeout(ctx, constants.AstroRequestTimeout)
	defer cancel()

	rankings, err := h.provider.GetKeywordPerformance(
		providerCtx,
		input.Keyword,
		input.Store,
		input.Platform,
	)
	if err != nil {
		return nil, KeywordRankingsOutput{}, fmt.Errorf(
			"get keyword rankings: %w", err,
		)
	}

	output := KeywordRankingsOutput{
		Rankings: make([]KeywordRankingItem, 0, len(rankings)),
	}

	for _, ranking := range rankings {
		currentRank, currentStatus := astro.NormalizeRank(ranking.CurrentRanking)
		previousRank, previousStatus := astro.NormalizeRank(ranking.PreviousRanking)

		item := KeywordRankingItem{
			App:                ranking.App,
			Keyword:            ranking.Keyword,
			Store:              ranking.Store,
			Platform:           ranking.Platform,
			CurrentRank:        currentRank,
			CurrentRankStatus:  currentStatus,
			PreviousRank:       previousRank,
			PreviousRankStatus: previousStatus,
			LastUpdate:         ranking.LastUpdate,
		}

		output.Rankings = append(output.Rankings, item)
	}

	return nil, output, nil
}

func mockRanking(keyword string) RankingSnapshot {
	previous := 18
	current := 11

	return RankingSnapshot{
		Keyword:      keyword,
		PreviousRank: &previous,
		CurrentRank:  &current,
	}
}

func (s RankingSnapshot) Change() (int, bool) {
	if s.CurrentRank == nil || s.PreviousRank == nil {
		return 0, false
	}

	return *s.PreviousRank - *s.CurrentRank, true
}
