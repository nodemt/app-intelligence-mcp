package astro

import "time"

type Ranking struct {
	App             string         `json:"app"`
	Keyword         string         `json:"keyword"`
	Store           string         `json:"store"`
	Platform        string         `json:"platform"`
	CurrentRanking  *int           `json:"currentRanking"`
	PreviousRanking *int           `json:"previousRanking"`
	LastUpdate      time.Time      `json:"lastUpdate"`
	History         []RankingPoint `json:"history"`
}

type RankingPoint struct {
	Date    time.Time `json:"date"`
	Ranking *int      `json:"ranking"`
}
