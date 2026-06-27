package insights

import (
	"testing"

	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

func TestDiagnoseDropRanksLossContributorsAndLabelsCause(t *testing.T) {
	current := &gsc.AnalyticsResult{Dimensions: []string{"PAGE"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"/a"}, Clicks: 60, Impressions: 1000, CTR: 0.06, Position: 3}, {Keys: []string{"/b"}, Clicks: 80, Impressions: 1000, CTR: 0.08, Position: 4}}}
	baseline := &gsc.AnalyticsResult{Dimensions: []string{"PAGE"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"/a"}, Clicks: 100, Impressions: 1000, CTR: 0.1, Position: 3}, {Keys: []string{"/b"}, Clicks: 90, Impressions: 1000, CTR: 0.09, Position: 4}}}

	report := DiagnoseDrop(current, baseline, DiagnoseOptions{Segment: "page", MinLoss: 5, Limit: 5})
	if report.TotalLoss >= 0 {
		t.Fatalf("TotalLoss = %v", report.TotalLoss)
	}
	if len(report.Contributors) == 0 || report.Contributors[0].Key != "/a" {
		t.Fatalf("Contributors = %#v", report.Contributors)
	}
	if report.LikelyCause != "ctr_drop" {
		t.Fatalf("LikelyCause = %q", report.LikelyCause)
	}
}

func TestDiagnoseDropIncludesRowsMissingFromCurrentPeriod(t *testing.T) {
	current := &gsc.AnalyticsResult{Dimensions: []string{"PAGE"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"/still-there"}, Clicks: 90, Impressions: 1000, CTR: 0.09, Position: 4}}}
	baseline := &gsc.AnalyticsResult{Dimensions: []string{"PAGE"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"/gone"}, Clicks: 100, Impressions: 1000, CTR: 0.1, Position: 3}, {Keys: []string{"/still-there"}, Clicks: 100, Impressions: 1000, CTR: 0.1, Position: 4}}}

	report := DiagnoseDrop(current, baseline, DiagnoseOptions{Segment: "page", MinLoss: 5, Limit: 5})
	if len(report.Contributors) == 0 || report.Contributors[0].Key != "/gone" {
		t.Fatalf("Contributors = %#v", report.Contributors)
	}
	if report.Contributors[0].CurrentClicks != 0 || report.Contributors[0].ExpectedClicks != 100 {
		t.Fatalf("missing row contribution = %#v", report.Contributors[0])
	}
}

func TestBuildHourlyComparison(t *testing.T) {
	current := &gsc.AnalyticsResult{Dimensions: []string{"HOUR"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"09"}, Clicks: 10, Impressions: 100, CTR: 0.1}, {Keys: []string{"10"}, Clicks: 12, Impressions: 120, CTR: 0.1}}}
	baseline := &gsc.AnalyticsResult{Dimensions: []string{"HOUR"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"09"}, Clicks: 20, Impressions: 100, CTR: 0.2}, {Keys: []string{"10"}, Clicks: 12, Impressions: 100, CTR: 0.12}}}

	report := BuildHourlyComparison(current, baseline)
	if !report.Provisional {
		t.Fatal("Provisional = false, want true")
	}
	if len(report.Rows) != 2 || report.Rows[0].Hour != "09" || report.Rows[0].ClickDelta != -10 {
		t.Fatalf("Rows = %#v", report.Rows)
	}
}

func TestClassifyInspectionIssues(t *testing.T) {
	issues := ClassifyInspectionIssues(gsc.InspectionResult{RobotsTxtState: "DISALLOWED", UserCanonical: "https://example.com/a", GoogleCanonical: "https://example.com/b"})
	if len(issues) < 2 || issues[0] != "robots_blocked" {
		t.Fatalf("issues = %#v", issues)
	}
}

func TestFindOpportunitiesReturnsNearWinAndCTRGap(t *testing.T) {
	result := &gsc.AnalyticsResult{Dimensions: []string{"QUERY", "PAGE"}, Rows: []gsc.AnalyticsRow{
		{Keys: []string{"alpha", "/a"}, Clicks: 10, Impressions: 1000, CTR: 0.01, Position: 8},
		{Keys: []string{"beta", "/b"}, Clicks: 50, Impressions: 1000, CTR: 0.05, Position: 3},
		{Keys: []string{"gamma", "/c"}, Clicks: 40, Impressions: 1000, CTR: 0.04, Position: 3},
	}}
	report := FindOpportunities(result, nil, OpportunityOptions{Types: []string{"near-win", "ctr-gap"}, MinImpressions: 100, Limit: 10})
	if len(report.Items) < 2 {
		t.Fatalf("Items = %#v", report.Items)
	}
	if report.Items[0].Type != "near-win" {
		t.Fatalf("first type = %q", report.Items[0].Type)
	}
	foundCTRGap := false
	for _, item := range report.Items {
		if item.Type == "ctr-gap" {
			foundCTRGap = true
		}
	}
	if !foundCTRGap {
		t.Fatalf("expected ctr-gap item in %#v", report.Items)
	}
}

func TestFindOpportunitiesReturnsDecayWhenPreviousClicksWereHigher(t *testing.T) {
	current := &gsc.AnalyticsResult{Dimensions: []string{"QUERY", "PAGE"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"alpha", "/a"}, Clicks: 20, Impressions: 1000, CTR: 0.02, Position: 9}}}
	previous := &gsc.AnalyticsResult{Dimensions: []string{"QUERY", "PAGE"}, Rows: []gsc.AnalyticsRow{{Keys: []string{"alpha", "/a"}, Clicks: 100, Impressions: 1000, CTR: 0.1, Position: 4}}}

	report := FindOpportunities(current, previous, OpportunityOptions{Types: []string{"decay"}, MinImpressions: 100, Limit: 10})
	if len(report.Items) != 1 {
		t.Fatalf("Items = %#v", report.Items)
	}
	if report.Items[0].Type != "decay" || report.Items[0].Score != 120 {
		t.Fatalf("decay item = %#v", report.Items[0])
	}
}

func TestAuditSitemap(t *testing.T) {
	urls := []string{"https://example.com/a", "https://example.com/b"}
	analytics := []gsc.AnalyticsRow{{Keys: []string{"https://example.com/a"}, Impressions: 10}, {Keys: []string{"https://example.com/c"}, Impressions: 5}}
	report := AuditSitemap(urls, &gsc.AnalyticsResult{Dimensions: []string{"PAGE"}, Rows: analytics})
	if len(report.MissingFromSitemap) != 1 || report.MissingFromSitemap[0] != "https://example.com/c" {
		t.Fatalf("report = %#v", report)
	}
	if len(report.WithoutImpressions) != 1 || report.WithoutImpressions[0] != "https://example.com/b" {
		t.Fatalf("report = %#v", report)
	}
}
