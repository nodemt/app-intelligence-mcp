package astro

import (
	"testing"
	"time"
)

func TestDecodeRankingsInvalidJSON(t *testing.T) {
	_, err := decodeRankings("this is not JSON")

	if err == nil {
		t.Fatalf("expected an error for invalid JSON")
	}
}

func TestDecodeRankingsEmptyList(t *testing.T) {
	rankings, err := decodeRankings("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rankings) != 0 {
		t.Errorf("got %d rankings, want 0", len(rankings))
	}
}

func TestDecodeRankingsValidRecord(t *testing.T) {
	input := `[{
		"keyword": "dibujo digital",
		"currentRanking": 131,
		"lastUpdate": "2026-04-08T15:23:53Z"
	}]`

	rankings, err := decodeRankings(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rankings) != 1 {
		t.Fatalf("got %d rankings, want 1", len(rankings))
	}

	ranking := rankings[0]

	if ranking.Keyword != "dibujo digital" {
		t.Errorf("Keyword = %q, want %q", ranking.Keyword, "dibujo digital")
	}

	if ranking.CurrentRanking == nil {
		t.Fatal("CurrentRanking is nil, want 131")
	}

	if *ranking.CurrentRanking != 131 {
		t.Errorf("CurrentRanking = %d, want 131", *ranking.CurrentRanking)
	}

	wantUpdate := time.Date(2026, time.April, 8, 15, 23, 53, 0, time.UTC)

	if !ranking.LastUpdate.Equal(wantUpdate) {
		t.Errorf("LastUpdate = %s, want %s", ranking.LastUpdate, wantUpdate)
	}

	if ranking.PreviousRanking != nil {
		t.Errorf("PreviousRanking = %v, want nil for omitted field", ranking.PreviousRanking)
	}
}

func TestDecodeRankingsRejectsNull(t *testing.T) {
	_, err := decodeRankings("null")
	if err == nil {
		t.Fatal("expected an error for null response")
	}
}
