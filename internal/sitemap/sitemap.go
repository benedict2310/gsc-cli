package sitemap

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
)

type urlSet struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

type sitemapIndex struct {
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

func Fetch(client *http.Client, url string) ([]string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	seen := map[string]struct{}{}
	return fetch(client, url, seen)
}

func fetch(client *http.Client, url string, seen map[string]struct{}) ([]string, error) {
	if _, ok := seen[url]; ok {
		return nil, nil
	}
	seen[url] = struct{}{}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch sitemap: %w", err)
	}
	defer resp.Body.Close()
	blob, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read sitemap: %w", err)
	}
	blob, err = maybeGunzip(blob)
	if err != nil {
		return nil, err
	}
	var index sitemapIndex
	if err := xml.Unmarshal(blob, &index); err == nil && len(index.Sitemaps) > 0 {
		var urls []string
		for _, child := range index.Sitemaps {
			childURLs, err := fetch(client, child.Loc, seen)
			if err != nil {
				return nil, err
			}
			urls = append(urls, childURLs...)
		}
		sort.Strings(urls)
		return urls, nil
	}
	var set urlSet
	if err := xml.Unmarshal(blob, &set); err != nil {
		return nil, fmt.Errorf("parse sitemap: %w", err)
	}
	urls := make([]string, 0, len(set.URLs))
	for _, entry := range set.URLs {
		urls = append(urls, entry.Loc)
	}
	sort.Strings(urls)
	return urls, nil
}

func maybeGunzip(blob []byte) ([]byte, error) {
	if len(blob) < 2 || blob[0] != 0x1f || blob[1] != 0x8b {
		return blob, nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(blob))
	if err != nil {
		return nil, fmt.Errorf("open gzip sitemap: %w", err)
	}
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read gzip sitemap: %w", err)
	}
	return decoded, nil
}
