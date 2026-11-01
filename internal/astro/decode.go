package astro

import (
	"encoding/json"
	"fmt"
)

func decodeRankings(text string) ([]Ranking, error) {
	var rankings []Ranking

	if err := json.Unmarshal([]byte(text), &rankings); err != nil {
		return nil, fmt.Errorf("decode Astro keyword performance: %w", err)
	}

	if rankings == nil {
		return nil, fmt.Errorf("Astro keyword performance response must be an array, got null")
	}

	return rankings, nil
}
