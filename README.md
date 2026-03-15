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
- `gsc-cli sitemaps list`
- `gsc-cli inspect url`

## Setup

1. Create a dedicated Google Cloud project for Search Console access.
2. Enable the Search Console API.
3. Create a dedicated service account.
4. Download the service account JSON key.
5. Add the service account email as a user on the Search Console property you want to query.

Recommended property format:
- `sc-domain:futurelab.studio`

Keep the service-account JSON key out of the repository. The project `.gitignore`
already excludes common local key filenames such as `service-account*.json`.

## Build

```bash
go build ./...
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
  --site sc-domain:futurelab.studio \
  --start 2026-03-01 \
  --end 2026-03-14 \
  --dimensions query,page \
  --search-type web \
  --output table
```

```bash
gsc-cli analytics top-pages \
  --credentials-file ./service-account.json \
  --site sc-domain:futurelab.studio \
  --output table
```

```bash
gsc-cli analytics top-queries \
  --credentials-file ./service-account.json \
  --site sc-domain:futurelab.studio \
  --output table
```

```bash
gsc-cli sitemaps list \
  --credentials-file ./service-account.json \
  --site sc-domain:futurelab.studio
```

```bash
gsc-cli inspect url \
  --credentials-file ./service-account.json \
  --site sc-domain:futurelab.studio \
  --url https://futurelab.studio/
```

You can also use environment variables:

```bash
export GSC_CLI_CREDENTIALS_FILE=./service-account.json
export GSC_CLI_SITE_URL=sc-domain:futurelab.studio
```

Then run:

```bash
gsc-cli analytics query --start 2026-03-01 --end 2026-03-14
```

## Output

Supported output modes:
- `json` (default)
- `table`

For `top-pages` and `top-queries`, omitting `--start` and `--end` uses a recent
28-day window ending yesterday.

## Validation

```bash
go test ./...
go test -race ./...
go vet ./...
```
