package main

import (
	"app-intelligence-mcp/internal/astro"
	"context"
)

type RankingProvider interface {
	GetKeywordPerformance(
		ctx context.Context,
		keyword string,
		store string,
		platform string,
	) ([]astro.Ranking, error)
}

type KeywordHandler struct {
	provider RankingProvider
}
