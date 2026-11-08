package mixpanel

import "encoding/json"

type DateRange struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

type SavedReport struct {
	ComputedAt string          `json:"computed_at"`
	DateRange  DateRange       `json:"date_range"`
	Headers    []string        `json:"headers"`
	Series     json.RawMessage `json:"series"`
}
