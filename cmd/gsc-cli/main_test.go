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
