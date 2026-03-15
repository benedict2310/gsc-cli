# Operator Workflows

This guide documents the fastest recurring workflows for `gsc-cli`.

## Environment

```bash
export GSC_CLI_CREDENTIALS_FILE=/Users/bene/Dev-Source-NoBackup/gsc-cli/service-account.futurelab.local.json
export GSC_CLI_SITE_URL=sc-domain:futurelab.studio
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
