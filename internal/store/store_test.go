package store

import (
	"path/filepath"
	"testing"

	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
	gscapi "google.golang.org/api/searchconsole/v1"
)

func TestStoreSaveAnalyticsAndStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cache := New(path)

	req := &gscapi.SearchAnalyticsQueryRequest{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-02",
		Dimensions: []string{"PAGE", "QUERY"},
		SearchType: "WEB",
		DataState:  "FINAL",
	}
	result := &gsc.AnalyticsResult{
		Site:       "sc-domain:example.com",
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-02",
		Dimensions: []string{"PAGE", "QUERY"},
		Rows: []gsc.AnalyticsRow{{
			Keys:        []string{"https://example.com/a", "alpha"},
			Clicks:      3,
			Impressions: 30,
			CTR:         0.1,
			Position:    4,
		}},
	}
	if err := cache.SaveAnalytics("abc", req, result, gsc.QueryAllMetadata{PagesFetched: 1, RowCount: 1}, false); err != nil {
		t.Fatalf("SaveAnalytics() error = %v", err)
	}

	status, err := cache.Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.AnalyticsRows != 1 || status.PullLogs != 1 {
		t.Fatalf("status = %#v", status)
	}

	rows, err := cache.FindAnalytics(AnalyticsFilter{Site: "sc-domain:example.com", Page: "https://example.com/a"})
	if err != nil {
		t.Fatalf("FindAnalytics() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Query != "alpha" {
		t.Fatalf("rows = %#v", rows)
	}
}

func TestStoreUpsertsProvisionalRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cache := New(path)

	req := &gscapi.SearchAnalyticsQueryRequest{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-01",
		Dimensions: []string{"HOUR"},
		SearchType: "WEB",
		DataState:  "HOURLY_ALL",
	}
	result := &gsc.AnalyticsResult{
		Site:       "sc-domain:example.com",
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-01",
		Dimensions: []string{"HOUR"},
		Rows:       []gsc.AnalyticsRow{{Keys: []string{"13"}, Clicks: 1, Impressions: 10, CTR: 0.1, Position: 2}},
	}
	if err := cache.SaveAnalytics("hourly", req, result, gsc.QueryAllMetadata{PagesFetched: 1, RowCount: 1}, true); err != nil {
		t.Fatalf("SaveAnalytics(provisional) error = %v", err)
	}
	result.Rows[0].Clicks = 2
	result.Rows[0].Impressions = 12
	req.DataState = "FINAL"
	if err := cache.SaveAnalytics("hourly", req, result, gsc.QueryAllMetadata{PagesFetched: 1, RowCount: 1}, false); err != nil {
		t.Fatalf("SaveAnalytics(final) error = %v", err)
	}

	rows, err := cache.FindAnalytics(AnalyticsFilter{Site: "sc-domain:example.com", Hour: "13"})
	if err != nil {
		t.Fatalf("FindAnalytics() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Clicks != 2 || rows[0].Provisional {
		t.Fatalf("rows = %#v", rows)
	}
}

func TestStoreSaveInspection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cache := New(path)
	if err := cache.SaveInspection(gsc.InspectionResult{Site: "sc-domain:example.com", InspectionURL: "https://example.com/", Verdict: "PASS"}, []string{"ok"}); err != nil {
		t.Fatalf("SaveInspection() error = %v", err)
	}
	status, err := cache.Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Inspections != 1 {
		t.Fatalf("status = %#v", status)
	}
}
