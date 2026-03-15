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
