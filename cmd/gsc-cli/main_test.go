package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfg "github.com/benedict2310/gsc-cli/internal/config"
	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

func TestRunCommandsJSONIncludesPublicCommandsWithoutCredentials(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{"commands", "--output", "json"}); err != nil {
			t.Fatalf("run() error = %v", err)
		}
	})

	var payload struct {
		Commands []struct {
			Path               string `json:"path"`
			Safety             string `json:"safety"`
			WriteScopeRequired bool   `json:"writeScopeRequired"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(payload.Commands) == 0 {
		t.Fatal("commands payload was empty")
	}
	foundSubmit := false
	for _, command := range payload.Commands {
		if command.Path == "sitemaps submit" {
			foundSubmit = true
			if command.Safety != "write-capable" || !command.WriteScopeRequired {
				t.Fatalf("submit markers = %#v", command)
			}
		}
	}
	if !foundSubmit {
		t.Fatal("commands payload missing sitemaps submit")
	}
}

func TestRunCommandDiagnoseDropJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{"command", "diagnose", "drop", "--output", "json"}); err != nil {
			t.Fatalf("run() error = %v", err)
		}
	})

	var payload struct {
		Path                string `json:"path"`
		RequiresSite        bool   `json:"requiresSite"`
		RequiresCredentials bool   `json:"requiresCredentials"`
		Safety              string `json:"safety"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Path != "diagnose drop" {
		t.Fatalf("path = %q", payload.Path)
	}
	if !payload.RequiresSite || !payload.RequiresCredentials {
		t.Fatalf("requirements = %#v", payload)
	}
	if payload.Safety != "read-only" {
		t.Fatalf("safety = %q", payload.Safety)
	}
}

func TestRunCommandsTableIncludesAgentDiscoveryFields(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{"commands", "--output", "table"}); err != nil {
			t.Fatalf("run() error = %v", err)
		}
	})
	for _, want := range []string{"REQUIRES_CREDENTIALS", "REQUIRES_SITE", "APIS_USED", "OUTPUT_MODES", "FLAGS"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Fatalf("commands table missing %q:\n%s", want, out)
		}
	}
}

func TestRunCommandTableIncludesFlagsExamplesAndRelatedCommands(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{"command", "diagnose", "drop", "--output", "table"}); err != nil {
			t.Fatalf("run() error = %v", err)
		}
	})
	for _, want := range []string{"FLAGS", "current:string:required", "EXAMPLES", "CAVEATS", "RELATED_COMMANDS"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Fatalf("command table missing %q:\n%s", want, out)
		}
	}
}

func TestRunCommandHelpPrintsUsage(t *testing.T) {
	out := captureStderr(t, func() {
		if err := run(context.Background(), []string{"command", "--help"}); !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("run() = %v, want flag.ErrHelp", err)
		}
	})
	if !bytes.Contains(out, []byte("gsc-cli command <command-path>")) {
		t.Fatalf("help output = %s", out)
	}
}

func TestRunCommandHelpCommandPathJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run(context.Background(), []string{"command", "help", "--output", "json"}); err != nil {
			t.Fatalf("run() error = %v", err)
		}
	})

	var payload struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output=%s", err, out)
	}
	if payload.Path != "help" {
		t.Fatalf("path = %q, want help", payload.Path)
	}
}

func TestRunCommandNotFound(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"command", "nope", "--output", "json"})
	if err == nil {
		t.Fatal("run() expected error")
	}
	if !strings.Contains(err.Error(), `unknown command spec "nope"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

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

func captureStderr(t *testing.T, fn func()) []byte {
	t.Helper()

	originalStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = writer

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	os.Stderr = originalStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		t.Fatalf("buf.ReadFrom() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() error = %v", err)
	}
	return buf.Bytes()
}

func TestAlignToLatestDataEndsWindowAtLastDateWithRows(t *testing.T) {
	var gotBody struct {
		StartDate  string   `json:"startDate"`
		EndDate    string   `json:"endDate"`
		Dimensions []string `json:"dimensions"`
		DataState  string   `json:"dataState"`
	}
	rows := `{"rows":[{"keys":["2026-11-03"],"impressions":5},{"keys":["2026-11-04"],"impressions":7}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(rows))
	}))
	defer server.Close()
	client, err := gsc.New(context.Background(), gsc.Options{HTTPClient: server.Client(), Endpoint: server.URL + "/"})
	if err != nil {
		t.Fatalf("gsc.New() error = %v", err)
	}

	window := cfg.DateRange{Start: "2026-10-10", End: "2026-11-06"}
	got, err := alignToLatestData(context.Background(), client, "sc-domain:example.com", window, "web", "final")
	if err != nil {
		t.Fatalf("alignToLatestData() error = %v", err)
	}
	if got != (cfg.DateRange{Start: "2026-10-08", End: "2026-11-04"}) {
		t.Fatalf("alignToLatestData() = %#v", got)
	}
	if gotBody.EndDate != "2026-11-06" || gotBody.StartDate != "2026-10-24" || len(gotBody.Dimensions) != 1 || gotBody.Dimensions[0] != "DATE" || gotBody.DataState != "FINAL" {
		t.Fatalf("probe request = %#v", gotBody)
	}

	rows = `{}`
	got, err = alignToLatestData(context.Background(), client, "sc-domain:example.com", window, "web", "final")
	if err != nil || got != window {
		t.Fatalf("alignToLatestData() without rows = %#v, %v; want unchanged", got, err)
	}

	got, err = alignToLatestData(context.Background(), client, "sc-domain:example.com", window, "web", "hourly_all")
	if err != nil || got != window {
		t.Fatalf("alignToLatestData() for hourly_all = %#v, %v; want unchanged", got, err)
	}
}
