package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommonResolveUsesEnvFallbacks(t *testing.T) {
	t.Setenv(envCredentialsFile, filepath.Join(t.TempDir(), "service-account.json"))
	t.Setenv(envSiteURL, "sc-domain:example.com")
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
		Site:            "sc-domain:example.com",
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

func TestBuildAnalyticsRequestRejectsHourCombinedWithOtherDimensions(t *testing.T) {
	_, err := BuildAnalyticsRequest(AnalyticsInput{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-01",
		Dimensions: "date,hour",
		SearchType: "web",
		DataState:  "hourly_all",
		Limit:      10,
	})
	if err == nil {
		t.Fatal("BuildAnalyticsRequest() expected combined hour dimension error")
	}
	if !strings.Contains(err.Error(), `dimension "HOUR" cannot be combined`) {
		t.Fatalf("unexpected error: %v", err)
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

func TestBuildAnalyticsRequestNormalizesSearchAppearanceAliases(t *testing.T) {
	t.Parallel()

	for _, dimensions := range []string{"searchAppearance", "search_appearance", "search-appearance"} {
		dimensions := dimensions
		t.Run(dimensions, func(t *testing.T) {
			t.Parallel()

			req, err := BuildAnalyticsRequest(AnalyticsInput{
				StartDate:  "2026-03-01",
				EndDate:    "2026-03-02",
				Dimensions: dimensions,
				SearchType: "web",
				DataState:  "final",
				Limit:      25,
			})
			if err != nil {
				t.Fatalf("BuildAnalyticsRequest() error = %v", err)
			}
			if len(req.Dimensions) != 1 || req.Dimensions[0] != "SEARCH_APPEARANCE" {
				t.Fatalf("Dimensions = %#v", req.Dimensions)
			}
		})
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
		Site: "sc-domain:example.com",
		URL:  "example.com",
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

func TestResolveDateRangeWithDaysRejectsMixedInputs(t *testing.T) {
	_, err := ResolveDateRangeWithDays("2026-03-01", "2026-03-14", 28)
	if err == nil {
		t.Fatal("ResolveDateRangeWithDays() expected mixed-input error")
	}
}

func TestResolveDateRangeWithDaysRejectsUnsupportedWindow(t *testing.T) {
	_, err := ResolveDateRangeWithDays("", "", 30)
	if err == nil {
		t.Fatal("ResolveDateRangeWithDays() expected unsupported days error")
	}
}

func TestInferInspectionURLFromDomainProperty(t *testing.T) {
	got, err := InferInspectionURL("sc-domain:example.com")
	if err != nil {
		t.Fatalf("InferInspectionURL() error = %v", err)
	}
	if got != "https://example.com/" {
		t.Fatalf("InferInspectionURL() = %q", got)
	}
}

func TestParseDayWindow(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "empty", input: "", want: 0},
		{name: "seven", input: "7", want: 7},
		{name: "twenty eight", input: "28", want: 28},
		{name: "invalid", input: "30", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDayWindow(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseDayWindow() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDayWindow() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseDayWindow() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParsePeriodRange(t *testing.T) {
	got, err := ParsePeriodRange("2026-03-01:2026-03-14")
	if err != nil {
		t.Fatalf("ParsePeriodRange() error = %v", err)
	}
	if got.Start != "2026-03-01" || got.End != "2026-03-14" {
		t.Fatalf("ParsePeriodRange() = %#v", got)
	}
}

func TestResolveBaseline(t *testing.T) {
	current := DateRange{Start: "2026-03-01", End: "2026-03-14"}

	previous, err := ResolveBaseline(current, "previous-period")
	if err != nil {
		t.Fatalf("ResolveBaseline(previous-period) error = %v", err)
	}
	if previous.Start != "2026-02-15" || previous.End != "2026-02-28" {
		t.Fatalf("previous = %#v", previous)
	}

	yoy, err := ResolveBaseline(current, "same-period-last-year")
	if err != nil {
		t.Fatalf("ResolveBaseline(same-period-last-year) error = %v", err)
	}
	if yoy.Start != "2025-03-01" || yoy.End != "2025-03-14" {
		t.Fatalf("yoy = %#v", yoy)
	}
}

func TestResolveNamedPeriodLastDays(t *testing.T) {
	got, err := ResolveNamedPeriod("last-7d", time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ResolveNamedPeriod() error = %v", err)
	}
	if got.Start != "2026-03-08" || got.End != "2026-03-14" {
		t.Fatalf("ResolveNamedPeriod() = %#v", got)
	}
}

func TestResolveNamedPeriodLastMonths(t *testing.T) {
	cases := []struct {
		period, now, start, end string
	}{
		{"last-1m", "2026-11-07", "2026-10-07", "2026-11-06"},
		{"last-16m", "2026-11-07", "2025-07-07", "2026-11-06"},
		{"last-1m", "2026-04-01", "2026-03-01", "2026-03-31"},
		{"last-1m", "2026-03-31", "2026-03-01", "2026-03-30"},
	}
	for _, tc := range cases {
		now, _ := time.Parse("2006-01-02", tc.now)
		got, err := ResolveNamedPeriod(tc.period, now.Add(12*time.Hour))
		if err != nil {
			t.Fatalf("ResolveNamedPeriod(%s, %s) error = %v", tc.period, tc.now, err)
		}
		if got.Start != tc.start || got.End != tc.end {
			t.Fatalf("ResolveNamedPeriod(%s, %s) = %#v, want %s to %s", tc.period, tc.now, got, tc.start, tc.end)
		}
	}
}

func TestPreviousDateRange(t *testing.T) {
	got, err := PreviousDateRange(DateRange{
		Start: "2026-02-15",
		End:   "2026-03-14",
	})
	if err != nil {
		t.Fatalf("PreviousDateRange() error = %v", err)
	}
	if got.Start != "2026-01-18" || got.End != "2026-02-14" {
		t.Fatalf("PreviousDateRange() = %#v", got)
	}
}

func TestEndRangeAtShiftsWindowKeepingLength(t *testing.T) {
	t.Parallel()

	window := DateRange{Start: "2026-10-10", End: "2026-11-06"}
	got, err := EndRangeAt(window, "2026-11-04")
	if err != nil {
		t.Fatalf("EndRangeAt() error = %v", err)
	}
	if got != (DateRange{Start: "2026-10-08", End: "2026-11-04"}) {
		t.Fatalf("EndRangeAt() = %#v", got)
	}
	for _, end := range []string{"", "2026-11-06", "2026-11-07"} {
		got, err := EndRangeAt(window, end)
		if err != nil || got != window {
			t.Fatalf("EndRangeAt(%q) = %#v, %v; want unchanged", end, got, err)
		}
	}
}
