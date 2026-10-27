package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RankingSnapshot struct {
	Keyword      string
	CurrentRank  *int
	PreviousRank *int
}

type KeywordRankingsInput struct {
	Keyword string `json:"keyword" jsonschema:"Keyword to look up"`
}

type KeywordRankingsOutput struct {
	Keyword      string `json:"keyword"`
	RankChange   int    `json:"rank_change"`
	CurrentRank  int    `json:"current_rank"`
	PreviousRank int    `json:"previous_rank"`
}

func main() {
	server := newServer()

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func newServer() *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "app-intelligence-mcp",
			Version: "0.1.0",
		},
		nil,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_keyword_rankings",
			Description: "Return mock keyword rankings for learning. Not real ASO data.",
		},
		getKeywordRankings,
	)

	return server
}

func getKeywordRankings(ctx context.Context, req *mcp.CallToolRequest, input KeywordRankingsInput) (*mcp.CallToolResult, KeywordRankingsOutput, error) {
	if strings.TrimSpace(input.Keyword) == "" {
		return nil, KeywordRankingsOutput{}, fmt.Errorf("keyword must not be empty")
	}

	previous := 18
	current := 11

	snapshot := RankingSnapshot{
		Keyword:      input.Keyword,
		PreviousRank: &previous,
		CurrentRank:  &current,
	}

	change, _ := snapshot.Change()

	output := KeywordRankingsOutput{
		Keyword:      input.Keyword,
		RankChange:   change,
		CurrentRank:  current,
		PreviousRank: previous,
	}

	return nil, output, nil
}

func (s RankingSnapshot) Change() (int, bool) {
	if s.CurrentRank == nil || s.PreviousRank == nil {
		return 0, false
	}

	return *s.PreviousRank - *s.CurrentRank, true
}
