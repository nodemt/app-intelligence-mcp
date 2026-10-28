package astro

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

	if *rank == 1000 {
		return nil, RankStatusOutsideCharts
	}

	return rank, RankStatusRanked
}
