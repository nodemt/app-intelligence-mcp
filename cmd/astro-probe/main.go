package main

import (
	"app-intelligence-mcp/internal/astro"
	"context"
	"encoding/json"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := astro.Connect(ctx, "http://127.0.0.1:8089/mcp")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	result, err := client.ListTools(ctx)
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

	rankings, err := client.GetKeywordPerformance(ctx, "dibujo digital", "es", "iphone")
	if err != nil {
		log.Fatal(err)
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
			ranking.LastUpdate.Format(time.RFC3339),
		)

		change, status := astro.HistoryChange(ranking.History)
		if status != astro.ComparisonAvailable {
			log.Printf("comparison unavailable: %s", status)
			continue
		}

		first := ranking.History[0]
		last := ranking.History[len(ranking.History)-1]

		log.Printf("from=%s to=%s change=%d", first.Date.Format(time.RFC3339), last.Date.Format(time.RFC3339), change)
	}

}
