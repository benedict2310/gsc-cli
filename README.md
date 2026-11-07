# gsc-cli

`gsc-cli` is a small Go CLI for Google Search Console diagnostics and SEO intelligence.

It is intentionally narrow:
- service-account only
- no browser OAuth
- no token persistence
- read-only by default
- sitemap submit/delete are the only write-capable operations and request write scope only for those commands
- no MCP server in v1

## Commands

- `gsc-cli sites list`
- `gsc-cli analytics query`
- `gsc-cli analytics top-pages`
- `gsc-cli analytics top-queries`
- `gsc-cli help <topic>`
- `gsc-cli report overview`
- `gsc-cli report compare`
- `gsc-cli cache status`
- `gsc-cli sync analytics`
- `gsc-cli sync hourly`
- `gsc-cli diagnose drop`
- `gsc-cli monitor hourly`
- `gsc-cli audit indexability`
- `gsc-cli audit sitemap`
- `gsc-cli opportunities`
- `gsc-cli sitemaps list|get|submit|delete`
- `gsc-cli inspect url|batch`
- `gsc-cli commands`
- `gsc-cli command <command-path>`

## Agent Discovery

Agents and automation should start with the machine-readable command manifest instead of scraping help text:

```bash
gsc-cli commands --output json
```

Inspect one command by path:

```bash
gsc-cli command diagnose drop --output json
gsc-cli command audit sitemap --output json
```

The manifest includes flags, examples, required credentials/site, Search Console API families, output modes, safety markers, and whether write scope is required. Only `sitemaps submit` and `sitemaps delete` are write-capable.

## Setup

1. Create a dedicated Google Cloud project for Search Console access.
2. Enable the Search Console API.
3. Create a dedicated service account.
4. Download the service account JSON key.
5. Add the service account email as a user on the Search Console property you want to query.

Recommended property format:
- `sc-domain:example.com`

Keep the service-account JSON key out of the repository. The project `.gitignore`
already excludes common local key filenames such as `service-account*.json`.

## Build

```bash
go build ./...
```

## Install

Install directly from the module:

```bash
go install github.com/benedict2310/gsc-cli/cmd/gsc-cli@latest
```

Or build and install locally to the Homebrew prefix bin directory:

```bash
go build -o ./gsc-cli ./cmd/gsc-cli
install -m 755 ./gsc-cli "$(brew --prefix)/bin/gsc-cli"
```

Verify:

```bash
command -v gsc-cli
gsc-cli --help
gsc-cli commands --output json
```

## Start Here

If you only know the property and credentials, start with:

```bash
gsc-cli report overview \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --output table
```

## Usage

You can pass credentials and site on the command line:

```bash
gsc-cli sites list \
  --credentials-file ./service-account.json
```

```bash
gsc-cli analytics query \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --start 2026-03-01 \
  --end 2026-03-14 \
  --dimensions query,page \
  --search-type web \
  --output table
```

```bash
gsc-cli analytics top-pages \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --output table
```

```bash
gsc-cli analytics top-queries \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --output table
```

```bash
gsc-cli report overview \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --days 28 \
  --limit 10 \
  --output table
```

```bash
gsc-cli report compare \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --days 28 \
  --dimension page \
  --output table
```

```bash
gsc-cli help analytics query
```

```bash
gsc-cli commands --output json
```

```bash
gsc-cli command opportunities --output json
```

```bash
gsc-cli cache status --cache-file .gsc-cli-cache.json --output table
```

```bash
gsc-cli sync analytics \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --period last-28d \
  --dimensions page,query
```

```bash
gsc-cli sync hourly \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --last 10
```

```bash
gsc-cli diagnose drop \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --current 2026-06-01:2026-06-14 \
  --baseline previous-period \
  --segments page,query,device \
  --output table
```

```bash
gsc-cli monitor hourly \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --output table
```

```bash
gsc-cli audit sitemap \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --sitemap-url https://example.com/sitemap.xml \
  --period last-90d \
  --output table
```

```bash
gsc-cli opportunities \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --period last-90d \
  --types near-win,ctr-gap,decay \
  --output table
```

```bash
gsc-cli sitemaps list \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com
```

```bash
gsc-cli inspect url \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --url https://example.com/
```

```bash
gsc-cli inspect batch \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --urls-file urls.txt \
  --max-urls 100 \
  --output table
```

```bash
gsc-cli doctor \
  --credentials-file ./service-account.json \
  --site sc-domain:example.com \
  --output table
```

```bash
gsc-cli version --output json
```

You can also use environment variables:

```bash
export GSC_CLI_CREDENTIALS_FILE=./service-account.json
export GSC_CLI_SITE_URL=sc-domain:example.com
```

Then run:

```bash
gsc-cli analytics query --start 2026-03-01 --end 2026-03-14
```

## Output

Supported output modes:
- `json` (default)
- `table`

For `top-pages`, `top-queries`, and `report overview`:
- `--days` supports `7`, `28`, or `90`
- default is `28`
- explicit `--start` and `--end` take precedence over `--days`
- without explicit dates, the window ends on the latest day Search Console has data for (final data usually lags 2-3 days), as do `report compare` and the `--period` windows of `sync analytics`, `audit sitemap` and `opportunities`

For `report compare`:
- compares the current period against the immediately preceding equal-length period
- supports `--dimension page` and `--dimension query`

## Validation

```bash
go test ./...
go test -race ./...
go vet ./...
```
