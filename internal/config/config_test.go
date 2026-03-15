package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommonResolveUsesEnvFallbacks(t *testing.T) {
	t.Setenv(envCredentialsFile, filepath.Join(t.TempDir(), "service-account.json"))
	t.Setenv(envSiteURL, "sc-domain:futurelab.studio")
	if err := os.WriteFile(os.Getenv(envCredentialsFile), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write credentials file: %v", err)
	}

	cfg := &Common{Output: OutputTable}
	if err := cfg.Resolve(true); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if cfg.CredentialsFile == "" || cfg.Site == "" {
		t.Fatalf("Resolve() did not populate credentials/site")
	}
}

func TestCommonResolveRejectsUnsupportedOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service-account.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write credentials file: %v", err)
	}
	cfg := &Common{
		CredentialsFile: path,
		Site:            "sc-domain:futurelab.studio",
		Output:          "yaml",
	}
	if err := cfg.Resolve(true); err == nil {
		t.Fatal("Resolve() expected error for unsupported output mode")
	}
}

func TestBuildAnalyticsRequestValidatesInput(t *testing.T) {
	_, err := BuildAnalyticsRequest(AnalyticsInput{
		StartDate:  "2026-03-02",
		EndDate:    "2026-03-01",
		Dimensions: "query",
		SearchType: "web",
		DataState:  "final",
		Limit:      10,
	})
	if err == nil {
		t.Fatal("BuildAnalyticsRequest() expected date range error")
	}
}

func TestBuildAnalyticsRequestRejectsUnsupportedDimension(t *testing.T) {
	_, err := BuildAnalyticsRequest(AnalyticsInput{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-02",
		Dimensions: "query,author",
		SearchType: "web",
		DataState:  "final",
		Limit:      10,
	})
	if err == nil {
		t.Fatal("BuildAnalyticsRequest() expected unsupported dimension error")
	}
}

func TestBuildAnalyticsRequestRequiresHourlyAllForHour(t *testing.T) {
	_, err := BuildAnalyticsRequest(AnalyticsInput{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-02",
		Dimensions: "hour",
		SearchType: "web",
		DataState:  "final",
		Limit:      10,
	})
	if err == nil {
		t.Fatal("BuildAnalyticsRequest() expected hourly_all requirement")
	}
}

func TestBuildAnalyticsRequestNormalizesValues(t *testing.T) {
	req, err := BuildAnalyticsRequest(AnalyticsInput{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-02",
		Dimensions: "query, page",
		SearchType: "google_news",
		DataState:  "all",
		Limit:      25,
		StartRow:   4,
	})
	if err != nil {
		t.Fatalf("BuildAnalyticsRequest() error = %v", err)
	}

	if got, want := req.SearchType, "GOOGLE_NEWS"; got != want {
		t.Fatalf("SearchType = %q, want %q", got, want)
	}
	if got, want := req.DataState, "ALL"; got != want {
		t.Fatalf("DataState = %q, want %q", got, want)
	}
	if len(req.Dimensions) != 2 || req.Dimensions[0] != "QUERY" || req.Dimensions[1] != "PAGE" {
		t.Fatalf("Dimensions = %#v", req.Dimensions)
	}
}

func TestBuildInspectRequestRejectsBadURL(t *testing.T) {
	_, err := BuildInspectRequest(InspectInput{
		Site: "sc-domain:futurelab.studio",
		URL:  "futurelab.studio",
	})
	if err == nil {
		t.Fatal("BuildInspectRequest() expected URL validation error")
	}
}

func TestResolveDateRangeDefaultsToRecentWindow(t *testing.T) {
	got, err := ResolveDateRange("", "")
	if err != nil {
		t.Fatalf("ResolveDateRange() error = %v", err)
	}
	if got.Start == "" || got.End == "" {
		t.Fatalf("ResolveDateRange() = %#v", got)
	}
}

func TestResolveDateRangeRequiresBothDates(t *testing.T) {
	_, err := ResolveDateRange("2026-03-01", "")
	if err == nil {
		t.Fatal("ResolveDateRange() expected error for partial range")
	}
}
