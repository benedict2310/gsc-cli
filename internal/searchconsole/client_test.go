package searchconsole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	gscapi "google.golang.org/api/searchconsole/v1"
)

func TestQueryAnalyticsShapesRequest(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
		gotBody   gscapi.SearchAnalyticsQueryRequest
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responseAggregationType":"byPage","rows":[{"keys":["q","/"],"clicks":2,"impressions":10,"ctr":0.2,"position":3.5}]}`))
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{
		HTTPClient: server.Client(),
		Endpoint:   server.URL + "/",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := &gscapi.SearchAnalyticsQueryRequest{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-03",
		Dimensions: []string{"QUERY", "PAGE"},
		SearchType: "WEB",
		RowLimit:   100,
	}
	result, err := client.QueryAnalytics(context.Background(), "sc-domain:example.com", req)
	if err != nil {
		t.Fatalf("QueryAnalytics() error = %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %s, want POST", gotMethod)
	}
	if !strings.Contains(gotPath, "/sites/sc-domain:example.com/searchAnalytics/query") {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody.StartDate != "2026-03-01" || gotBody.RowLimit != 100 {
		t.Fatalf("request body = %#v", gotBody)
	}
	if len(result.Rows) != 1 || result.Rows[0].Clicks != 2 {
		t.Fatalf("result = %#v", result)
	}
}

func TestSitemapsMutationsShapeRequests(t *testing.T) {
	t.Parallel()

	var methods []string
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"path":"https://example.com/sitemap.xml","warnings":"0","errors":"0"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{HTTPClient: server.Client(), Endpoint: server.URL + "/"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.GetSitemap(context.Background(), "sc-domain:example.com", "https://example.com/sitemap.xml"); err != nil {
		t.Fatalf("GetSitemap() error = %v", err)
	}
	if err := client.SubmitSitemap(context.Background(), "sc-domain:example.com", "https://example.com/sitemap.xml"); err != nil {
		t.Fatalf("SubmitSitemap() error = %v", err)
	}
	if err := client.DeleteSitemap(context.Background(), "sc-domain:example.com", "https://example.com/sitemap.xml"); err != nil {
		t.Fatalf("DeleteSitemap() error = %v", err)
	}
	if len(methods) != 3 || methods[0] != http.MethodGet || methods[1] != http.MethodPut || methods[2] != http.MethodDelete {
		t.Fatalf("methods = %v", methods)
	}
	for _, path := range paths {
		if !strings.Contains(path, "/sites/sc-domain:example.com/sitemaps/") {
			t.Fatalf("path = %q", path)
		}
	}
}

func TestInspectURLShapesRequest(t *testing.T) {
	t.Parallel()

	var gotBody gscapi.InspectUrlIndexRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/urlInspection/index:inspect") {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"inspectionResult":{"inspectionResultLink":"https://search.google.com/test","indexStatusResult":{"verdict":"PASS","coverageState":"Submitted and indexed","robotsTxtState":"ALLOWED","pageFetchState":"SUCCESSFUL"}}}`))
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{
		HTTPClient: server.Client(),
		Endpoint:   server.URL + "/",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := client.InspectURL(context.Background(), &gscapi.InspectUrlIndexRequest{
		SiteUrl:       "sc-domain:example.com",
		InspectionUrl: "https://example.com/",
		LanguageCode:  "en-US",
	})
	if err != nil {
		t.Fatalf("InspectURL() error = %v", err)
	}

	if gotBody.SiteUrl != "sc-domain:example.com" || gotBody.InspectionUrl != "https://example.com/" {
		t.Fatalf("request body = %#v", gotBody)
	}
	if result.Verdict != "PASS" || result.PageFetchState != "SUCCESSFUL" {
		t.Fatalf("result = %#v", result)
	}
}

func TestQueryAnalyticsAllPaginatesAndRespectsMaxRows(t *testing.T) {
	t.Parallel()

	var startRows []int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var gotBody gscapi.SearchAnalyticsQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		startRows = append(startRows, gotBody.StartRow)
		w.Header().Set("Content-Type", "application/json")
		if gotBody.StartRow == 0 {
			_, _ = w.Write([]byte(`{"rows":[` + strings.Repeat(`{"keys":["page"],"clicks":1,"impressions":10,"ctr":0.1,"position":3},`, 24999) + `{"keys":["page"],"clicks":1,"impressions":10,"ctr":0.1,"position":3}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"rows":[{"keys":["page-25000"],"clicks":2,"impressions":20,"ctr":0.1,"position":4}]}`))
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{HTTPClient: server.Client(), Endpoint: server.URL + "/"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := &gscapi.SearchAnalyticsQueryRequest{StartDate: "2026-03-01", EndDate: "2026-03-03", Dimensions: []string{"PAGE"}, RowLimit: 10}
	result, meta, err := client.QueryAnalyticsAll(context.Background(), "sc-domain:example.com", req, QueryAllOptions{MaxRows: 25001})
	if err != nil {
		t.Fatalf("QueryAnalyticsAll() error = %v", err)
	}
	if len(startRows) != 2 || startRows[0] != 0 || startRows[1] != 25000 {
		t.Fatalf("startRows = %v", startRows)
	}
	if len(result.Rows) != 25001 {
		t.Fatalf("rows = %d, want 25001", len(result.Rows))
	}
	if meta.PagesFetched != 2 || meta.RowCount != 25001 || !meta.Truncated {
		t.Fatalf("meta = %#v", meta)
	}
	if req.RowLimit != 10 || req.StartRow != 0 {
		t.Fatalf("request mutated = %#v", req)
	}
}

func TestAnalyticsRequestHashIsStable(t *testing.T) {
	t.Parallel()

	reqA := &gscapi.SearchAnalyticsQueryRequest{
		StartDate:  "2026-03-01",
		EndDate:    "2026-03-03",
		Dimensions: []string{"QUERY", "PAGE"},
		SearchType: "WEB",
		DataState:  "ALL",
		DimensionFilterGroups: []*gscapi.ApiDimensionFilterGroup{{
			GroupType: "and",
			Filters:   []*gscapi.ApiDimensionFilter{{Dimension: "country", Expression: "usa", Operator: "equals"}},
		}},
	}
	reqB := &gscapi.SearchAnalyticsQueryRequest{
		EndDate:    "2026-03-03",
		StartDate:  "2026-03-01",
		SearchType: "WEB",
		DataState:  "ALL",
		Dimensions: []string{"QUERY", "PAGE"},
		DimensionFilterGroups: []*gscapi.ApiDimensionFilterGroup{{
			Filters:   []*gscapi.ApiDimensionFilter{{Operator: "equals", Expression: "usa", Dimension: "country"}},
			GroupType: "and",
		}},
		StartRow: 99,
		RowLimit: 10,
	}

	hashA, err := AnalyticsRequestHash("sc-domain:example.com", reqA)
	if err != nil {
		t.Fatalf("AnalyticsRequestHash(reqA) error = %v", err)
	}
	hashB, err := AnalyticsRequestHash("sc-domain:example.com", reqB)
	if err != nil {
		t.Fatalf("AnalyticsRequestHash(reqB) error = %v", err)
	}
	if hashA != hashB {
		t.Fatalf("hashA = %q, hashB = %q", hashA, hashB)
	}
	if _, err := strconv.ParseUint(hashA[:8], 16, 32); err != nil {
		t.Fatalf("hashA is not hex-like: %q", hashA)
	}
}

func TestListSitesSanitizesAPIErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"code":403,"message":"forbidden\nraw detail"}}`, http.StatusForbidden)
	}))
	defer server.Close()

	client, err := New(context.Background(), Options{
		HTTPClient: server.Client(),
		Endpoint:   server.URL + "/",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.ListSites(context.Background())
	if err == nil {
		t.Fatal("ListSites() expected error")
	}
	if strings.Contains(err.Error(), "\n") {
		t.Fatalf("error contains newline: %q", err)
	}
}
