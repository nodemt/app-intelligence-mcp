package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AstroRanking struct {
	App             string              `json:"app"`
	Keyword         string              `json:"keyword"`
	Store           string              `json:"store"`
	Platform        string              `json:"platform"`
	CurrentRanking  *int                `json:"currentRanking"`
	PreviousRanking *int                `json:"previousRanking"`
	LastUpdate      string              `json:"lastUpdate"`
	History         []AstroRankingPoint `json:"history"`
}

type AstroRankingPoint struct {
	Date    time.Time `json:"date"`
	Ranking *int      `json:"ranking"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "app-intelligence-astro-probe",
			Version: "0.1.0",
		},
		nil,
	)

	transport := &mcp.StreamableClientTransport{
		Endpoint: "http://127.0.0.1:8089/mcp",
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, tool := range result.Tools {
		if tool.Name != "get_keyword_performance" {
			continue
		}

		data, err := json.MarshalIndent(tool.InputSchema, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("get_keyword_performance input schema:\n%s", data)
	}

	rankingResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_keyword_performance",
		Arguments: map[string]any{
			"keyword":        "dibujo digital",
			"store":          "es",
			"platform":       "iphone",
			"includeHistory": true,
			"period":         "all",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if rankingResult.IsError {
		for _, content := range rankingResult.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				log.Printf("Astro tool error: %s", textContent.Text)
			}
		}

		log.Fatal("get_keyword_performance failed")
	}

	data, err := json.MarshalIndent(rankingResult, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("get_keyword_performance result:\n%s", data)

	for _, content := range rankingResult.Content {
		textContent, ok := content.(*mcp.TextContent)
		if !ok {
			continue
		}

		var rankings []AstroRanking

		if err := json.Unmarshal([]byte(textContent.Text), &rankings); err != nil {
			log.Fatalf("decode rankings: %v", err)
		}

		for _, ranking := range rankings {
			for _, historyPoint := range ranking.History {
				if historyPoint.Ranking == nil {
					log.Printf("date=%s rank=missing", historyPoint.Date.Format(time.RFC3339))
					continue
				}

				log.Printf(
					"date=%s rank=%d",
					historyPoint.Date.Format(time.RFC3339),
					*historyPoint.Ranking,
				)

			}
			log.Printf(
				"app=%s keyword=%q store=%q updated=%s",
				ranking.App,
				ranking.Keyword,
				ranking.Store,
				ranking.LastUpdate,
			)

			if len(ranking.History) < 2 {
				log.Print("not enough history to compare")
				continue
			}

			first := ranking.History[0]
			last := ranking.History[len(ranking.History)-1]

			if first.Ranking == nil || last.Ranking == nil {
				log.Printf("comparison unavailable: missing rank")
				continue
			}

			if *first.Ranking == 1000 || *last.Ranking == 1000 {
				log.Print("comparison unavailable: endpoint outside charts")
				continue
			}

			change := *first.Ranking - *last.Ranking

			log.Printf("from=%s to=%s change=%d", first.Date.Format(time.RFC3339), last.Date.Format(time.RFC3339), change)
		}
	}
}
