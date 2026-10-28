package astro

import "testing"

func TestNormalizeRankOutsideCharts(t *testing.T) {
	rank := 1000
	value, status := NormalizeRank(&rank)

	if value != nil || status != RankStatusOutsideCharts {
		t.Errorf(
			"NormalizeRank(1000) = (%v, %q), want (nil, %q)",
			value, status, RankStatusOutsideCharts,
		)
	}
}

func TestNormalizeRankRanked(t *testing.T) {
	rank := 131
	value, status := NormalizeRank(&rank)

	if value == nil {
		t.Fatal("NormalizeRank(131) returned nil, want a rank value")
	}

	if *value != 131 || status != RankStatusRanked {
		t.Errorf(
			"NormalizeRank(131) = (%d, %q), want (131, %q)",
			*value, status, RankStatusRanked,
		)
	}
}

func TestNormalizeRankMissing(t *testing.T) {
	value, status := NormalizeRank(nil)

	if value != nil || status != RankStatusMissing {
		t.Errorf(
			"NormalizeRank(nil) = (%v, %q), want (nil, %q)",
			value, status, RankStatusMissing,
		)
	}
}
