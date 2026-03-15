package main

import (
	"context"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestRunAnalyticsQueryParsesFirstFlag(t *testing.T) {
	t.Parallel()

	err := runAnalyticsQuery(context.Background(), []string{
		"--site", "sc-domain:futurelab.studio",
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
