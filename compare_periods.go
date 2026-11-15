package main

import (
	"app-intelligence-mcp/internal/astro"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DateRangeInput struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type CompareAppPeriodsInput struct {
	AppIntelligenceInput
	Before DateRangeInput `json:"before"`
	After  DateRangeInput `json:"after"`
	Metric string         `json:"metric" jsonschema:"Exact metric name in the saved Mixpanel report"`
}

type CompareAppPeriodsOutput struct {
	Before   DateRangeInput            `json:"before"`
	After    DateRangeInput            `json:"after"`
	Rankings []KeywordPeriodComparison `json:"rankings"`
	Mixpanel MixpanelPeriodComparison  `json:"mixpanel"`
}

type PeriodRanking struct {
	ObservedAt *time.Time       `json:"observed_at"`
	Rank       *int             `json:"rank"`
	Status     astro.RankStatus `json:"status"`
}

type KeywordPeriodComparison struct {
	App        string        `json:"app"`
	Keyword    string        `json:"keyword"`
	Before     PeriodRanking `json:"before"`
	After      PeriodRanking `json:"after"`
	RankChange *int          `json:"rank_change"`
}

type mixpanelMetricTable struct {
	Headers []string            `json:"headers"`
	Rows    [][]json.RawMessage `json:"rows"`
}

type MixpanelPeriodSummary struct {
	Total        *int64 `json:"total"`
	ObservedDays int    `json:"observed_days"`
	ExpectedDays int    `json:"expected_days"`
}

type MixpanelPeriodComparison struct {
	Metric      string                `json:"metric"`
	Before      MixpanelPeriodSummary `json:"before"`
	After       MixpanelPeriodSummary `json:"after"`
	TotalChange *int64                `json:"total_change"`
}

func findMixpanelMetric(results json.RawMessage, metric string) (mixpanelMetricTable, error) {
	var tables map[string]mixpanelMetricTable

	if len(results) == 0 {
		return mixpanelMetricTable{}, fmt.Errorf("Mixpanel report has no results")
	}

	if err := json.Unmarshal(results, &tables); err != nil {
		return mixpanelMetricTable{}, fmt.Errorf("decode Mixpanel results: %w", err)
	}

	table, ok := tables[metric]
	if !ok {
		return mixpanelMetricTable{}, fmt.Errorf(
			"metric %q not found in Mixpanel report results", metric,
		)
	}

	return table, nil
}

func mixpanelMetricColumns(headers []string) (int, int, error) {
	dateIndex, valueIndex := -1, -1

	for i, header := range headers {
		switch header {
		case "date":
			dateIndex = i
		case "value":
			valueIndex = i
		}
	}

	if dateIndex == -1 || valueIndex == -1 {
		return 0, 0, fmt.Errorf("Mixpanel metric must have date and value columns")
	}

	return dateIndex, valueIndex, nil
}

func parseMixpanelDailyRow(row []json.RawMessage, dateIndex, valueIndex int) (time.Time, *int64, error) {
	if dateIndex < 0 || valueIndex < 0 ||
		dateIndex >= len(row) || valueIndex >= len(row) {
		return time.Time{}, nil, fmt.Errorf("Mixpanel row is missing a date or value cell")
	}

	var dateText string
	if err := json.Unmarshal(row[dateIndex], &dateText); err != nil {
		return time.Time{}, nil, fmt.Errorf("decode Mixpanel date: %w", err)
	}

	day, err := time.Parse(time.DateOnly, dateText)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("expected a daily Mixpanel date, got %q: %w", dateText, err)
	}

	var count *int64
	if err := json.Unmarshal(row[valueIndex], &count); err != nil {
		return time.Time{}, nil, fmt.Errorf("decode Mixpanel event count: %w", err)
	}

	return day, count, nil
}

func mixpanelDailyCounts(table mixpanelMetricTable) (map[time.Time]*int64, error) {
	dateIndex, valueIndex, err := mixpanelMetricColumns(table.Headers)
	if err != nil {
		return nil, err
	}

	counts := make(map[time.Time]*int64, len(table.Rows))
	for i, row := range table.Rows {
		day, count, err := parseMixpanelDailyRow(row, dateIndex, valueIndex)
		if err != nil {
			return nil, fmt.Errorf("Mixpanel row %d: %w", i+1, err)
		}

		if _, exists := counts[day]; exists {
			return nil, fmt.Errorf("multiple Mixpanel rows for %s", day.Format(time.DateOnly))
		}

		counts[day] = count
	}

	return counts, nil
}

func summarizeMixpanelPeriod(counts map[time.Time]*int64, from, to time.Time) MixpanelPeriodSummary {
	var summary MixpanelPeriodSummary
	var total int64

	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		summary.ExpectedDays++

		count := counts[day]
		if count == nil {
			continue
		}

		summary.ObservedDays++
		total += *count
	}

	if summary.ObservedDays > 0 {
		summary.Total = &total
	}

	return summary
}

func parseDateRange(input DateRangeInput) (time.Time, time.Time, error) {
	from, err := time.Parse(time.DateOnly, input.From)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid from date %q: %w", input.From, err)
	}

	to, err := time.Parse(time.DateOnly, input.To)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid to date %q: %w", input.To, err)
	}

	if from.After(to) {
		return time.Time{}, time.Time{}, fmt.Errorf("from date must be on or before to date")
	}

	return from, to, nil
}

func validateComparePeriods(input CompareAppPeriodsInput) error {
	_, beforeEnd, err := parseDateRange(input.Before)
	if err != nil {
		return fmt.Errorf("before period: %w", err)
	}

	afterStart, _, err := parseDateRange(input.After)
	if err != nil {
		return fmt.Errorf("after period: %w", err)
	}

	if !beforeEnd.Before(afterStart) {
		return fmt.Errorf("before period must end before the after period starts")
	}

	return nil
}

func latestRankingInPeriod(history []astro.RankingPoint, from, to time.Time) (astro.RankingPoint, bool) {
	endExclusive := to.AddDate(0, 0, 1)
	var latest astro.RankingPoint
	found := false

	for _, h := range history {
		if h.Date.Before(from) || !h.Date.Before(endExclusive) {
			continue
		}

		if !found || h.Date.After(latest.Date) {
			latest = h
			found = true
		}
	}

	return latest, found
}

func summarizePeriodRanking(history []astro.RankingPoint, from, to time.Time) PeriodRanking {
	point, found := latestRankingInPeriod(history, from, to)

	if !found {
		return PeriodRanking{Status: astro.RankStatusMissing}
	}

	rank, status := astro.NormalizeRank(point.Ranking)

	return PeriodRanking{
		ObservedAt: &point.Date,
		Rank:       rank,
		Status:     status,
	}
}

func (h *appIntelligenceHandler) compareAppPeriods(ctx context.Context, req *mcp.CallToolRequest, input CompareAppPeriodsInput) (*mcp.CallToolResult, CompareAppPeriodsOutput, error) {
	if err := validateComparePeriods(input); err != nil {
		return nil, CompareAppPeriodsOutput{}, err
	}

	rankings, err := h.astro.fetchKeywordRankings(ctx, input.KeywordRankingsInput)
	if err != nil {
		return nil, CompareAppPeriodsOutput{}, err
	}

	beforeFrom, beforeTo, err := parseDateRange(input.Before)
	afterFrom, afterTo, _ := parseDateRange(input.After)

	output := CompareAppPeriodsOutput{
		Before:   input.Before,
		After:    input.After,
		Rankings: make([]KeywordPeriodComparison, 0, len(rankings)),
	}

	for _, ranking := range rankings {
		before := summarizePeriodRanking(ranking.History, beforeFrom, beforeTo)
		after := summarizePeriodRanking(ranking.History, afterFrom, afterTo)

		comparison := KeywordPeriodComparison{
			App:     ranking.App,
			Keyword: ranking.Keyword,
			Before:  before,
			After:   after,
		}

		if before.Rank != nil && after.Rank != nil {
			change := *before.Rank - *after.Rank
			comparison.RankChange = &change
		}

		output.Rankings = append(output.Rankings, comparison)
	}

	_, report, err := h.mixpanel.GetSavedProductReport(ctx, req, input.SavedProductReportInput)
	if err != nil {
		return nil, CompareAppPeriodsOutput{}, fmt.Errorf("Mixpanel: %w", err)
	}

	table, err := findMixpanelMetric(report.Results, input.Metric)
	if err != nil {
		return nil, CompareAppPeriodsOutput{}, err
	}

	counts, err := mixpanelDailyCounts(table)
	if err != nil {
		return nil, CompareAppPeriodsOutput{}, err
	}

	before := summarizeMixpanelPeriod(counts, beforeFrom, beforeTo)
	after := summarizeMixpanelPeriod(counts, afterFrom, afterTo)

	output.Mixpanel = MixpanelPeriodComparison{
		Metric: input.Metric,
		Before: before,
		After:  after,
	}

	if before.Total != nil && after.Total != nil && before.ObservedDays == before.ExpectedDays && after.ObservedDays == after.ExpectedDays {
		change := *after.Total - *before.Total
		output.Mixpanel.TotalChange = &change
	}

	return nil, output, nil
}
