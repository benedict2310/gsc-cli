package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfg "github.com/benedict2310/gsc-cli/internal/config"
	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

func TestRunAnalyticsQueryParsesFirstFlag(t *testing.T) {
	t.Parallel()

	err := runAnalyticsQuery(context.Background(), []string{
		"--site", "sc-domain:example.com",
		"--credentials-file", "/does/not/exist.json",
		"--start", "2026-03-01",
		"--end", "2026-03-14",
		"--dimensions", "query",
	})
	if err == nil {
		t.Fatal("runAnalyticsQuery() expected error")
	}
	if strings.Contains(err.Error(), "invalid start date") {
		t.Fatalf("runAnalyticsQuery() still skipped the first flag: %v", err)
	}
}

func TestRunAnalyticsHelpReturnsNil(t *testing.T) {
	t.Parallel()

	if err := run(context.Background(), []string{"analytics", "--help"}); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

func TestRunVersionRejectsUnsupportedOutput(t *testing.T) {
	t.Parallel()

	err := runVersion([]string{"--output", "yaml"})
	if err == nil {
		t.Fatal("runVersion() expected error")
	}
	if !strings.Contains(err.Error(), "supported values are json and table") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunQueryHelpReturnsFlagErrHelp(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"analytics", "query", "--help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("run() = %v", err)
	}
}

func TestRunHelpAnalyticsQueryReturnsFlagErrHelp(t *testing.T) {
	t.Parallel()

	err := runHelp([]string{"analytics", "query"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("runHelp() = %v", err)
	}
}

func TestRunHelpNewCommandGroupsReturnsNilOrFlagErrHelp(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		{"cache"},
		{"sync"},
		{"diagnose"},
		{"monitor"},
		{"audit"},
		{"sitemaps"},
		{"inspect"},
	}
	for _, args := range cases {
		args := args
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()
			err := runHelp(args)
			if err != nil && !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("runHelp(%v) = %v, want nil or flag.ErrHelp", args, err)
			}
		})
	}
}

func TestRunNewCommandGroupDashHelpDoesNotRequireCredentials(t *testing.T) {
	t.Parallel()

	groups := []string{"cache", "sync", "diagnose", "monitor", "audit", "sitemaps", "inspect"}
	for _, group := range groups {
		group := group
		t.Run(group, func(t *testing.T) {
			t.Parallel()
			err := run(context.Background(), []string{group, "--help"})
			if err != nil && !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("run(%s --help) = %v, want nil or flag.ErrHelp", group, err)
			}
		})
	}
}

func TestRunHelpNewGroupedSubcommandsReturnsFlagErrHelp(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		{"sync", "analytics"},
		{"diagnose", "drop"},
		{"monitor", "hourly"},
		{"audit", "indexability"},
		{"sitemaps", "get"},
		{"inspect", "batch"},
	}
	for _, args := range cases {
		args := args
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()
			if err := runHelp(args); !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("runHelp(%v) = %v, want flag.ErrHelp", args, err)
			}
		})
	}
}

func TestRunCacheStatusDoesNotRequireCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "cache.json")

	out := captureStdout(t, func() {
		err := run(context.Background(), []string{"cache", "status", "--cache-file", cacheFile, "--output", "json"})
		if err != nil {
			t.Fatalf("run() error = %v", err)
		}
	})

	if !bytes.Contains(out, []byte(`"analyticsRows": 0`)) {
		t.Fatalf("cache status output = %s", out)
	}
}

func TestBuildCompareReportUsesAggregateRowsForSummary(t *testing.T) {
	t.Parallel()

	currentDetail := &gsc.AnalyticsResult{
		Rows: []gsc.AnalyticsRow{{Keys: []string{"/a"}, Clicks: 2, Impressions: 20}, {Keys: []string{"/b"}, Clicks: 3, Impressions: 30}},
	}
	previousDetail := &gsc.AnalyticsResult{
		Rows: []gsc.AnalyticsRow{{Keys: []string{"/a"}, Clicks: 1, Impressions: 10}},
	}
	currentAggregate := &gsc.AnalyticsResult{
		Rows: []gsc.AnalyticsRow{{Keys: []string{"2026-03-01"}, Clicks: 10, Impressions: 100}, {Keys: []string{"2026-03-02"}, Clicks: 5, Impressions: 50}},
	}
	previousAggregate := &gsc.AnalyticsResult{
		Rows: []gsc.AnalyticsRow{{Keys: []string{"2026-02-27"}, Clicks: 4, Impressions: 40}, {Keys: []string{"2026-02-28"}, Clicks: 2, Impressions: 20}},
	}

	report := buildCompareReport(
		"sc-domain:example.com",
		"page",
		cfg.DateRange{Start: "2026-03-01", End: "2026-03-02"},
		cfg.DateRange{Start: "2026-02-27", End: "2026-02-28"},
		currentDetail,
		previousDetail,
		currentAggregate,
		previousAggregate,
	)

	if report.Summary.CurrentRows != len(currentDetail.Rows) || report.Summary.PreviousRows != len(previousDetail.Rows) {
		t.Fatalf("summary rows = %#v", report.Summary)
	}
	if report.Summary.CurrentClicks != 15 || report.Summary.CurrentImpressions != 150 {
		t.Fatalf("current summary = %#v", report.Summary)
	}
	if report.Summary.PreviousClicks != 6 || report.Summary.PreviousImpressions != 60 {
		t.Fatalf("previous summary = %#v", report.Summary)
	}
	if report.Summary.ClickDelta != 9 || report.Summary.ImpressionDelta != 90 {
		t.Fatalf("delta summary = %#v", report.Summary)
	}
	if report.Current != currentDetail || report.Previous != previousDetail {
		t.Fatalf("detail rows changed: %#v", report)
	}
}

func TestRunDoctorRedactsCredentialsPath(t *testing.T) {
	tmpDir := t.TempDir()
	credentialsFile := filepath.Join(tmpDir, "service-account.json")
	if err := os.WriteFile(credentialsFile, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	out := captureStdout(t, func() {
		err := runDoctor(context.Background(), []string{"--credentials-file", credentialsFile, "--output", "json"})
		if err != nil {
			t.Fatalf("runDoctor() error = %v", err)
		}
	})

	var report doctorReport
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	for _, check := range report.Checks {
		if check.Name == "credentials_file" {
			if check.Detail != "<provided>" {
				t.Fatalf("credentials_file detail = %q, want <provided>", check.Detail)
			}
			return
		}
	}
	t.Fatal("credentials_file check not found")
}

func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	os.Stdout = originalStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		t.Fatalf("buf.ReadFrom() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() error = %v", err)
	}
	return buf.Bytes()
}
