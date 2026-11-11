package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

func TestAuditSitemapAndOpportunitiesSendTheChosenDataState(t *testing.T) {
	var mu sync.Mutex
	var states []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sitemap.xml" {
			_, _ = w.Write([]byte(`<urlset><url><loc>https://example.com/</loc></url></urlset>`))
			return
		}
		var body struct {
			DataState string `json:"dataState"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		mu.Lock()
		states = append(states, body.DataState)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer server.Close()
	stubClient(t, server)
	creds := filepath.Join(t.TempDir(), "key.json")
	if err := os.WriteFile(creds, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	common := []string{"--site", "sc-domain:example.com", "--credentials-file", creds, "--period", "last-28d", "--output", "json"}

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"opportunities default", []string{"opportunities"}, "FINAL"},
		{"opportunities all", []string{"opportunities", "--data-state", "all"}, "ALL"},
		{"audit sitemap default", []string{"audit", "sitemap", "--sitemap-url", server.URL + "/sitemap.xml"}, "FINAL"},
		{"audit sitemap all", []string{"audit", "sitemap", "--sitemap-url", server.URL + "/sitemap.xml", "--data-state", "all"}, "ALL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			states = nil
			captureStdout(t, func() {
				if err := run(context.Background(), append(tc.args, common...)); err != nil {
					t.Fatalf("run() error = %v", err)
				}
			})
			if len(states) == 0 {
				t.Fatal("no Search Analytics request was sent")
			}
			for _, got := range states {
				if !strings.EqualFold(got, tc.want) {
					t.Fatalf("dataState = %v, want every request %s", states, tc.want)
				}
			}
		})
	}
}

func TestAuditSitemapAndOpportunitiesRejectHourlyData(t *testing.T) {
	for _, args := range [][]string{
		{"opportunities", "--data-state", "hourly_all"},
		{"audit", "sitemap", "--sitemap-url", "https://example.com/sitemap.xml", "--data-state", "hourly_all"},
	} {
		err := run(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "data-state must be final or all") {
			t.Fatalf("run(%v) error = %v, want data-state must be final or all", args, err)
		}
	}
}

func stubClient(t *testing.T, server *httptest.Server) {
	t.Helper()
	original := newClient
	newClient = func(ctx context.Context, opts gsc.Options) (*gsc.Client, error) {
		opts.CredentialsFile = ""
		opts.HTTPClient = server.Client()
		opts.Endpoint = server.URL + "/"
		return gsc.New(ctx, opts)
	}
	t.Cleanup(func() { newClient = original })
}
