# Operator Workflows

This guide documents the fastest recurring workflows for `gsc-cli`.

## Environment

```bash
export GSC_CLI_CREDENTIALS_FILE="$HOME/.config/gsc-cli/service-account.json"
export GSC_CLI_SITE_URL="sc-domain:example.com"
```

## System Install

Build and install to the Homebrew prefix bin directory:

```bash
cd /path/to/gsc-cli
go build -o ./gsc-cli ./cmd/gsc-cli
install -m 755 ./gsc-cli "$(brew --prefix)/bin/gsc-cli"
```

Verify:

```bash
command -v gsc-cli
gsc-cli --help
```

## Top Pages

Recent 28-day window:

```bash
gsc-cli analytics top-pages --output table
```

Explicit 7-day window:

```bash
gsc-cli analytics top-pages --days 7 --output table
```

## Top Queries

Recent 90-day window:

```bash
gsc-cli analytics top-queries --days 90 --output table
```

## Overview Report

Use this when you want one combined operator snapshot:

```bash
gsc-cli report overview --days 28 --limit 10 --output table
```

The overview report includes:

- top pages
- top queries
- sitemap status
- one URL inspection

If `--inspect-url` is not supplied, `gsc-cli` infers a homepage URL from the configured property when possible.

## Compare Report

Use this when you want to know whether visibility is improving or declining
versus the previous equal period:

```bash
gsc-cli report compare --days 28 --dimension page --limit 10 --output table
```

You can also compare query visibility:

```bash
gsc-cli report compare --days 28 --dimension query --limit 10 --output table
```

## Deep Help

Use topic help when an agent or operator needs a command without opening the docs:

```bash
gsc-cli help analytics query
gsc-cli help report compare
```

## Doctor

Check whether credentials, visible properties, and the configured site are usable:

```bash
gsc-cli doctor --output table
```

## Version

Print build metadata in machine-readable form:

```bash
gsc-cli version --output json
```
