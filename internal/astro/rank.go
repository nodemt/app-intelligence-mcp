package astro

import "app-intelligence-mcp/internal/constants"

type RankStatus string

const (
	RankStatusRanked        RankStatus = "ranked"
	RankStatusMissing       RankStatus = "missing"
	RankStatusOutsideCharts RankStatus = "outside_charts"
)

func NormalizeRank(rank *int) (*int, RankStatus) {
	if rank == nil {
		return nil, RankStatusMissing
	}

	if *rank == constants.AstroOutsideChartRank {
		return nil, RankStatusOutsideCharts
	}

	return rank, RankStatusRanked
}
