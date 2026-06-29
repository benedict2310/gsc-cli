package commandspec

import "strings"

const (
	SafetyReadOnly     = "read-only"
	SafetyWriteCapable = "write-capable"
)

type Flag struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Default       string   `json:"default,omitempty"`
	Required      bool     `json:"required"`
	EnvFallback   string   `json:"envFallback,omitempty"`
	AllowedValues []string `json:"allowedValues,omitempty"`
}

type Command struct {
	Path                string   `json:"path"`
	Summary             string   `json:"summary"`
	Description         string   `json:"description"`
	Flags               []Flag   `json:"flags,omitempty"`
	Examples            []string `json:"examples,omitempty"`
	OutputModes         []string `json:"outputModes,omitempty"`
	RequiresCredentials bool     `json:"requiresCredentials"`
	RequiresSite        bool     `json:"requiresSite"`
	APIsUsed            []string `json:"apisUsed,omitempty"`
	Safety              string   `json:"safety"`
	WriteScopeRequired  bool     `json:"writeScopeRequired"`
	Caveats             []string `json:"caveats,omitempty"`
	RelatedCommands     []string `json:"relatedCommands,omitempty"`
}

type Manifest struct {
	Commands []Command `json:"commands"`
}

func PublicManifest() Manifest {
	return Manifest{Commands: publicCommands()}
}

func Find(path string) (Command, bool) {
	normalized := normalizePath(path)
	for _, command := range publicCommands() {
		if normalizePath(command.Path) == normalized {
			return command, true
		}
	}
	return Command{}, false
}

func normalizePath(path string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(path)), " ")
}

func publicCommands() []Command {
	outputModes := []string{"json", "table"}
	commonOutput := []Flag{outputFlag()}
	commonFlags := func(site bool) []Flag {
		flags := []Flag{credentialsFlag()}
		if site {
			flags = append(flags, siteFlag())
		}
		flags = append(flags, outputFlag())
		return flags
	}
	analyticsFlags := func(site bool) []Flag {
		flags := commonFlags(site)
		flags = append(flags,
			Flag{Name: "start", Type: "string"},
			Flag{Name: "end", Type: "string"},
			Flag{Name: "search-type", Type: "string", Default: "web", AllowedValues: []string{"web", "image", "video", "news", "discover", "google_news"}},
			Flag{Name: "data-state", Type: "string", Default: "final", AllowedValues: []string{"final", "all", "hourly_all"}},
		)
		return flags
	}
	topAnalyticsFlags := func() []Flag {
		return append(commonFlags(true),
			Flag{Name: "start", Type: "string"},
			Flag{Name: "end", Type: "string"},
			Flag{Name: "days", Type: "string", AllowedValues: []string{"7", "28", "90"}},
			Flag{Name: "search-type", Type: "string", Default: "web", AllowedValues: []string{"web", "image", "video", "news", "discover", "google_news"}},
			Flag{Name: "data-state", Type: "string", Default: "final", AllowedValues: []string{"final", "all", "hourly_all"}},
			Flag{Name: "limit", Type: "int64", Default: "10"},
		)
	}

	return []Command{
		{
			Path:                "sites list",
			Summary:             "List visible Search Console properties",
			Description:         "Lists Search Console properties visible to the configured service account.",
			Flags:               commonFlags(true),
			Examples:            []string{"gsc-cli sites list --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			APIsUsed:            []string{"Sites"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"doctor"},
		},
		{
			Path:        "analytics query",
			Summary:     "Run a custom Search Analytics query",
			Description: "Runs a Search Analytics query with explicit dimensions, date range, row limit, and pagination offset.",
			Flags: append(analyticsFlags(true),
				Flag{Name: "dimensions", Type: "string", Default: "query,page"},
				Flag{Name: "limit", Type: "int64", Default: "1000"},
				Flag{Name: "start-row", Type: "int64", Default: "0"},
			),
			Examples:            []string{"gsc-cli analytics query --site sc-domain:example.com --start 2026-03-01 --end 2026-03-14 --dimensions query,page --output json"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"analytics top-pages", "analytics top-queries", "sync analytics"},
		},
		{
			Path:                "analytics top-pages",
			Summary:             "Show top pages",
			Description:         "Runs a page-grouped Search Analytics query over a recent or explicit date range.",
			Flags:               topAnalyticsFlags(),
			Examples:            []string{"gsc-cli analytics top-pages --site sc-domain:example.com --days 28 --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"analytics query", "report overview"},
		},
		{
			Path:                "analytics top-queries",
			Summary:             "Show top queries",
			Description:         "Runs a query-grouped Search Analytics query over a recent or explicit date range.",
			Flags:               topAnalyticsFlags(),
			Examples:            []string{"gsc-cli analytics top-queries --site sc-domain:example.com --days 28 --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"analytics query", "report overview"},
		},
		{
			Path:                "doctor",
			Summary:             "Check auth and property access",
			Description:         "Verifies credentials, client initialization, property visibility, and optional configured site access.",
			Flags:               commonFlags(true),
			Examples:            []string{"gsc-cli doctor --site sc-domain:example.com --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			APIsUsed:            []string{"Sites"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"sites list", "cache status"},
		},
		{
			Path:        "report overview",
			Summary:     "One-shot operator snapshot",
			Description: "Combines top pages, top queries, sitemap status, and one URL inspection into a single report.",
			Flags: append(commonFlags(true),
				Flag{Name: "start", Type: "string"},
				Flag{Name: "end", Type: "string"},
				Flag{Name: "days", Type: "string", AllowedValues: []string{"7", "28", "90"}},
				Flag{Name: "limit", Type: "int64", Default: "10"},
				Flag{Name: "search-type", Type: "string", Default: "web", AllowedValues: []string{"web", "image", "video", "news", "discover", "google_news"}},
				Flag{Name: "data-state", Type: "string", Default: "final", AllowedValues: []string{"final", "all", "hourly_all"}},
				Flag{Name: "inspect-url", Type: "string"},
			),
			Examples:            []string{"gsc-cli report overview --site sc-domain:example.com --days 28 --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics", "Sitemaps", "URL Inspection"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"report compare", "inspect url", "sitemaps list"},
		},
		{
			Path:        "report compare",
			Summary:     "Compare two equal periods",
			Description: "Compares the current period against the immediately preceding equal-length period and reports summary deltas plus top detail rows.",
			Flags: append(commonFlags(true),
				Flag{Name: "start", Type: "string"},
				Flag{Name: "end", Type: "string"},
				Flag{Name: "days", Type: "string", AllowedValues: []string{"7", "28", "90"}},
				Flag{Name: "limit", Type: "int64", Default: "10"},
				Flag{Name: "search-type", Type: "string", Default: "web", AllowedValues: []string{"web", "image", "video", "news", "discover", "google_news"}},
				Flag{Name: "data-state", Type: "string", Default: "final", AllowedValues: []string{"final", "all", "hourly_all"}},
				Flag{Name: "dimension", Type: "string", Default: "page", AllowedValues: []string{"page", "query"}},
			),
			Examples:            []string{"gsc-cli report compare --site sc-domain:example.com --days 28 --dimension page --output json"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"report overview", "diagnose drop"},
		},
		{
			Path:        "cache status",
			Summary:     "Show cache/store status",
			Description: "Shows the local cache file path and record counts for analytics rows, pull logs, and inspections.",
			Flags: []Flag{
				Flag{Name: "output", Type: "string", Default: "json", AllowedValues: outputModes},
				Flag{Name: "cache-file", Type: "string", Default: ".gsc-cli-cache.json", EnvFallback: "GSC_CLI_CACHE_FILE"},
			},
			Examples:        []string{"gsc-cli cache status --output table"},
			OutputModes:     outputModes,
			Safety:          SafetyReadOnly,
			RelatedCommands: []string{"sync analytics", "sync hourly"},
		},
		{
			Path:        "sync analytics",
			Summary:     "Sync Search Analytics into local cache",
			Description: "Fetches paginated Search Analytics rows and stores them in the local cache with pull-log metadata.",
			Flags: append(commonFlags(true),
				Flag{Name: "period", Type: "string", Default: "last-28d"},
				Flag{Name: "dimensions", Type: "string", Default: "page,query"},
				Flag{Name: "search-type", Type: "string", Default: "web", AllowedValues: []string{"web", "image", "video", "news", "discover", "google_news"}},
				Flag{Name: "data-state", Type: "string", Default: "final", AllowedValues: []string{"final", "all", "hourly_all"}},
				Flag{Name: "cache-file", Type: "string", Default: ".gsc-cli-cache.json", EnvFallback: "GSC_CLI_CACHE_FILE"},
				Flag{Name: "max-rows", Type: "int", Default: "0"},
			),
			Examples:            []string{"gsc-cli sync analytics --site sc-domain:example.com --period last-90d --dimensions query,page --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"cache status", "sync hourly", "analytics query"},
		},
		{
			Path:        "sync hourly",
			Summary:     "Sync provisional hourly rows",
			Description: "Fetches hourly Search Analytics rows for recent days and stores them as provisional data.",
			Flags: append(commonFlags(true),
				Flag{Name: "last", Type: "int", Default: "10"},
				Flag{Name: "cache-file", Type: "string", Default: ".gsc-cli-cache.json", EnvFallback: "GSC_CLI_CACHE_FILE"},
			),
			Examples:            []string{"gsc-cli sync hourly --site sc-domain:example.com --last 2 --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			Caveats:             []string{"Hourly data is provisional and can change.", "The HOUR dimension cannot be combined with other dimensions."},
			RelatedCommands:     []string{"monitor hourly", "cache status"},
		},
		{
			Path:        "diagnose drop",
			Summary:     "Explain likely causes behind traffic drops",
			Description: "Compares current and baseline periods across segments and computes expected-click contribution to rank likely loss drivers.",
			Flags: append(commonFlags(true),
				Flag{Name: "current", Type: "string", Required: true},
				Flag{Name: "baseline", Type: "string", Default: "previous-period"},
				Flag{Name: "segments", Type: "string", Default: "page"},
				Flag{Name: "limit", Type: "int", Default: "5"},
				Flag{Name: "min-loss", Type: "float64", Default: "1"},
				Flag{Name: "search-type", Type: "string", Default: "web", AllowedValues: []string{"web", "image", "video", "news", "discover", "google_news"}},
			),
			Examples:            []string{"gsc-cli diagnose drop --site sc-domain:example.com --current 2026-06-01:2026-06-14 --baseline previous-period --segments page,query,device --output json"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			Caveats:             []string{"Treat fresh/hourly data as provisional if you later adapt this workflow for near-real-time monitoring."},
			RelatedCommands:     []string{"report compare", "opportunities", "monitor hourly"},
		},
		{
			Path:                "monitor hourly",
			Summary:             "Compare today's hourly performance with previous week",
			Description:         "Compares hourly rows for the chosen date against the same day one week earlier.",
			Flags:               append(commonFlags(true), Flag{Name: "date", Type: "string"}),
			Examples:            []string{"gsc-cli monitor hourly --site sc-domain:example.com --date 2026-06-28 --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			Caveats:             []string{"Hourly data is provisional and can change."},
			RelatedCommands:     []string{"sync hourly", "diagnose drop"},
		},
		{
			Path:        "audit indexability",
			Summary:     "Inspect a batch of URLs and classify issues",
			Description: "Runs URL Inspection for a list of URLs and classifies indexability issues such as robots blocks, noindex, 404, 5xx, and canonical mismatches.",
			Flags: append(commonFlags(true),
				Flag{Name: "urls-file", Type: "string", Required: true},
				Flag{Name: "max-urls", Type: "int", Default: "100"},
				Flag{Name: "concurrency", Type: "int", Default: "1"},
				Flag{Name: "cache-file", Type: "string", Default: ".gsc-cli-cache.json", EnvFallback: "GSC_CLI_CACHE_FILE"},
			),
			Examples:            []string{"gsc-cli audit indexability --site sc-domain:example.com --urls-file urls.txt --max-urls 20 --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"URL Inspection"},
			Safety:              SafetyReadOnly,
			Caveats:             []string{"Keep concurrency at 1 unless the CLI explicitly supports more; URL Inspection quotas are limited."},
			RelatedCommands:     []string{"inspect batch", "inspect url", "audit sitemap"},
		},
		{
			Path:        "audit sitemap",
			Summary:     "Compare a sitemap against Search Analytics pages",
			Description: "Fetches a live sitemap and compares it against Search Analytics page impressions.",
			Flags: append(commonFlags(true),
				Flag{Name: "sitemap-url", Type: "string", Required: true},
				Flag{Name: "period", Type: "string", Default: "last-90d"},
				Flag{Name: "limit", Type: "int", Default: "25000"},
			),
			Examples:            []string{"gsc-cli audit sitemap --site sc-domain:example.com --sitemap-url https://example.com/sitemap.xml --output json"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics", "Sitemaps"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"sitemaps list", "sitemaps submit", "audit indexability"},
		},
		{
			Path:        "opportunities",
			Summary:     "Rank near-win and CTR-gap opportunities",
			Description: "Finds near-win, CTR-gap, and decay opportunities from Search Analytics query/page data.",
			Flags: append(commonFlags(true),
				Flag{Name: "period", Type: "string", Default: "last-90d"},
				Flag{Name: "types", Type: "string", Default: "near-win,ctr-gap"},
				Flag{Name: "min-impressions", Type: "float64", Default: "100"},
				Flag{Name: "limit", Type: "int", Default: "20"},
			),
			Examples:            []string{"gsc-cli opportunities --site sc-domain:example.com --period last-90d --types near-win,ctr-gap,decay --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Search Analytics"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"diagnose drop", "analytics query", "report compare"},
		},
		{
			Path:                "sitemaps list",
			Summary:             "List sitemap status for a property",
			Description:         "Lists sitemap resources known to Search Console for the property.",
			Flags:               commonFlags(true),
			Examples:            []string{"gsc-cli sitemaps list --site sc-domain:example.com --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Sitemaps"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"sitemaps get", "audit sitemap"},
		},
		{
			Path:                "sitemaps get",
			Summary:             "Inspect one sitemap resource",
			Description:         "Gets one sitemap resource by feedpath from Search Console.",
			Flags:               append(commonFlags(true), Flag{Name: "feedpath", Type: "string", Required: true}),
			Examples:            []string{"gsc-cli sitemaps get --site sc-domain:example.com --feedpath https://example.com/sitemap.xml --output json"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Sitemaps"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"sitemaps list", "sitemaps submit", "sitemaps delete"},
		},
		{
			Path:                "sitemaps submit",
			Summary:             "Submit a sitemap to Search Console",
			Description:         "Submits a sitemap feedpath to Search Console. This is one of the only write-capable commands.",
			Flags:               append(commonFlags(true), Flag{Name: "feedpath", Type: "string", Required: true}),
			Examples:            []string{"gsc-cli sitemaps submit --site sc-domain:example.com --feedpath https://example.com/sitemap.xml --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Sitemaps"},
			Safety:              SafetyWriteCapable,
			WriteScopeRequired:  true,
			Caveats:             []string{"This command requests write scope for the Search Console API."},
			RelatedCommands:     []string{"sitemaps list", "sitemaps get", "sitemaps delete"},
		},
		{
			Path:                "sitemaps delete",
			Summary:             "Delete a sitemap from Search Console",
			Description:         "Deletes a sitemap feedpath from Search Console. This is one of the only write-capable commands.",
			Flags:               append(commonFlags(true), Flag{Name: "feedpath", Type: "string", Required: true}),
			Examples:            []string{"gsc-cli sitemaps delete --site sc-domain:example.com --feedpath https://example.com/sitemap.xml --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"Sitemaps"},
			Safety:              SafetyWriteCapable,
			WriteScopeRequired:  true,
			Caveats:             []string{"This command requests write scope for the Search Console API."},
			RelatedCommands:     []string{"sitemaps list", "sitemaps get", "sitemaps submit"},
		},
		{
			Path:        "inspect url",
			Summary:     "Inspect one URL for indexing state",
			Description: "Runs URL Inspection for one URL.",
			Flags: append(commonFlags(true),
				Flag{Name: "url", Type: "string", Required: true},
				Flag{Name: "language", Type: "string", Default: "en-US"},
			),
			Examples:            []string{"gsc-cli inspect url --site sc-domain:example.com --url https://example.com/ --output json"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"URL Inspection"},
			Safety:              SafetyReadOnly,
			RelatedCommands:     []string{"inspect batch", "audit indexability"},
		},
		{
			Path:        "inspect batch",
			Summary:     "Inspect many URLs",
			Description: "Runs URL Inspection for a batch of URLs from a file and stores results in the local cache.",
			Flags: append(commonFlags(true),
				Flag{Name: "urls-file", Type: "string", Required: true},
				Flag{Name: "max-urls", Type: "int", Default: "100"},
				Flag{Name: "concurrency", Type: "int", Default: "1"},
				Flag{Name: "cache-file", Type: "string", Default: ".gsc-cli-cache.json", EnvFallback: "GSC_CLI_CACHE_FILE"},
			),
			Examples:            []string{"gsc-cli inspect batch --site sc-domain:example.com --urls-file urls.txt --max-urls 20 --output table"},
			OutputModes:         outputModes,
			RequiresCredentials: true,
			RequiresSite:        true,
			APIsUsed:            []string{"URL Inspection"},
			Safety:              SafetyReadOnly,
			Caveats:             []string{"Keep concurrency at 1 unless the CLI explicitly supports more; URL Inspection quotas are limited."},
			RelatedCommands:     []string{"inspect url", "audit indexability"},
		},
		{
			Path:            "version",
			Summary:         "Print build and version information",
			Description:     "Prints version/build metadata for the installed CLI.",
			Flags:           []Flag{{Name: "output", Type: "string", Default: "table", AllowedValues: outputModes}},
			Examples:        []string{"gsc-cli version --output json"},
			OutputModes:     outputModes,
			Safety:          SafetyReadOnly,
			RelatedCommands: []string{"commands"},
		},
		{
			Path:            "help",
			Summary:         "Show human-readable help",
			Description:     "Prints human-readable usage for the CLI, a command group, or a subcommand. Agents should prefer `commands --output json` for structured discovery.",
			Examples:        []string{"gsc-cli help analytics query", "gsc-cli help report compare"},
			OutputModes:     []string{"text"},
			Safety:          SafetyReadOnly,
			Caveats:         []string{"Topic and subcommand are positional arguments, not flags."},
			RelatedCommands: []string{"commands", "command"},
		},
		{
			Path:            "commands",
			Summary:         "List all public commands in machine-readable form",
			Description:     "Lists all public gsc-cli commands with summaries, safety markers, flags, examples, and related metadata for agent/tool discovery.",
			Flags:           commonOutput,
			Examples:        []string{"gsc-cli commands --output json"},
			OutputModes:     outputModes,
			Safety:          SafetyReadOnly,
			RelatedCommands: []string{"command", "help"},
		},
		{
			Path:            "command",
			Summary:         "Show one command manifest entry",
			Description:     "Shows the machine-readable manifest entry for one command path, such as `diagnose drop` or `audit sitemap`.",
			Flags:           commonOutput,
			Examples:        []string{"gsc-cli command diagnose drop --output json", "gsc-cli command audit sitemap --output json"},
			OutputModes:     outputModes,
			Safety:          SafetyReadOnly,
			Caveats:         []string{"The command path is positional and comes before flags, for example `gsc-cli command diagnose drop --output json`."},
			RelatedCommands: []string{"commands", "help"},
		},
	}
}

func credentialsFlag() Flag {
	return Flag{Name: "credentials-file", Type: "string", EnvFallback: "GSC_CLI_CREDENTIALS_FILE"}
}

func siteFlag() Flag {
	return Flag{Name: "site", Type: "string", EnvFallback: "GSC_CLI_SITE_URL"}
}

func outputFlag() Flag {
	return Flag{Name: "output", Type: "string", Default: "json", AllowedValues: []string{"json", "table"}}
}
