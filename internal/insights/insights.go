package insights

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

type DiagnoseOptions struct {
	Segment string
	MinLoss float64
	Limit   int
}

type DropContributor struct {
	Key            string  `json:"key"`
	CurrentClicks  float64 `json:"currentClicks"`
	ExpectedClicks float64 `json:"expectedClicks"`
	Contribution   float64 `json:"contribution"`
	ShareOfLoss    float64 `json:"shareOfLoss"`
}

type DropReport struct {
	Segment      string            `json:"segment"`
	TotalLoss    float64           `json:"totalLoss"`
	LikelyCause  string            `json:"likelyCause"`
	Contributors []DropContributor `json:"contributors"`
}

func DiagnoseDrop(current, baseline *gsc.AnalyticsResult, opts DiagnoseOptions) DropReport {
	baselineMap := analyticsByKey(baseline)
	currentMap := analyticsByKey(current)
	totalCurrentClicks, totalCurrentImpr, totalCurrentCtr, totalCurrentPos := totals(current)
	totalBaselineClicks, totalBaselineImpr, totalBaselineCtr, totalBaselinePos := totals(baseline)
	report := DropReport{Segment: opts.Segment, TotalLoss: totalCurrentClicks - totalBaselineClicks, LikelyCause: classifyCause(totalCurrentImpr, totalBaselineImpr, totalCurrentCtr, totalBaselineCtr, totalCurrentPos, totalBaselinePos)}
	for key, baselineRow := range baselineMap {
		currentRow := currentMap[key]
		expected := baselineRow.Impressions * baselineRow.CTR
		contribution := currentRow.Clicks - expected
		if contribution > -math.Abs(opts.MinLoss) {
			continue
		}
		report.Contributors = append(report.Contributors, DropContributor{Key: key, CurrentClicks: currentRow.Clicks, ExpectedClicks: expected, Contribution: contribution})
	}
	sort.Slice(report.Contributors, func(i, j int) bool { return report.Contributors[i].Contribution < report.Contributors[j].Contribution })
	if opts.Limit > 0 && len(report.Contributors) > opts.Limit {
		report.Contributors = report.Contributors[:opts.Limit]
	}
	denom := math.Abs(report.TotalLoss)
	if denom == 0 {
		denom = 1
	}
	for i := range report.Contributors {
		report.Contributors[i].ShareOfLoss = math.Abs(report.Contributors[i].Contribution) / denom
	}
	return report
}

// ShiftDateKeys returns a copy of a DATE-keyed result with each date moved by days, so a
// baseline period lines up day by day with the current period it is compared against.
func ShiftDateKeys(result *gsc.AnalyticsResult, days int) (*gsc.AnalyticsResult, error) {
	if result == nil {
		return nil, nil
	}
	shifted := *result
	shifted.Rows = make([]gsc.AnalyticsRow, 0, len(result.Rows))
	for _, row := range result.Rows {
		keys := append([]string(nil), row.Keys...)
		if len(keys) > 0 {
			date, err := time.Parse("2006-01-02", keys[0])
			if err != nil {
				return nil, fmt.Errorf("parse date key %q: %w", keys[0], err)
			}
			keys[0] = date.AddDate(0, 0, days).Format("2006-01-02")
		}
		row.Keys = keys
		shifted.Rows = append(shifted.Rows, row)
	}
	return &shifted, nil
}

type HourlyComparisonRow struct {
	Hour                string  `json:"hour"`
	CurrentClicks       float64 `json:"currentClicks"`
	BaselineClicks      float64 `json:"baselineClicks"`
	ClickDelta          float64 `json:"clickDelta"`
	CurrentImpressions  float64 `json:"currentImpressions"`
	BaselineImpressions float64 `json:"baselineImpressions"`
}

type HourlyComparison struct {
	Provisional bool                  `json:"provisional"`
	Rows        []HourlyComparisonRow `json:"rows"`
}

func BuildHourlyComparison(current, baseline *gsc.AnalyticsResult) HourlyComparison {
	currentMap := analyticsByLocalHour(current)
	baselineMap := analyticsByLocalHour(baseline)
	allKeys := make(map[string]struct{})
	for key := range currentMap {
		allKeys[key] = struct{}{}
	}
	for key := range baselineMap {
		allKeys[key] = struct{}{}
	}
	report := HourlyComparison{Provisional: true}
	for key := range allKeys {
		c := currentMap[key]
		b := baselineMap[key]
		report.Rows = append(report.Rows, HourlyComparisonRow{Hour: key, CurrentClicks: c.Clicks, BaselineClicks: b.Clicks, ClickDelta: c.Clicks - b.Clicks, CurrentImpressions: c.Impressions, BaselineImpressions: b.Impressions})
	}
	sort.Slice(report.Rows, func(i, j int) bool { return report.Rows[i].Hour < report.Rows[j].Hour })
	return report
}

func ClassifyInspectionIssues(result gsc.InspectionResult) []string {
	var issues []string
	if result.RobotsTxtState == "DISALLOWED" {
		issues = append(issues, "robots_blocked")
	}
	if result.IndexingState == "BLOCKED_BY_META_TAG" {
		issues = append(issues, "noindex_meta")
	}
	if result.IndexingState == "BLOCKED_BY_HTTP_HEADER" {
		issues = append(issues, "x_robots_noindex")
	}
	if result.PageFetchState == "NOT_FOUND" {
		issues = append(issues, "404")
	}
	if result.PageFetchState == "SERVER_ERROR" {
		issues = append(issues, "5xx")
	}
	if result.GoogleCanonical != "" && result.UserCanonical != "" && result.GoogleCanonical != result.UserCanonical {
		issues = append(issues, "canonical_mismatch")
	}
	if result.LastCrawlTime == "" {
		issues = append(issues, "never_crawled")
	}
	return issues
}

type OpportunityOptions struct {
	Types          []string
	MinImpressions float64
	Limit          int
}

type Opportunity struct {
	Type                 string  `json:"type"`
	Query                string  `json:"query,omitempty"`
	Page                 string  `json:"page,omitempty"`
	Impressions          float64 `json:"impressions"`
	Clicks               float64 `json:"clicks"`
	CTR                  float64 `json:"ctr"`
	Position             float64 `json:"position"`
	Score                float64 `json:"score"`
	EstimatedExtraClicks float64 `json:"estimatedExtraClicks,omitempty"`
	Action               string  `json:"action,omitempty"`
}

type OpportunityReport struct {
	Items []Opportunity `json:"items"`
}

func FindOpportunities(current, previous *gsc.AnalyticsResult, opts OpportunityOptions) OpportunityReport {
	enabled := make(map[string]struct{}, len(opts.Types))
	for _, kind := range opts.Types {
		enabled[strings.ToLower(kind)] = struct{}{}
	}
	ctrCurve := buildCTRCurve(current, opts.MinImpressions)
	var items []Opportunity
	for _, row := range current.Rows {
		query, page := splitQueryPage(current.Dimensions, row.Keys)
		if row.Impressions < opts.MinImpressions {
			continue
		}
		if _, ok := enabled["near-win"]; ok && row.Position >= 4 && row.Position <= 20 {
			items = append(items, Opportunity{Type: "near-win", Query: query, Page: page, Impressions: row.Impressions, Clicks: row.Clicks, CTR: row.CTR, Position: row.Position, Score: row.Impressions * math.Max(0, 20-row.Position), Action: "refresh content and strengthen title promise"})
		}
		if _, ok := enabled["ctr-gap"]; ok {
			bucket := int(math.Round(row.Position))
			expected := ctrCurve[bucket]
			if expected > row.CTR {
				items = append(items, Opportunity{Type: "ctr-gap", Query: query, Page: page, Impressions: row.Impressions, Clicks: row.Clicks, CTR: row.CTR, Position: row.Position, EstimatedExtraClicks: row.Impressions * (expected - row.CTR), Score: row.Impressions * (expected - row.CTR), Action: "rewrite title/meta and improve snippet intent"})
			}
		}
	}
	if previous != nil {
		prevMap := analyticsByKey(previous)
		for _, row := range current.Rows {
			key := strings.Join(row.Keys, "|")
			prev, ok := prevMap[key]
			if !ok || row.Impressions < opts.MinImpressions {
				continue
			}
			if _, ok := enabled["decay"]; ok && prev.Clicks > row.Clicks {
				query, page := splitQueryPage(current.Dimensions, row.Keys)
				score := prev.Clicks - row.Clicks
				if row.Position > prev.Position {
					score *= 1.5
				}
				items = append(items, Opportunity{Type: "decay", Query: query, Page: page, Impressions: row.Impressions, Clicks: row.Clicks, CTR: row.CTR, Position: row.Position, Score: score, Action: "refresh content and improve relevance"})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	if opts.Limit > 0 && len(items) > opts.Limit {
		items = items[:opts.Limit]
	}
	return OpportunityReport{Items: items}
}

type SitemapAuditReport struct {
	InSitemapWithImpressions int      `json:"inSitemapWithImpressions"`
	WithoutImpressions       []string `json:"withoutImpressions"`
	MissingFromSitemap       []string `json:"missingFromSitemap"`
}

func AuditSitemap(sitemapURLs []string, analytics *gsc.AnalyticsResult) SitemapAuditReport {
	sitemapSet := make(map[string]struct{}, len(sitemapURLs))
	for _, url := range sitemapURLs {
		sitemapSet[url] = struct{}{}
	}
	analyticsSet := make(map[string]float64)
	for _, row := range analytics.Rows {
		if len(row.Keys) == 0 {
			continue
		}
		analyticsSet[row.Keys[0]] += row.Impressions
	}
	report := SitemapAuditReport{}
	for _, url := range sitemapURLs {
		if analyticsSet[url] > 0 {
			report.InSitemapWithImpressions++
			continue
		}
		report.WithoutImpressions = append(report.WithoutImpressions, url)
	}
	for url, impressions := range analyticsSet {
		if impressions <= 0 {
			continue
		}
		if _, ok := sitemapSet[url]; !ok {
			report.MissingFromSitemap = append(report.MissingFromSitemap, url)
		}
	}
	sort.Strings(report.WithoutImpressions)
	sort.Strings(report.MissingFromSitemap)
	return report
}

// analyticsByLocalHour keys rows by clock hour ("15:04"), because HOUR keys are full
// timestamps and the compared days differ.
func analyticsByLocalHour(result *gsc.AnalyticsResult) map[string]gsc.AnalyticsRow {
	rows := make(map[string]gsc.AnalyticsRow)
	if result == nil {
		return rows
	}
	for _, row := range result.Rows {
		key := strings.Join(row.Keys, "|")
		if stamp, err := time.Parse(time.RFC3339, key); err == nil {
			key = stamp.Format("15:04")
		}
		rows[key] = row
	}
	return rows
}

func analyticsByKey(result *gsc.AnalyticsResult) map[string]gsc.AnalyticsRow {
	rows := make(map[string]gsc.AnalyticsRow)
	if result == nil {
		return rows
	}
	for _, row := range result.Rows {
		rows[strings.Join(row.Keys, "|")] = row
	}
	return rows
}

func totals(result *gsc.AnalyticsResult) (clicks, impressions, ctr, position float64) {
	if result == nil {
		return
	}
	for _, row := range result.Rows {
		clicks += row.Clicks
		impressions += row.Impressions
		ctr += row.CTR * row.Impressions
		position += row.Position * row.Impressions
	}
	if impressions > 0 {
		ctr /= impressions
		position /= impressions
	}
	return
}

func classifyCause(currentImpr, baselineImpr, currentCTR, baselineCTR, currentPos, baselinePos float64) string {
	positionStable := math.Abs(currentPos-baselinePos) < 0.5
	impressionsStable := math.Abs(currentImpr-baselineImpr) <= math.Max(1, baselineImpr*0.05)
	if currentImpr < baselineImpr && positionStable {
		return "demand_decline"
	}
	if currentImpr < baselineImpr && currentPos > baselinePos+0.5 {
		return "ranking_loss"
	}
	if impressionsStable && currentCTR < baselineCTR {
		return "ctr_drop"
	}
	return "mixed"
}

func buildCTRCurve(result *gsc.AnalyticsResult, minImpressions float64) map[int]float64 {
	buckets := make(map[int][]float64)
	for _, row := range result.Rows {
		if row.Impressions < minImpressions {
			continue
		}
		bucket := int(math.Round(row.Position))
		buckets[bucket] = append(buckets[bucket], row.CTR)
	}
	curve := make(map[int]float64, len(buckets))
	for bucket, values := range buckets {
		sort.Float64s(values)
		curve[bucket] = values[len(values)/2]
	}
	return curve
}

func splitQueryPage(dimensions, keys []string) (string, string) {
	var query, page string
	for i, dimension := range dimensions {
		if i >= len(keys) {
			break
		}
		switch dimension {
		case "QUERY":
			query = keys[i]
		case "PAGE":
			page = keys[i]
		}
	}
	return query, page
}
