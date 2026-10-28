package astro

type ComparisonStatus string

const (
	ComparisonAvailable        ComparisonStatus = "available"
	ComparisonInsufficientData ComparisonStatus = "insufficient_data"
	ComparisonMissingRank      ComparisonStatus = "missing_rank"
	ComparisonOutsideCharts    ComparisonStatus = "outside_charts"
)

func HistoryChange(history []RankingPoint) (int, ComparisonStatus) {
	if len(history) < 2 {
		return 0, ComparisonInsufficientData
	}

	first := history[0]
	last := history[len(history)-1]

	if first.Ranking == nil || last.Ranking == nil {
		return 0, ComparisonMissingRank
	}

	if *first.Ranking == 1000 || *last.Ranking == 1000 {
		return 0, ComparisonOutsideCharts
	}

	return *first.Ranking - *last.Ranking, ComparisonAvailable
}
