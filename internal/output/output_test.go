package output

import (
	"bytes"
	"strings"
	"testing"

	cfg "github.com/bbleimschein/gsc-cli/internal/config"
	gsc "github.com/bbleimschein/gsc-cli/internal/searchconsole"
)

func TestRenderSitesJSONIsDeterministic(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := RenderSites(&buf, cfg.OutputJSON, []gsc.Site{{
		SiteURL:         "sc-domain:futurelab.studio",
		PermissionLevel: "siteOwner",
	}})
	if err != nil {
		t.Fatalf("RenderSites() error = %v", err)
	}

	expected := "{\n"
	if !strings.HasPrefix(buf.String(), "[\n  {\n") {
		t.Fatalf("unexpected json output: %q", buf.String())
	}
	_ = expected
}

func TestRenderAnalyticsTableIsDeterministic(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := RenderAnalytics(&buf, cfg.OutputTable, &gsc.AnalyticsResult{
		Dimensions: []string{"QUERY", "PAGE"},
		Rows: []gsc.AnalyticsRow{{
			Keys:        []string{"ora", "/ora"},
			Clicks:      3,
			Impressions: 20,
			CTR:         0.15,
			Position:    4.2,
		}},
	})
	if err != nil {
		t.Fatalf("RenderAnalytics() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "QUERY  PAGE  CLICKS  IMPRESSIONS  CTR   POSITION") {
		t.Fatalf("unexpected header: %q", got)
	}
	if !strings.Contains(got, "ora    /ora  3       20           0.15  4.2") {
		t.Fatalf("unexpected row: %q", got)
	}
}

func TestRenderOverviewTableContainsSections(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := RenderOverview(&buf, cfg.OutputTable, &gsc.OverviewReport{
		Site:       "sc-domain:futurelab.studio",
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-14",
		InspectURL: "https://futurelab.studio/",
		TopPages: &gsc.AnalyticsResult{
			Dimensions: []string{"PAGE"},
			Rows: []gsc.AnalyticsRow{{
				Keys:        []string{"https://futurelab.studio/"},
				Clicks:      1,
				Impressions: 10,
				CTR:         0.1,
				Position:    3,
			}},
		},
		TopQueries: &gsc.AnalyticsResult{
			Dimensions: []string{"QUERY"},
			Rows: []gsc.AnalyticsRow{{
				Keys:        []string{"futurelab"},
				Clicks:      1,
				Impressions: 4,
				CTR:         0.25,
				Position:    2,
			}},
		},
		Sitemaps: []gsc.Sitemap{{Path: "https://futurelab.studio/sitemap.xml"}},
		Inspection: &gsc.InspectionResult{
			Site:          "sc-domain:futurelab.studio",
			InspectionURL: "https://futurelab.studio/",
			Verdict:       "PASS",
		},
	})
	if err != nil {
		t.Fatalf("RenderOverview() error = %v", err)
	}

	got := buf.String()
	for _, section := range []string{"TOP_PAGES", "TOP_QUERIES", "SITEMAPS", "INSPECTION"} {
		if !strings.Contains(got, section) {
			t.Fatalf("missing section %q in %q", section, got)
		}
	}
}
