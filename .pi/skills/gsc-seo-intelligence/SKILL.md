---
name: gsc-seo-intelligence
description: Use when diagnosing Google Search Console performance, syncing GSC data, finding SEO/content opportunities, auditing sitemap/indexability issues, or producing search intelligence briefs with gsc-cli.
---

# GSC SEO Intelligence

Use `gsc-cli` to diagnose Google Search Console performance and turn Search Analytics, URL Inspection, and sitemap data into actionable SEO/content work.

## When to Use

Use this skill when the user asks to:

- investigate search traffic drops or anomalies
- inspect GSC clicks, impressions, CTR, position, pages, or queries
- sync Search Console data locally
- audit sitemap coverage or submitted sitemaps
- inspect URL indexability in Search Console
- find content opportunities such as near-wins, CTR gaps, or decay
- compare recent/hourly performance
- validate SEO state after publishing a site change

Do **not** use this for generic web analytics without Search Console access. Do not use it as a true rank tracker; GSC average position is an aggregated trend signal.

## Prerequisites

`gsc-cli` should be installed and available on `PATH`:

```bash
gsc-cli version --output table
```

The CLI uses a Google service-account JSON file, not browser OAuth. If auth fails, likely causes are:

- service-account key deleted/disabled
- Search Console API disabled for the Google Cloud project
- service-account email lacks property access in Search Console
- wrong property string, e.g. missing `sc-domain:`

Recommended environment variables:

```bash
export GSC_CLI_CREDENTIALS_FILE=./service-account.json
export GSC_CLI_SITE_URL=sc-domain:example.com
export GSC_CLI_CACHE_FILE="$HOME/.cache/gsc-cli/example.com.json"
```

Never print or commit service-account JSON contents.

## Initial Health Check

Start every unfamiliar property with:

```bash
gsc-cli doctor --output table

gsc-cli sites list --output table
```

Then verify cache location:

```bash
gsc-cli cache status --output table
```

## Local Data Sync

Before deeper analysis, warm the local cache:

```bash
gsc-cli sync analytics \
  --period last-28d \
  --dimensions page,query \
  --max-rows 25000 \
  --output table
```

For broader opportunity analysis:

```bash
gsc-cli sync analytics \
  --period last-90d \
  --dimensions query,page \
  --max-rows 25000 \
  --output table
```

For fresh monitoring, sync hourly data. Hourly data is provisional and can change:

```bash
gsc-cli sync hourly --last 2 --output table
```

## Traffic Drop Diagnosis

Use complete date ranges when possible. Compare a current period against the immediately preceding equal-length period:

```bash
gsc-cli diagnose drop \
  --current YYYY-MM-DD:YYYY-MM-DD \
  --baseline previous-period \
  --segments page,query,device,country,searchAppearance \
  --limit 10 \
  --output table
```

Interpretation guide:

- impressions down, position stable → demand/seasonality/topic interest decline
- impressions down, position worse → ranking loss or algorithmic reassessment
- impressions stable, CTR down → SERP/snippet/title/zero-click issue
- device-only drop → device UX or device SERP difference
- URL-cluster drop → technical/indexing/template/content issue
- site-wide drop → algorithm update, reporting issue, or site-wide technical issue

If using recent/hourly data, clearly call it provisional.

## Hourly Monitoring

Compare a date to the same day in the previous week:

```bash
gsc-cli monitor hourly --date YYYY-MM-DD --output table
```

If `--date` is omitted, it defaults to today. Treat results as early-warning data, not final reporting.

## Content Opportunities

Find near-win, CTR-gap, and decay opportunities:

```bash
gsc-cli opportunities \
  --period last-90d \
  --types near-win,ctr-gap,decay \
  --min-impressions 100 \
  --limit 20 \
  --output table
```

Use the output to prioritize:

- near-win → content expansion, internal links, better topical coverage
- CTR gap → title/meta/snippet angle rewrite
- decay → refresh relevance, update examples, check competitors/technical issues

Do not overfit one short window. Prefer 28d/90d and compare seasonality where available.

## Sitemap Audit

List submitted sitemaps:

```bash
gsc-cli sitemaps list --output table
```

Audit live sitemap coverage against GSC page impressions:

```bash
gsc-cli audit sitemap \
  --sitemap-url https://example.com/sitemap.xml \
  --period last-90d \
  --limit 25000 \
  --output table
```

Look for:

- URLs in sitemap with impressions
- sitemap URLs with no impressions
- pages receiving impressions but missing from sitemap

The sitemap fetcher supports normal XML and gzip-compressed sitemaps.

### Sitemap Writes

`gsc-cli` is read-only by default. These are the only write-capable commands and request write scope only for the command:

```bash
gsc-cli sitemaps submit --feedpath https://example.com/sitemap.xml

gsc-cli sitemaps delete --feedpath https://example.com/sitemap.xml
```

Only submit/delete sitemaps when explicitly requested or when it is clearly part of the user's goal. Never delete a sitemap as a cleanup step without confirmation.

## Indexability Audit

Inspect one URL:

```bash
gsc-cli inspect url \
  --url https://example.com/ \
  --output table
```

Inspect a batch from a file:

```bash
gsc-cli inspect batch \
  --urls-file urls.txt \
  --max-urls 100 \
  --concurrency 1 \
  --output table
```

Equivalent audit entry point:

```bash
gsc-cli audit indexability \
  --urls-file urls.txt \
  --max-urls 100 \
  --concurrency 1 \
  --output table
```

Issue classes include:

- `robots_blocked`
- `noindex_meta`
- `x_robots_noindex`
- `404`
- `5xx`
- `canonical_mismatch`
- `never_crawled`

Respect URL Inspection API quotas. Keep `--concurrency 1` unless the CLI explicitly supports more.

## Post-Publish SEO Validation

After publishing or promoting a site change, run a bounded validation:

```bash
gsc-cli sitemaps list --output table

gsc-cli audit sitemap \
  --sitemap-url https://example.com/sitemap.xml \
  --period last-90d \
  --output table
```

If specific URLs changed, inspect only those URLs:

```bash
gsc-cli inspect batch --urls-file changed-urls.txt --max-urls 20 --output table
```

Do not expect traffic changes immediately after publishing. Schedule or recommend a later follow-up for performance impact.

## Output Guidance

Prefer `--output table` for human-readable summaries and `--output json` when another tool or agent will consume the result.

When summarizing results to the user, include:

- date range and baseline
- whether data is final or provisional
- top contributors/opportunities
- likely cause, not just metric deltas
- recommended next actions
- any caveats about API limits or incomplete/fresh data

## Safe Defaults

- Use service-account auth; do not start OAuth/browser auth flows.
- Do not print credentials.
- Use cache files under `~/.cache/gsc-cli/`.
- Treat hourly/fresh data as provisional.
- Treat average position as a trend signal, not exact rank.
- Avoid write-capable sitemap commands unless explicitly needed.
