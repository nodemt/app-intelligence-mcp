package astro

import (
	"testing"
)

func TestHistoryChangeDecline(t *testing.T) {
	first := 121
	last := 131

	history := []RankingPoint{
		{Ranking: &first},
		{Ranking: &last},
	}

	change, status := HistoryChange(history)
	if change != -10 || status != ComparisonAvailable {
		t.Errorf("HistoryChange() = (%d, %q), want (-10, %q)", change, status, ComparisonAvailable)
	}
}

func TestHistoryChangeEmpty(t *testing.T) {
	change, status := HistoryChange(nil)

	if change != 0 || status != ComparisonInsufficientData {
		t.Errorf(
			"HistoryChange(nil) = (%d, %q), want (0, %q)",
			change, status, ComparisonInsufficientData,
		)
	}
}

func TestHistoryChangeSinglePoint(t *testing.T) {
	rank := 121

	history := []RankingPoint{
		{Ranking: &rank},
	}

	change, status := HistoryChange(history)

	if change != 0 || status != ComparisonInsufficientData {
		t.Errorf(
			"HistoryChange(single point) = (%d, %q), want (0, %q)",
			change, status, ComparisonInsufficientData,
		)
	}
}

func TestHistoryChangeOutsideCharts(t *testing.T) {
	first := 1000
	last := 131

	history := []RankingPoint{
		{Ranking: &first},
		{Ranking: &last},
	}

	change, status := HistoryChange(history)
	if change != 0 || status != ComparisonOutsideCharts {
		t.Errorf(
			"HistoryChange(history) = (%d, %q), want (0, %q)",
			change, status, ComparisonOutsideCharts,
		)
	}
}

func TestHistoryChangeLastOutsideCharts(t *testing.T) {
	first := 121
	last := 1000

	history := []RankingPoint{
		{Ranking: &first},
		{Ranking: &last},
	}

	change, status := HistoryChange(history)
	if change != 0 || status != ComparisonOutsideCharts {
		t.Errorf(
			"HistoryChange(history) = (%d, %q), want (0, %q)",
			change, status, ComparisonOutsideCharts,
		)
	}
}

func TestHistoryChangeMissingRank(t *testing.T) {
	last := 131

	history := []RankingPoint{
		{},
		{Ranking: &last},
	}

	change, status := HistoryChange(history)
	if change != 0 || status != ComparisonMissingRank {
		t.Errorf(
			"HistoryChange(missing rank) = (%d, %q), want (0, %q)",
			change, status, ComparisonMissingRank,
		)
	}
}
