package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	cfg "github.com/benedict2310/gsc-cli/internal/config"
	"github.com/benedict2310/gsc-cli/internal/insights"
	"github.com/benedict2310/gsc-cli/internal/output"
	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
	"github.com/benedict2310/gsc-cli/internal/sitemap"
	"github.com/benedict2310/gsc-cli/internal/store"
)

const envCacheFile = "GSC_CLI_CACHE_FILE"

type syncReport struct {
	Path        string               `json:"path"`
	Hash        string               `json:"hash"`
	Rows        int                  `json:"rows"`
	Meta        gsc.QueryAllMetadata `json:"meta"`
	Stored      store.Status         `json:"stored"`
	Provisional bool                 `json:"provisional"`
}

type dropReports struct {
	Current  cfg.DateRange         `json:"current"`
	Baseline cfg.DateRange         `json:"baseline"`
	Reports  []insights.DropReport `json:"reports"`
}

type indexabilityReport struct {
	Results []indexabilityRow `json:"results"`
}

type indexabilityRow struct {
	URL     string   `json:"url"`
	Verdict string   `json:"verdict,omitempty"`
	Issues  []string `json:"issues,omitempty"`
}

func runCache(ctx context.Context, args []string) error {
	_ = ctx
	if len(args) == 0 || isHelpArg(args[0]) {
		printGroupUsage("cache", "gsc-cli cache status [flags]")
		return nil
	}
	if args[0] != "status" {
		return errors.New("usage: gsc-cli cache status [--cache-file path] [--output json|table]")
	}
	fs := flag.NewFlagSet("cache status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	outputMode := fs.String("output", cfg.OutputJSON, "output mode: json or table")
	cacheFile := fs.String("cache-file", defaultCacheFile(), "cache file path")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	status, err := store.New(*cacheFile).Status()
	if err != nil {
		return err
	}
	return renderCacheStatus(strings.ToLower(strings.TrimSpace(*outputMode)), status)
}

func runSync(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printGroupUsage("sync", "gsc-cli sync analytics [flags]\n  gsc-cli sync hourly [flags]")
		return nil
	}
	switch args[0] {
	case "analytics":
		return runSyncAnalytics(ctx, args[1:])
	case "hourly":
		return runSyncHourly(ctx, args[1:])
	default:
		return fmt.Errorf("unknown sync subcommand %q", args[0])
	}
}

func runSyncAnalytics(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("sync analytics", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	period := fs.String("period", "last-28d", "period like last-28d or last-16m")
	dimensions := fs.String("dimensions", "page,query", "comma-separated dimensions")
	searchType := fs.String("search-type", "web", "search type")
	dataState := fs.String("data-state", "final", "data state")
	cacheFile := fs.String("cache-file", defaultCacheFile(), "cache file path")
	maxRows := fs.Int("max-rows", 0, "optional maximum rows to store")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	rangeValue, err := cfg.ResolveNamedPeriod(*period, time.Now().UTC())
	if err != nil {
		return err
	}
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	rangeValue, err = alignToLatestData(ctx, client, base.Site, rangeValue, *searchType, *dataState)
	if err != nil {
		return err
	}
	req, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: rangeValue.Start, EndDate: rangeValue.End, Dimensions: *dimensions, SearchType: *searchType, DataState: *dataState, Limit: 25000})
	if err != nil {
		return err
	}
	result, meta, err := client.QueryAnalyticsAll(ctx, base.Site, req, gsc.QueryAllOptions{MaxRows: *maxRows})
	if err != nil {
		return err
	}
	hash, err := gsc.AnalyticsRequestHash(base.Site, req)
	if err != nil {
		return err
	}
	cache := store.New(*cacheFile)
	result.Site = base.Site
	if err := cache.SaveAnalytics(hash, req, result, meta, req.DataState == "HOURLY_ALL"); err != nil {
		return err
	}
	status, err := cache.Status()
	if err != nil {
		return err
	}
	return renderSyncReport(base.Output, syncReport{Path: *cacheFile, Hash: hash, Rows: len(result.Rows), Meta: meta, Stored: status, Provisional: req.DataState == "HOURLY_ALL"})
}

func runSyncHourly(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("sync hourly", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	last := fs.Int("last", 10, "number of trailing days to sync")
	cacheFile := fs.String("cache-file", defaultCacheFile(), "cache file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	period := fmt.Sprintf("last-%dd", *last)
	rangeValue, err := cfg.ResolveNamedPeriod(period, time.Now().UTC())
	if err != nil {
		return err
	}
	req, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: rangeValue.Start, EndDate: rangeValue.End, Dimensions: "hour", SearchType: "web", DataState: "hourly_all", Limit: 25000})
	if err != nil {
		return err
	}
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	result, meta, err := client.QueryAnalyticsAll(ctx, base.Site, req, gsc.QueryAllOptions{})
	if err != nil {
		return err
	}
	hash, err := gsc.AnalyticsRequestHash(base.Site, req)
	if err != nil {
		return err
	}
	cache := store.New(*cacheFile)
	result.Site = base.Site
	if err := cache.SaveAnalytics(hash, req, result, meta, true); err != nil {
		return err
	}
	status, err := cache.Status()
	if err != nil {
		return err
	}
	return renderSyncReport(base.Output, syncReport{Path: *cacheFile, Hash: hash, Rows: len(result.Rows), Meta: meta, Stored: status, Provisional: true})
}

func runDiagnose(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printGroupUsage("diagnose", "gsc-cli diagnose drop [flags]")
		return nil
	}
	if args[0] != "drop" {
		return errors.New("usage: gsc-cli diagnose drop [flags]")
	}
	fs := flag.NewFlagSet("diagnose drop", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	current := fs.String("current", "", "current period YYYY-MM-DD:YYYY-MM-DD")
	baseline := fs.String("baseline", "previous-period", "baseline period or previous-period")
	segments := fs.String("segments", "page", "comma-separated dimensions: page,query,country,device,searchAppearance,date")
	limit := fs.Int("limit", 5, "maximum contributors per segment")
	minLoss := fs.Float64("min-loss", 1, "minimum click loss to include")
	searchType := fs.String("search-type", "web", "search type")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	currentRange, err := cfg.ParsePeriodRange(*current)
	if err != nil {
		return err
	}
	baselineRange, err := cfg.ResolveBaseline(currentRange, *baseline)
	if err != nil {
		return err
	}
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	var reports []insights.DropReport
	for _, segment := range splitCSV(*segments) {
		dimension := normalizeSegmentDimension(segment)
		currentReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: currentRange.Start, EndDate: currentRange.End, Dimensions: dimension, SearchType: *searchType, DataState: "final", Limit: 25000})
		if err != nil {
			return err
		}
		baselineReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: baselineRange.Start, EndDate: baselineRange.End, Dimensions: dimension, SearchType: *searchType, DataState: "final", Limit: 25000})
		if err != nil {
			return err
		}
		currentRows, _, err := client.QueryAnalyticsAll(ctx, base.Site, currentReq, gsc.QueryAllOptions{})
		if err != nil {
			return err
		}
		baselineRows, _, err := client.QueryAnalyticsAll(ctx, base.Site, baselineReq, gsc.QueryAllOptions{})
		if err != nil {
			return err
		}
		reports = append(reports, insights.DiagnoseDrop(currentRows, baselineRows, insights.DiagnoseOptions{Segment: segment, MinLoss: *minLoss, Limit: *limit}))
	}
	return renderDropReports(base.Output, dropReports{Current: currentRange, Baseline: baselineRange, Reports: reports})
}

func runMonitor(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printGroupUsage("monitor", "gsc-cli monitor hourly [flags]")
		return nil
	}
	if args[0] != "hourly" {
		return errors.New("usage: gsc-cli monitor hourly [flags]")
	}
	fs := flag.NewFlagSet("monitor hourly", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	dateValue := fs.String("date", time.Now().UTC().Format("2006-01-02"), "current date in YYYY-MM-DD")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	currentDate := *dateValue
	parsed, err := time.Parse("2006-01-02", currentDate)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}
	baselineDate := parsed.AddDate(0, 0, -7).Format("2006-01-02")
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	currentReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: currentDate, EndDate: currentDate, Dimensions: "hour", SearchType: "web", DataState: "hourly_all", Limit: 25000})
	if err != nil {
		return err
	}
	baselineReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: baselineDate, EndDate: baselineDate, Dimensions: "hour", SearchType: "web", DataState: "hourly_all", Limit: 25000})
	if err != nil {
		return err
	}
	currentRows, _, err := client.QueryAnalyticsAll(ctx, base.Site, currentReq, gsc.QueryAllOptions{})
	if err != nil {
		return err
	}
	baselineRows, _, err := client.QueryAnalyticsAll(ctx, base.Site, baselineReq, gsc.QueryAllOptions{})
	if err != nil {
		return err
	}
	return renderHourlyComparison(base.Output, insights.BuildHourlyComparison(currentRows, baselineRows))
}

func runAudit(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printGroupUsage("audit", "gsc-cli audit indexability [flags]\n  gsc-cli audit sitemap [flags]")
		return nil
	}
	switch args[0] {
	case "indexability":
		return runAuditIndexability(ctx, args[1:])
	case "sitemap":
		return runAuditSitemap(ctx, args[1:])
	default:
		return fmt.Errorf("unknown audit subcommand %q", args[0])
	}
}

func runAuditIndexability(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("audit indexability", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	urlsFile := fs.String("urls-file", "", "file with one URL per line")
	maxURLs := fs.Int("max-urls", 100, "maximum URLs to inspect")
	concurrency := fs.Int("concurrency", 1, "inspection concurrency")
	cacheFile := fs.String("cache-file", defaultCacheFile(), "cache file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	return inspectBatch(ctx, base, *urlsFile, *maxURLs, *concurrency, *cacheFile)
}

func runAuditSitemap(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("audit sitemap", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	sitemapURL := fs.String("sitemap-url", "", "sitemap XML URL")
	period := fs.String("period", "last-90d", "period like last-90d")
	limit := fs.Int("limit", 25000, "maximum analytics rows")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	if strings.TrimSpace(*sitemapURL) == "" {
		return errors.New("sitemap-url is required")
	}
	urls, err := sitemap.Fetch(httpClientOrDefault(nil), *sitemapURL)
	if err != nil {
		return err
	}
	rangeValue, err := cfg.ResolveNamedPeriod(*period, time.Now().UTC())
	if err != nil {
		return err
	}
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	rangeValue, err = alignToLatestData(ctx, client, base.Site, rangeValue, "web", "final")
	if err != nil {
		return err
	}
	req, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: rangeValue.Start, EndDate: rangeValue.End, Dimensions: "page", SearchType: "web", DataState: "final", Limit: int64(*limit)})
	if err != nil {
		return err
	}
	analytics, _, err := client.QueryAnalyticsAll(ctx, base.Site, req, gsc.QueryAllOptions{MaxRows: *limit})
	if err != nil {
		return err
	}
	return renderSitemapAudit(base.Output, insights.AuditSitemap(urls, analytics))
}

func runOpportunities(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("opportunities", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	period := fs.String("period", "last-90d", "period like last-90d")
	types := fs.String("types", "near-win,ctr-gap", "opportunity types")
	minImpressions := fs.Float64("min-impressions", 100, "minimum impressions")
	limit := fs.Int("limit", 20, "maximum items")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	rangeValue, err := cfg.ResolveNamedPeriod(*period, time.Now().UTC())
	if err != nil {
		return err
	}
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	rangeValue, err = alignToLatestData(ctx, client, base.Site, rangeValue, "web", "final")
	if err != nil {
		return err
	}
	currentReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: rangeValue.Start, EndDate: rangeValue.End, Dimensions: "query,page", SearchType: "web", DataState: "final", Limit: 25000})
	if err != nil {
		return err
	}
	currentRows, _, err := client.QueryAnalyticsAll(ctx, base.Site, currentReq, gsc.QueryAllOptions{})
	if err != nil {
		return err
	}
	var previousRows *gsc.AnalyticsResult
	if hasType(*types, "decay") {
		baselineRange, err := cfg.ResolveBaseline(rangeValue, "previous-period")
		if err != nil {
			return err
		}
		previousReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{StartDate: baselineRange.Start, EndDate: baselineRange.End, Dimensions: "query,page", SearchType: "web", DataState: "final", Limit: 25000})
		if err != nil {
			return err
		}
		previousRows, _, err = client.QueryAnalyticsAll(ctx, base.Site, previousReq, gsc.QueryAllOptions{})
		if err != nil {
			return err
		}
	}
	report := insights.FindOpportunities(currentRows, previousRows, insights.OpportunityOptions{Types: splitCSV(*types), MinImpressions: *minImpressions, Limit: *limit})
	return renderOpportunities(base.Output, report)
}

func runSitemaps(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printGroupUsage("sitemaps", "gsc-cli sitemaps list [flags]\n  gsc-cli sitemaps get --feedpath URL [flags]\n  gsc-cli sitemaps submit --feedpath URL [flags]\n  gsc-cli sitemaps delete --feedpath URL [flags]")
		return nil
	}
	fs := flag.NewFlagSet("sitemaps", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	feedpath := fs.String("feedpath", "", "sitemap URL")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}
	writeScope := args[0] == "submit" || args[0] == "delete"
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile, WriteScope: writeScope})
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		result, err := client.ListSitemaps(ctx, base.Site)
		if err != nil {
			return err
		}
		return output.RenderSitemaps(os.Stdout, base.Output, result)
	case "get":
		if strings.TrimSpace(*feedpath) == "" {
			return errors.New("feedpath is required")
		}
		result, err := client.GetSitemap(ctx, base.Site, *feedpath)
		if err != nil {
			return err
		}
		return output.RenderSitemaps(os.Stdout, base.Output, []gsc.Sitemap{*result})
	case "submit":
		if strings.TrimSpace(*feedpath) == "" {
			return errors.New("feedpath is required")
		}
		if err := client.SubmitSitemap(ctx, base.Site, *feedpath); err != nil {
			return err
		}
		return renderAction(base.Output, "submit_sitemap", *feedpath, "ok")
	case "delete":
		if strings.TrimSpace(*feedpath) == "" {
			return errors.New("feedpath is required")
		}
		if err := client.DeleteSitemap(ctx, base.Site, *feedpath); err != nil {
			return err
		}
		return renderAction(base.Output, "delete_sitemap", *feedpath, "ok")
	default:
		return fmt.Errorf("unknown sitemaps subcommand %q", args[0])
	}
}

func runInspect(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printGroupUsage("inspect", "gsc-cli inspect url [flags]\n  gsc-cli inspect batch [flags]")
		return nil
	}
	if args[0] == "batch" {
		fs := flag.NewFlagSet("inspect batch", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		base := bindCommonFlags(fs)
		urlsFile := fs.String("urls-file", "", "file with one URL per line")
		maxURLs := fs.Int("max-urls", 100, "maximum URLs to inspect")
		concurrency := fs.Int("concurrency", 1, "inspection concurrency")
		cacheFile := fs.String("cache-file", defaultCacheFile(), "cache file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := base.Resolve(true); err != nil {
			return err
		}
		return inspectBatch(ctx, base, *urlsFile, *maxURLs, *concurrency, *cacheFile)
	}
	if args[0] != "url" {
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

func inspectBatch(ctx context.Context, base *cfg.Common, urlsFile string, maxURLs, concurrency int, cacheFile string) error {
	if concurrency != 1 {
		return errors.New("only concurrency=1 is supported in this MVP")
	}
	urls, err := readLines(urlsFile, maxURLs)
	if err != nil {
		return err
	}
	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	cache := store.New(cacheFile)
	report := indexabilityReport{}
	for _, url := range urls {
		req, err := cfg.BuildInspectRequest(cfg.InspectInput{Site: base.Site, URL: url, LanguageCode: "en-US"})
		if err != nil {
			return err
		}
		result, err := client.InspectURL(ctx, req)
		if err != nil {
			return err
		}
		issues := insights.ClassifyInspectionIssues(*result)
		if err := cache.SaveInspection(*result, issues); err != nil {
			return err
		}
		report.Results = append(report.Results, indexabilityRow{URL: url, Verdict: result.Verdict, Issues: issues})
		time.Sleep(100 * time.Millisecond)
	}
	return renderIndexability(base.Output, report)
}

func isHelpArg(arg string) bool {
	return arg == "help" || arg == "-h" || arg == "--help"
}

func printGroupUsage(name, usage string) {
	fmt.Fprintf(os.Stderr, "gsc-cli %s\n\nUsage:\n  %s\n", name, usage)
}

func defaultCacheFile() string {
	if value := strings.TrimSpace(os.Getenv(envCacheFile)); value != "" {
		return value
	}
	return ".gsc-cli-cache.json"
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func normalizeSegmentDimension(segment string) string {
	switch strings.ToLower(strings.TrimSpace(segment)) {
	case "searchappearance", "search_appearance":
		return "search_appearance"
	default:
		return strings.ToLower(strings.TrimSpace(segment))
	}
}

func hasType(types, target string) bool {
	for _, item := range splitCSV(types) {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func readLines(path string, max int) ([]string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("urls-file is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if max > 0 && len(lines) >= max {
			break
		}
	}
	return lines, scanner.Err()
}

func httpClientOrDefault(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return http.DefaultClient
}

func renderCacheStatus(mode string, status store.Status) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, status)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	fmt.Fprintf(tw, "PATH\t%s\n", status.Path)
	fmt.Fprintf(tw, "ANALYTICS_ROWS\t%d\n", status.AnalyticsRows)
	fmt.Fprintf(tw, "PULL_LOGS\t%d\n", status.PullLogs)
	fmt.Fprintf(tw, "INSPECTIONS\t%d\n", status.Inspections)
	return tw.Flush()
}

func renderSyncReport(mode string, report syncReport) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	fmt.Fprintf(tw, "CACHE_FILE\t%s\n", report.Path)
	fmt.Fprintf(tw, "HASH\t%s\n", report.Hash)
	fmt.Fprintf(tw, "ROWS\t%d\n", report.Rows)
	fmt.Fprintf(tw, "PAGES_FETCHED\t%d\n", report.Meta.PagesFetched)
	fmt.Fprintf(tw, "PROVISIONAL\t%t\n", report.Provisional)
	return tw.Flush()
}

func renderDropReports(mode string, report dropReports) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SEGMENT\tLIKELY_CAUSE\tKEY\tCONTRIBUTION\tSHARE_OF_LOSS")
	for _, segment := range report.Reports {
		for _, item := range segment.Contributors {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%.0f\t%.2f\n", segment.Segment, segment.LikelyCause, item.Key, item.Contribution, item.ShareOfLoss)
		}
	}
	return tw.Flush()
}

func renderHourlyComparison(mode string, report insights.HourlyComparison) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "HOUR\tCURRENT_CLICKS\tBASELINE_CLICKS\tCLICK_DELTA\tPROVISIONAL")
	for _, row := range report.Rows {
		fmt.Fprintf(tw, "%s\t%.0f\t%.0f\t%.0f\t%t\n", row.Hour, row.CurrentClicks, row.BaselineClicks, row.ClickDelta, report.Provisional)
	}
	return tw.Flush()
}

func renderIndexability(mode string, report indexabilityReport) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "URL\tVERDICT\tISSUES")
	for _, row := range report.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", row.URL, row.Verdict, strings.Join(row.Issues, ","))
	}
	return tw.Flush()
}

func renderSitemapAudit(mode string, report insights.SitemapAuditReport) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FIELD\tVALUE")
	fmt.Fprintf(tw, "IN_SITEMAP_WITH_IMPRESSIONS\t%d\n", report.InSitemapWithImpressions)
	fmt.Fprintf(tw, "WITHOUT_IMPRESSIONS\t%s\n", strings.Join(report.WithoutImpressions, ", "))
	fmt.Fprintf(tw, "MISSING_FROM_SITEMAP\t%s\n", strings.Join(report.MissingFromSitemap, ", "))
	return tw.Flush()
}

func renderOpportunities(mode string, report insights.OpportunityReport) error {
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, report)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tQUERY\tPAGE\tIMPRESSIONS\tCTR\tPOSITION\tSCORE\tACTION")
	for _, item := range report.Items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%.0f\t%.4f\t%.1f\t%.0f\t%s\n", item.Type, item.Query, item.Page, item.Impressions, item.CTR, item.Position, item.Score, item.Action)
	}
	return tw.Flush()
}

func renderAction(mode, action, target, status string) error {
	payload := map[string]string{"action": action, "target": target, "status": status}
	if mode == cfg.OutputJSON {
		return writeJSON(os.Stdout, payload)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ACTION\tTARGET\tSTATUS")
	fmt.Fprintf(tw, "%s\t%s\t%s\n", action, target, status)
	return tw.Flush()
}
