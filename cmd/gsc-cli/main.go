package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	cfg "github.com/bbleimschein/gsc-cli/internal/config"
	"github.com/bbleimschein/gsc-cli/internal/output"
	gsc "github.com/bbleimschein/gsc-cli/internal/searchconsole"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		printUsage()
		return errors.New("missing command")
	}

	switch args[0] {
	case "sites":
		return runSites(ctx, args[1:])
	case "analytics":
		return runAnalytics(ctx, args[1:])
	case "report":
		return runReport(ctx, args[1:])
	case "sitemaps":
		return runSitemaps(ctx, args[1:])
	case "inspect":
		return runInspect(ctx, args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runSites(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "list" {
		return errors.New("usage: gsc-cli sites list [flags]")
	}

	fs := flag.NewFlagSet("sites list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(false); err != nil {
		return err
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}

	result, err := client.ListSites(ctx)
	if err != nil {
		return err
	}
	return output.RenderSites(os.Stdout, base.Output, result)
}

func runAnalytics(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: gsc-cli analytics <query|top-pages|top-queries> [flags]")
	}
	switch args[0] {
	case "query":
		return runAnalyticsQuery(ctx, args[1:])
	case "top-pages":
		return runAnalyticsTopPages(ctx, args[1:])
	case "top-queries":
		return runAnalyticsTopQueries(ctx, args[1:])
	default:
		return errors.New("usage: gsc-cli analytics <query|top-pages|top-queries> [flags]")
	}
}

func runAnalyticsQuery(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("analytics query", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	var (
		startDate  = fs.String("start", "", "start date in YYYY-MM-DD")
		endDate    = fs.String("end", "", "end date in YYYY-MM-DD")
		dimensions = fs.String("dimensions", "query,page", "comma-separated dimensions")
		searchType = fs.String("search-type", "web", "search type: web,image,video,news,discover,google_news")
		dataState  = fs.String("data-state", "final", "data state: final,all,hourly_all")
		limit      = fs.Int64("limit", 1000, "maximum rows to return (1-25000)")
		startRow   = fs.Int64("start-row", 0, "zero-based row offset")
	)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}

	req, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  *startDate,
		EndDate:    *endDate,
		Dimensions: *dimensions,
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      *limit,
		StartRow:   *startRow,
	})
	if err != nil {
		return err
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}

	result, err := client.QueryAnalytics(ctx, base.Site, req)
	if err != nil {
		return err
	}
	return output.RenderAnalytics(os.Stdout, base.Output, result)
}

func runAnalyticsTopPages(ctx context.Context, args []string) error {
	return runAnalyticsTop(ctx, args, "page")
}

func runAnalyticsTopQueries(ctx context.Context, args []string) error {
	return runAnalyticsTop(ctx, args, "query")
}

func runAnalyticsTop(ctx context.Context, args []string, dimension string) error {
	fs := flag.NewFlagSet("analytics top", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	var (
		startDate  = fs.String("start", "", "start date in YYYY-MM-DD; defaults to a recent 28-day window ending yesterday")
		endDate    = fs.String("end", "", "end date in YYYY-MM-DD; defaults to a recent 28-day window ending yesterday")
		days       = fs.Int("days", 28, "shortcut window in days: 7, 28, or 90")
		searchType = fs.String("search-type", "web", "search type: web,image,video,news,discover,google_news")
		dataState  = fs.String("data-state", "final", "data state: final,all,hourly_all")
		limit      = fs.Int64("limit", 10, "maximum rows to return (1-25000)")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}

	dateRange, err := cfg.ResolveDateRangeWithDays(*startDate, *endDate, *days)
	if err != nil {
		return err
	}
	req, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  dateRange.Start,
		EndDate:    dateRange.End,
		Dimensions: dimension,
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      *limit,
		StartRow:   0,
	})
	if err != nil {
		return err
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}

	result, err := client.QueryAnalytics(ctx, base.Site, req)
	if err != nil {
		return err
	}
	return output.RenderAnalytics(os.Stdout, base.Output, result)
}

func runReport(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "overview" {
		return errors.New("usage: gsc-cli report overview [flags]")
	}

	fs := flag.NewFlagSet("report overview", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	var (
		startDate  = fs.String("start", "", "start date in YYYY-MM-DD")
		endDate    = fs.String("end", "", "end date in YYYY-MM-DD")
		days       = fs.Int("days", 28, "shortcut window in days: 7, 28, or 90")
		limit      = fs.Int64("limit", 10, "maximum rows for top pages and top queries (1-25000)")
		searchType = fs.String("search-type", "web", "search type: web,image,video,news,discover,google_news")
		dataState  = fs.String("data-state", "final", "data state: final,all,hourly_all")
		inspectURL = fs.String("inspect-url", "", "URL to inspect; inferred from site when omitted")
	)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}

	dateRange, err := cfg.ResolveDateRangeWithDays(*startDate, *endDate, *days)
	if err != nil {
		return err
	}
	targetURL := strings.TrimSpace(*inspectURL)
	if targetURL == "" {
		targetURL, err = cfg.InferInspectionURL(base.Site)
		if err != nil {
			return err
		}
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}

	topPagesReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  dateRange.Start,
		EndDate:    dateRange.End,
		Dimensions: "page",
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      *limit,
	})
	if err != nil {
		return err
	}
	topQueriesReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  dateRange.Start,
		EndDate:    dateRange.End,
		Dimensions: "query",
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      *limit,
	})
	if err != nil {
		return err
	}
	inspectionReq, err := cfg.BuildInspectRequest(cfg.InspectInput{
		Site:         base.Site,
		URL:          targetURL,
		LanguageCode: "en-US",
	})
	if err != nil {
		return err
	}

	topPages, err := client.QueryAnalytics(ctx, base.Site, topPagesReq)
	if err != nil {
		return err
	}
	topQueries, err := client.QueryAnalytics(ctx, base.Site, topQueriesReq)
	if err != nil {
		return err
	}
	sitemaps, err := client.ListSitemaps(ctx, base.Site)
	if err != nil {
		return err
	}
	inspection, err := client.InspectURL(ctx, inspectionReq)
	if err != nil {
		return err
	}

	report := &gsc.OverviewReport{
		Site:       base.Site,
		StartDate:  dateRange.Start,
		EndDate:    dateRange.End,
		InspectURL: targetURL,
		TopPages:   topPages,
		TopQueries: topQueries,
		Sitemaps:   sitemaps,
		Inspection: inspection,
	}
	return output.RenderOverview(os.Stdout, base.Output, report)
}

func runSitemaps(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "list" {
		return errors.New("usage: gsc-cli sitemaps list [flags]")
	}

	fs := flag.NewFlagSet("sitemaps list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}

	result, err := client.ListSitemaps(ctx, base.Site)
	if err != nil {
		return err
	}
	return output.RenderSitemaps(os.Stdout, base.Output, result)
}

func runInspect(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "url" {
		return errors.New("usage: gsc-cli inspect url [flags]")
	}

	fs := flag.NewFlagSet("inspect url", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	var (
		targetURL = fs.String("url", "", "URL to inspect")
		language  = fs.String("language", "en-US", "response language, for example en-US")
	)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}

	req, err := cfg.BuildInspectRequest(cfg.InspectInput{
		Site:         base.Site,
		URL:          *targetURL,
		LanguageCode: *language,
	})
	if err != nil {
		return err
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}

	result, err := client.InspectURL(ctx, req)
	if err != nil {
		return err
	}
	return output.RenderInspection(os.Stdout, base.Output, result)
}

func bindCommonFlags(fs *flag.FlagSet) *cfg.Common {
	base := &cfg.Common{}
	fs.StringVar(&base.CredentialsFile, "credentials-file", "", "path to a Google service account JSON file")
	fs.StringVar(&base.Site, "site", "", "Search Console property, for example sc-domain:futurelab.studio")
	fs.StringVar(&base.Output, "output", cfg.OutputJSON, "output mode: json or table")
	return base
}

func printUsage() {
	fmt.Fprint(os.Stderr, strings.TrimSpace(fmt.Sprintf(`
gsc-cli - read-only Google Search Console CLI

Usage:
  gsc-cli sites list [flags]
  gsc-cli analytics query [flags]
  gsc-cli analytics top-pages [flags]
  gsc-cli analytics top-queries [flags]
  gsc-cli report overview [flags]
  gsc-cli sitemaps list [flags]
  gsc-cli inspect url [flags]

Examples:
  gsc-cli sites list --credentials-file service-account.json
  gsc-cli analytics query --credentials-file service-account.json --site sc-domain:futurelab.studio --start %s --end %s
  gsc-cli analytics top-pages --credentials-file service-account.json --site sc-domain:futurelab.studio --output table
  gsc-cli analytics top-queries --credentials-file service-account.json --site sc-domain:futurelab.studio --output table
  gsc-cli report overview --credentials-file service-account.json --site sc-domain:futurelab.studio --days 28 --output table
  gsc-cli sitemaps list --credentials-file service-account.json --site sc-domain:futurelab.studio
  gsc-cli inspect url --credentials-file service-account.json --site sc-domain:futurelab.studio --url https://futurelab.studio/
`, time.Now().AddDate(0, 0, -7).Format("2006-01-02"), time.Now().Format("2006-01-02"))))
	fmt.Fprintln(os.Stderr)
}
