# gsc-cli

`gsc-cli` is a small, read-only Go CLI for Google Search Console.

It is intentionally narrow:
- service-account only
- no browser OAuth
- no token persistence
- no write operations
- no MCP server in v1

## Commands

- `gsc-cli sites list`
- `gsc-cli analytics query`
- `gsc-cli analytics top-pages`
- `gsc-cli analytics top-queries`
- `gsc-cli help <topic>`
- `gsc-cli report overview`
- `gsc-cli report compare`
- `gsc-cli sitemaps list`
- `gsc-cli inspect url`

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

Build and install locally to the Homebrew prefix bin directory:

```bash
go build -o ./gsc-cli ./cmd/gsc-cli
install -m 755 ./gsc-cli "$(brew --prefix)/bin/gsc-cli"
```

Verify:

```bash
command -v gsc-cli
gsc-cli --help
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

For `report compare`:
- compares the current period against the immediately preceding equal-length period
- supports `--dimension page` and `--dimension query`

## Validation

```bash
go test ./...
go test -race ./...
go vet ./...
```
