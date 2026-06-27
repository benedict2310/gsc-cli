package sitemap

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseSitemapIndexRecursively(t *testing.T) {
	var baseURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.xml":
			_, _ = w.Write([]byte(`<?xml version="1.0"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><sitemap><loc>` + baseURL + `/child.xml</loc></sitemap></sitemapindex>`))
		case "/child.xml":
			_, _ = w.Write([]byte(`<?xml version="1.0"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://example.com/a</loc></url><url><loc>https://example.com/b</loc></url></urlset>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	baseURL = server.URL

	got, err := Fetch(server.Client(), server.URL+"/index.xml")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	want := []string{"https://example.com/a", "https://example.com/b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Fetch() = %#v, want %#v", got, want)
	}
}

func TestFetchReadsGzipCompressedSitemap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		var compressed bytes.Buffer
		writer := gzip.NewWriter(&compressed)
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://example.com/a</loc></url></urlset>`))
		_ = writer.Close()
		_, _ = w.Write(compressed.Bytes())
	}))
	defer server.Close()

	got, err := Fetch(server.Client(), server.URL+"/sitemap.xml.gz")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	want := []string{"https://example.com/a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Fetch() = %#v, want %#v", got, want)
	}
}
