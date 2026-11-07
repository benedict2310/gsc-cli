package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/benedict2310/gsc-cli/internal/commandspec"
	cfg "github.com/benedict2310/gsc-cli/internal/config"
	"github.com/benedict2310/gsc-cli/internal/output"
	gsc "github.com/benedict2310/gsc-cli/internal/searchconsole"
)

var (
	buildVersion = "dev"
	buildCommit  = ""
	buildDate    = ""
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
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
	case "doctor":
		return runDoctor(ctx, args[1:])
	case "report":
		return runReport(ctx, args[1:])
	case "cache":
		return runCache(ctx, args[1:])
	case "sync":
		return runSync(ctx, args[1:])
	case "diagnose":
		return runDiagnose(ctx, args[1:])
	case "monitor":
		return runMonitor(ctx, args[1:])
	case "audit":
		return runAudit(ctx, args[1:])
	case "opportunities":
		return runOpportunities(ctx, args[1:])
	case "sitemaps":
		return runSitemaps(ctx, args[1:])
	case "inspect":
		return runInspect(ctx, args[1:])
	case "version":
		return runVersion(args[1:])
	case "commands":
		return runCommands(args[1:])
	case "command":
		return runCommand(args[1:])
	case "help":
		return runHelp(args[1:])
	case "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runCommands(args []string) error {
	fs := flag.NewFlagSet("commands", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	outputMode := fs.String("output", cfg.OutputJSON, "output mode: json or table")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mode, err := normalizeOutputMode(*outputMode)
	if err != nil {
		return err
	}
	return renderCommandManifest(mode, commandspec.PublicManifest())
}

func runCommand(args []string) error {
	pathParts, outputMode, err := parseCommandSpecArgs(args)
	if err != nil {
		return err
	}
	mode, err := normalizeOutputMode(outputMode)
	if err != nil {
		return err
	}
	path := strings.Join(pathParts, " ")
	command, ok := commandspec.Find(path)
	if !ok {
		return fmt.Errorf("unknown command spec %q", path)
	}
	return renderCommandSpec(mode, command)
}

func parseCommandSpecArgs(args []string) ([]string, string, error) {
	outputMode := cfg.OutputJSON
	var pathParts []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--output" || arg == "-output":
			if i+1 >= len(args) {
				return nil, "", errors.New("output requires a value")
			}
			outputMode = args[i+1]
			i++
		case strings.HasPrefix(arg, "--output="):
			outputMode = strings.TrimPrefix(arg, "--output=")
		case arg == "--help" || arg == "-h":
			printCommandSpecUsage()
			return nil, "", flag.ErrHelp
		case strings.HasPrefix(arg, "-"):
			return nil, "", fmt.Errorf("unsupported command flag %q", arg)
		default:
			pathParts = append(pathParts, arg)
		}
	}
	if len(pathParts) == 0 {
		return nil, "", errors.New("usage: gsc-cli command <command-path> [--output json|table]")
	}
	return pathParts, outputMode, nil
}

func printCommandSpecUsage() {
	fmt.Fprintln(os.Stderr, "Usage: gsc-cli command <command-path> [--output json|table]")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  gsc-cli command diagnose drop --output json")
	fmt.Fprintln(os.Stderr, "  gsc-cli command audit sitemap --output table")
}

func normalizeOutputMode(mode string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != cfg.OutputJSON && mode != cfg.OutputTable {
		return "", fmt.Errorf("unsupported output mode %q; supported values are json and table", mode)
	}
	return mode, nil
}

func runHelp(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}
	switch args[0] {
	case "analytics":
		if len(args) == 1 {
			printAnalyticsUsage()
			return nil
		}
		return run(context.Background(), []string{"analytics", args[1], "--help"})
	case "report":
		if len(args) == 1 {
			printReportUsage()
			return nil
		}
		return run(context.Background(), []string{"report", args[1], "--help"})
	case "cache", "sync", "diagnose", "monitor", "audit", "sitemaps", "inspect":
		if len(args) > 1 {
			return run(context.Background(), []string{args[0], args[1], "--help"})
		}
		return run(context.Background(), []string{args[0], "--help"})
	case "doctor", "version", "sites", "opportunities", "commands", "command":
		return run(context.Background(), append([]string{args[0]}, "--help"))
	default:
		return fmt.Errorf("unknown help topic %q", args[0])
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
		printAnalyticsUsage()
		return errors.New("missing analytics subcommand")
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printAnalyticsUsage()
		return nil
	}
	switch args[0] {
	case "query":
		return runAnalyticsQuery(ctx, args[1:])
	case "top-pages":
		return runAnalyticsTopPages(ctx, args[1:])
	case "top-queries":
		return runAnalyticsTopQueries(ctx, args[1:])
	case "help":
		printAnalyticsUsage()
		return nil
	default:
		printAnalyticsUsage()
		return fmt.Errorf("unknown analytics subcommand %q", args[0])
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
	if err := fs.Parse(args); err != nil {
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
		startDate  = fs.String("start", "", "start date in YYYY-MM-DD; defaults to a 28-day window ending on the latest day with data")
		endDate    = fs.String("end", "", "end date in YYYY-MM-DD; defaults to a 28-day window ending on the latest day with data")
		days       = fs.String("days", "", "shortcut window in days: 7, 28, or 90; defaults to 28 when omitted")
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

	dayWindow, err := cfg.ParseDayWindow(*days)
	if err != nil {
		return err
	}
	dateRange, err := cfg.ResolveDateRangeWithDays(*startDate, *endDate, dayWindow)
	if err != nil {
		return err
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	if *startDate == "" && *endDate == "" {
		dateRange, err = alignToLatestData(ctx, client, base.Site, dateRange, *searchType, *dataState)
		if err != nil {
			return err
		}
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

	result, err := client.QueryAnalytics(ctx, base.Site, req)
	if err != nil {
		return err
	}
	return output.RenderAnalytics(os.Stdout, base.Output, result)
}

func runReport(ctx context.Context, args []string) error {
	if len(args) == 0 {
		printReportUsage()
		return errors.New("missing report subcommand")
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printReportUsage()
		return nil
	}
	if args[0] == "compare" {
		return runReportCompare(ctx, args[1:])
	}
	if args[0] != "overview" {
		printReportUsage()
		return fmt.Errorf("unknown report subcommand %q", args[0])
	}

	fs := flag.NewFlagSet("report overview", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	var (
		startDate  = fs.String("start", "", "start date in YYYY-MM-DD")
		endDate    = fs.String("end", "", "end date in YYYY-MM-DD")
		days       = fs.String("days", "", "shortcut window in days: 7, 28, or 90; defaults to 28 when omitted")
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

	dayWindow, err := cfg.ParseDayWindow(*days)
	if err != nil {
		return err
	}
	dateRange, err := cfg.ResolveDateRangeWithDays(*startDate, *endDate, dayWindow)
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
	if *startDate == "" && *endDate == "" {
		dateRange, err = alignToLatestData(ctx, client, base.Site, dateRange, *searchType, *dataState)
		if err != nil {
			return err
		}
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

func runReportCompare(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("report compare", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	var (
		startDate  = fs.String("start", "", "current period start date in YYYY-MM-DD")
		endDate    = fs.String("end", "", "current period end date in YYYY-MM-DD")
		days       = fs.String("days", "", "shortcut window in days: 7, 28, or 90; defaults to 28 when omitted")
		limit      = fs.Int64("limit", 10, "maximum rows for each period")
		searchType = fs.String("search-type", "web", "search type: web,image,video,news,discover,google_news")
		dataState  = fs.String("data-state", "final", "data state: final,all,hourly_all")
		dimension  = fs.String("dimension", "page", "comparison dimension: page or query")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(true); err != nil {
		return err
	}

	dim := strings.ToLower(strings.TrimSpace(*dimension))
	if dim != "page" && dim != "query" {
		return errors.New("dimension must be page or query")
	}
	dayWindow, err := cfg.ParseDayWindow(*days)
	if err != nil {
		return err
	}
	currentRange, err := cfg.ResolveDateRangeWithDays(*startDate, *endDate, dayWindow)
	if err != nil {
		return err
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		return err
	}
	if *startDate == "" && *endDate == "" {
		currentRange, err = alignToLatestData(ctx, client, base.Site, currentRange, *searchType, *dataState)
		if err != nil {
			return err
		}
	}
	previousRange, err := cfg.PreviousDateRange(currentRange)
	if err != nil {
		return err
	}
	currentReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  currentRange.Start,
		EndDate:    currentRange.End,
		Dimensions: dim,
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      *limit,
	})
	if err != nil {
		return err
	}
	previousReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  previousRange.Start,
		EndDate:    previousRange.End,
		Dimensions: dim,
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      *limit,
	})
	if err != nil {
		return err
	}

	currentResult, err := client.QueryAnalytics(ctx, base.Site, currentReq)
	if err != nil {
		return err
	}
	previousResult, err := client.QueryAnalytics(ctx, base.Site, previousReq)
	if err != nil {
		return err
	}

	currentAggregateReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  currentRange.Start,
		EndDate:    currentRange.End,
		Dimensions: "date",
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      25000,
	})
	if err != nil {
		return err
	}
	previousAggregateReq, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  previousRange.Start,
		EndDate:    previousRange.End,
		Dimensions: "date",
		SearchType: *searchType,
		DataState:  *dataState,
		Limit:      25000,
	})
	if err != nil {
		return err
	}
	currentAggregate, err := client.QueryAnalytics(ctx, base.Site, currentAggregateReq)
	if err != nil {
		return err
	}
	previousAggregate, err := client.QueryAnalytics(ctx, base.Site, previousAggregateReq)
	if err != nil {
		return err
	}

	report := buildCompareReport(base.Site, dim, currentRange, previousRange, currentResult, previousResult, currentAggregate, previousAggregate)
	return renderCompare(base.Output, report)
}

func buildCompareReport(site, dimension string, currentRange, previousRange cfg.DateRange, currentDetail, previousDetail, currentAggregate, previousAggregate *gsc.AnalyticsResult) gsc.CompareReport {
	currentImpressions, currentClicks := sumAnalyticsRows(currentAggregate.Rows)
	previousImpressions, previousClicks := sumAnalyticsRows(previousAggregate.Rows)

	return gsc.CompareReport{
		Site:      site,
		Dimension: strings.ToUpper(dimension),
		Summary: gsc.CompareSummary{
			CurrentStart:        currentRange.Start,
			CurrentEnd:          currentRange.End,
			PreviousStart:       previousRange.Start,
			PreviousEnd:         previousRange.End,
			CurrentRows:         len(currentDetail.Rows),
			PreviousRows:        len(previousDetail.Rows),
			CurrentImpressions:  currentImpressions,
			PreviousImpressions: previousImpressions,
			CurrentClicks:       currentClicks,
			PreviousClicks:      previousClicks,
			ImpressionDelta:     currentImpressions - previousImpressions,
			ClickDelta:          currentClicks - previousClicks,
		},
		Current:  currentDetail,
		Previous: previousDetail,
	}
}

func runDoctor(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := bindCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := base.Resolve(false); err != nil {
		return err
	}

	report := doctorReport{
		Version: currentVersionInfo().Version,
		Site:    base.Site,
		Checks: []doctorCheck{
			{Name: "credentials_file", Status: "PASS", Detail: credentialDetail(base.CredentialsFile)},
		},
	}

	client, err := gsc.New(ctx, gsc.Options{CredentialsFile: base.CredentialsFile})
	if err != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name:   "client_init",
			Status: "FAIL",
			Detail: err.Error(),
		})
		return renderDoctor(base.Output, report)
	}
	report.Checks = append(report.Checks, doctorCheck{
		Name:   "client_init",
		Status: "PASS",
		Detail: "Search Console client initialized",
	})

	sites, err := client.ListSites(ctx)
	if err != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name:   "site_list",
			Status: "FAIL",
			Detail: err.Error(),
		})
		return renderDoctor(base.Output, report)
	}
	report.Checks = append(report.Checks, doctorCheck{
		Name:   "site_list",
		Status: "PASS",
		Detail: fmt.Sprintf("%d visible properties", len(sites)),
	})

	if strings.TrimSpace(base.Site) != "" {
		found := hasSite(sites, base.Site)
		report.Checks = append(report.Checks, doctorCheck{
			Name:   "configured_site_access",
			Status: statusForMatch(found),
			Detail: siteAccessDetail(base.Site, client.ServiceAccountEmail(), found),
		})
		if inferredURL, inferErr := cfg.InferInspectionURL(base.Site); inferErr == nil {
			report.Checks = append(report.Checks, doctorCheck{
				Name:   "inspection_url_inference",
				Status: "PASS",
				Detail: inferredURL,
			})
		} else {
			report.Checks = append(report.Checks, doctorCheck{
				Name:   "inspection_url_inference",
				Status: "FAIL",
				Detail: inferErr.Error(),
			})
		}
	} else {
		report.Checks = append(report.Checks, doctorCheck{
			Name:   "configured_site_access",
			Status: "SKIP",
			Detail: "no site configured; pass --site or set GSC_CLI_SITE_URL",
		})
	}

	return renderDoctor(base.Output, report)
}

func runVersion(args []string) error {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	outputMode := fs.String("output", cfg.OutputTable, "output mode: json or table")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mode := strings.ToLower(strings.TrimSpace(*outputMode))
	if mode != cfg.OutputJSON && mode != cfg.OutputTable {
		return fmt.Errorf("unsupported output mode %q; supported values are json and table", mode)
	}
	return renderVersion(mode, currentVersionInfo())
}

func currentVersionInfo() versionInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		info = nil
	}
	return resolveVersionInfo(buildVersion, buildCommit, buildDate, info)
}

// resolveVersionInfo prefers ldflags values and falls back to what the Go toolchain stamps into every build.
func resolveVersionInfo(version, commit, date string, info *debug.BuildInfo) versionInfo {
	result := versionInfo{Version: version, Commit: commit, Date: date}
	if info == nil {
		return result
	}
	if result.Version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		result.Version = info.Main.Version
	}
	if result.Commit != "" && result.Date != "" {
		return result
	}
	var revision, vcsTime string
	modified := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.time":
			vcsTime = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if result.Commit == "" && revision != "" {
		result.Commit = revision
		if modified {
			result.Commit += "-dirty"
		}
	}
	if result.Date == "" {
		result.Date = vcsTime
	}
	return result
}

func credentialDetail(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return "<provided>"
}

func bindCommonFlags(fs *flag.FlagSet) *cfg.Common {
	base := &cfg.Common{}
	fs.StringVar(&base.CredentialsFile, "credentials-file", "", "path to a Google service account JSON file")
	fs.StringVar(&base.Site, "site", "", "Search Console property, for example sc-domain:example.com")
	fs.StringVar(&base.Output, "output", cfg.OutputJSON, "output mode: json or table")
	return base
}

func printUsage() {
	fmt.Fprint(os.Stderr, strings.TrimSpace(fmt.Sprintf(`
gsc-cli - Google Search Console diagnostics and SEO intelligence CLI

Start here:
  gsc-cli report overview --site sc-domain:example.com --credentials-file ./service-account.json --output table

Usage:
  gsc-cli sites list [flags]
  gsc-cli analytics query [flags]
  gsc-cli analytics top-pages [flags]
  gsc-cli analytics top-queries [flags]
  gsc-cli doctor [flags]
  gsc-cli help <topic>
  gsc-cli help <topic> <subcommand>
  gsc-cli report overview [flags]
  gsc-cli report compare [flags]
  gsc-cli cache status [flags]
  gsc-cli sync analytics [flags]
  gsc-cli sync hourly [flags]
  gsc-cli diagnose drop [flags]
  gsc-cli monitor hourly [flags]
  gsc-cli audit indexability [flags]
  gsc-cli audit sitemap [flags]
  gsc-cli opportunities [flags]
  gsc-cli sitemaps list|get|submit|delete [flags]
  gsc-cli inspect url|batch [flags]
  gsc-cli version [flags]
  gsc-cli commands [flags]
  gsc-cli command <command-path> [flags]

Commands:
  sites list             List visible Search Console properties
  analytics query        Run a custom Search Analytics query with explicit dimensions
  analytics top-pages    Fast top-pages report
  analytics top-queries  Fast top-queries report
  doctor                 Check credentials, property visibility, and site access
  report overview        One-shot operator snapshot: pages, queries, sitemap, inspection
  report compare         Compare current period vs previous equal period
  cache status           Show local cache/store status
  sync analytics         Pull Search Analytics rows into the local cache
  sync hourly            Pull provisional hourly Search Analytics rows into the cache
  diagnose drop          Explain likely causes behind traffic drops
  monitor hourly         Compare today's hourly performance with previous week
  audit indexability     Inspect a batch of URLs and classify issues
  audit sitemap          Compare a sitemap against Search Analytics pages
  opportunities          Rank near-win and CTR-gap opportunities
  sitemaps list|get|submit|delete  Manage sitemaps for a property
  inspect url|batch      Inspect one or many URLs for indexing state
  version                Print build and version information
  commands               List machine-readable command metadata
  command                Show metadata for one command path

Examples:
  gsc-cli sites list --credentials-file service-account.json
  gsc-cli analytics query --credentials-file service-account.json --site sc-domain:example.com --start %s --end %s
  gsc-cli analytics top-pages --credentials-file service-account.json --site sc-domain:example.com --output table
  gsc-cli analytics top-queries --credentials-file service-account.json --site sc-domain:example.com --output table
  gsc-cli doctor --credentials-file service-account.json --site sc-domain:example.com --output table
  gsc-cli help analytics query
  gsc-cli report overview --credentials-file service-account.json --site sc-domain:example.com --days 28 --output table
  gsc-cli report compare --credentials-file service-account.json --site sc-domain:example.com --days 28 --dimension page --output table
  gsc-cli cache status --cache-file .gsc-cli-cache.json
  gsc-cli sync analytics --credentials-file service-account.json --site sc-domain:example.com --period last-28d --dimensions page,query
  gsc-cli diagnose drop --credentials-file service-account.json --site sc-domain:example.com --current 2026-03-01:2026-03-14 --baseline previous-period
  gsc-cli monitor hourly --credentials-file service-account.json --site sc-domain:example.com --output table
  gsc-cli audit sitemap --credentials-file service-account.json --site sc-domain:example.com --sitemap-url https://example.com/sitemap.xml --period last-90d
  gsc-cli opportunities --credentials-file service-account.json --site sc-domain:example.com --period last-90d --types near-win,ctr-gap
  gsc-cli sitemaps list --credentials-file service-account.json --site sc-domain:example.com
  gsc-cli inspect url --credentials-file service-account.json --site sc-domain:example.com --url https://example.com/
  gsc-cli version --output json
  gsc-cli commands --output json
  gsc-cli command diagnose drop --output json
`, time.Now().AddDate(0, 0, -7).Format("2006-01-02"), time.Now().Format("2006-01-02"))))
	fmt.Fprintln(os.Stderr)
}

func printAnalyticsUsage() {
	fmt.Fprint(os.Stderr, strings.TrimSpace(`
gsc-cli analytics

Usage:
  gsc-cli analytics query [flags]
  gsc-cli analytics top-pages [flags]
  gsc-cli analytics top-queries [flags]
  gsc-cli help analytics <subcommand>

Recommended:
  Start with top-pages or top-queries.
  Use query only when you need custom dimensions or row windows.
`))
	fmt.Fprintln(os.Stderr)
}

func printReportUsage() {
	fmt.Fprint(os.Stderr, strings.TrimSpace(`
gsc-cli report

Usage:
  gsc-cli report overview [flags]
  gsc-cli report compare [flags]
  gsc-cli help report <subcommand>

Recommended:
  Start here for a cold read on a property.
`))
	fmt.Fprintln(os.Stderr)
}

// alignToLatestData ends a defaulted window at the last date Search Console has data for,
// because final data lags a few days and empty trailing days read as a traffic drop.
func alignToLatestData(ctx context.Context, client *gsc.Client, site string, window cfg.DateRange, searchType, dataState string) (cfg.DateRange, error) {
	if strings.EqualFold(strings.TrimSpace(dataState), "hourly_all") {
		return window, nil
	}
	end, err := time.Parse("2006-01-02", window.End)
	if err != nil {
		return cfg.DateRange{}, fmt.Errorf("parse end date: %w", err)
	}
	req, err := cfg.BuildAnalyticsRequest(cfg.AnalyticsInput{
		StartDate:  end.AddDate(0, 0, -13).Format("2006-01-02"),
		EndDate:    window.End,
		Dimensions: "date",
		SearchType: searchType,
		DataState:  dataState,
		Limit:      25000,
	})
	if err != nil {
		return cfg.DateRange{}, err
	}
	probe, err := client.QueryAnalytics(ctx, site, req)
	if err != nil {
		return cfg.DateRange{}, err
	}
	latest := ""
	for _, row := range probe.Rows {
		if len(row.Keys) > 0 && row.Keys[0] > latest {
			latest = row.Keys[0]
		}
	}
	return cfg.EndRangeAt(window, latest)
}
