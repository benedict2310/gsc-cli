package searchconsole

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	gscapi "google.golang.org/api/searchconsole/v1"
)

type Options struct {
	CredentialsFile string
	HTTPClient      *http.Client
	Endpoint        string
	WriteScope      bool
}

type Client struct {
	service *gscapi.Service
}

func New(ctx context.Context, opts Options) (*Client, error) {
	if opts.HTTPClient == nil && strings.TrimSpace(opts.CredentialsFile) == "" {
		return nil, errors.New("credentials file is required")
	}

	scope := gscapi.WebmastersReadonlyScope
	if opts.WriteScope {
		scope = gscapi.WebmastersScope
	}
	clientOptions := []option.ClientOption{
		option.WithScopes(scope),
	}
	if opts.HTTPClient != nil {
		clientOptions = append(clientOptions, option.WithHTTPClient(opts.HTTPClient))
	}
	if opts.CredentialsFile != "" {
		clientOptions = append(clientOptions, option.WithCredentialsFile(opts.CredentialsFile))
	}
	if opts.Endpoint != "" {
		clientOptions = append(clientOptions, option.WithEndpoint(opts.Endpoint))
	}

	service, err := gscapi.NewService(ctx, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("create Search Console client: %w", err)
	}
	return &Client{service: service}, nil
}

func (c *Client) ListSites(ctx context.Context) ([]Site, error) {
	resp, err := c.service.Sites.List().Context(ctx).Do()
	if err != nil {
		return nil, sanitizeError("list sites", err)
	}

	result := make([]Site, 0, len(resp.SiteEntry))
	for _, site := range resp.SiteEntry {
		result = append(result, Site{
			SiteURL:         site.SiteUrl,
			PermissionLevel: site.PermissionLevel,
		})
	}
	return result, nil
}

func (c *Client) QueryAnalytics(ctx context.Context, site string, req *gscapi.SearchAnalyticsQueryRequest) (*AnalyticsResult, error) {
	resp, err := c.service.Searchanalytics.Query(site, req).Context(ctx).Do()
	if err != nil {
		return nil, sanitizeError("query analytics", err)
	}

	result := &AnalyticsResult{
		Site:                    site,
		StartDate:               req.StartDate,
		EndDate:                 req.EndDate,
		Dimensions:              append([]string(nil), req.Dimensions...),
		ResponseAggregationType: resp.ResponseAggregationType,
		Rows:                    make([]AnalyticsRow, 0, len(resp.Rows)),
	}
	for _, row := range resp.Rows {
		result.Rows = append(result.Rows, AnalyticsRow{
			Keys:        append([]string(nil), row.Keys...),
			Clicks:      row.Clicks,
			Impressions: row.Impressions,
			CTR:         row.Ctr,
			Position:    row.Position,
		})
	}
	return result, nil
}

func (c *Client) ListSitemaps(ctx context.Context, site string) ([]Sitemap, error) {
	resp, err := c.service.Sitemaps.List(site).Context(ctx).Do()
	if err != nil {
		return nil, sanitizeError("list sitemaps", err)
	}

	result := make([]Sitemap, 0, len(resp.Sitemap))
	for _, sitemap := range resp.Sitemap {
		result = append(result, toSitemap(sitemap))
	}
	return result, nil
}

func (c *Client) GetSitemap(ctx context.Context, site, feedpath string) (*Sitemap, error) {
	resp, err := c.service.Sitemaps.Get(site, feedpath).Context(ctx).Do()
	if err != nil {
		return nil, sanitizeError("get sitemap", err)
	}
	result := toSitemap(resp)
	return &result, nil
}

func (c *Client) SubmitSitemap(ctx context.Context, site, feedpath string) error {
	if err := c.service.Sitemaps.Submit(site, feedpath).Context(ctx).Do(); err != nil {
		return sanitizeError("submit sitemap", err)
	}
	return nil
}

func (c *Client) DeleteSitemap(ctx context.Context, site, feedpath string) error {
	if err := c.service.Sitemaps.Delete(site, feedpath).Context(ctx).Do(); err != nil {
		return sanitizeError("delete sitemap", err)
	}
	return nil
}

func (c *Client) InspectURL(ctx context.Context, req *gscapi.InspectUrlIndexRequest) (*InspectionResult, error) {
	resp, err := c.service.UrlInspection.Index.Inspect(req).Context(ctx).Do()
	if err != nil {
		return nil, sanitizeError("inspect url", err)
	}

	result := &InspectionResult{
		Site:          req.SiteUrl,
		InspectionURL: req.InspectionUrl,
	}
	if resp.InspectionResult == nil || resp.InspectionResult.IndexStatusResult == nil {
		return result, nil
	}

	index := resp.InspectionResult.IndexStatusResult
	result.Verdict = index.Verdict
	result.CoverageState = index.CoverageState
	result.IndexingState = index.IndexingState
	result.PageFetchState = index.PageFetchState
	result.RobotsTxtState = index.RobotsTxtState
	result.CrawledAs = index.CrawledAs
	result.LastCrawlTime = index.LastCrawlTime
	result.GoogleCanonical = index.GoogleCanonical
	result.UserCanonical = index.UserCanonical
	result.ReferringURLs = append([]string(nil), index.ReferringUrls...)
	result.Sitemaps = append([]string(nil), index.Sitemap...)
	result.InspectionResultLink = resp.InspectionResult.InspectionResultLink
	return result, nil
}

func toSitemap(sitemap *gscapi.WmxSitemap) Sitemap {
	if sitemap == nil {
		return Sitemap{}
	}
	return Sitemap{
		Path:           sitemap.Path,
		IsPending:      sitemap.IsPending,
		LastSubmitted:  sitemap.LastSubmitted,
		LastDownloaded: sitemap.LastDownloaded,
		Warnings:       sitemap.Warnings,
		Errors:         sitemap.Errors,
	}
}

func sanitizeError(action string, err error) error {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		message := strings.TrimSpace(apiErr.Message)
		if message == "" {
			message = http.StatusText(apiErr.Code)
		}
		message = strings.ReplaceAll(message, "\n", " ")
		return fmt.Errorf("%s: api returned %d %s", action, apiErr.Code, message)
	}
	return fmt.Errorf("%s: %w", action, err)
}
