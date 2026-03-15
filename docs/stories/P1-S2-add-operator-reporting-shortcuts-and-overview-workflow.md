# P1-S2 - Add Operator Reporting Shortcuts and Overview Workflow

**Epic:** Phase 1 — CLI Foundation  
**Status:** Done  
**Priority:** P1  
**Estimated Effort:** 0.5-1 day  
**Dependencies:** P1-S1 complete; working service-account access to `sc-domain:futurelab.studio`  
**Target:** `/Users/bene/Dev-Source-NoBackup/gsc-cli`

---

## 1. Summary

Extend `gsc-cli` so it is practical for repeated operator use, not just raw API probing.

Add:

1. `--days 7|28|90` shorthand on the analytics top commands
2. `gsc-cli report overview`

The overview command should combine the most important Search Console checks into a single report:

- top pages
- top queries
- sitemap status
- one URL inspection

## 2. Architecture Guidance

Keep the project narrow and operator-centric.

- Reuse the existing Search Console client package.
- Reuse the existing output package.
- Do not add persistence or config files.
- Keep read-only scope only.

## 3. Proposed Changes

### Date shorthand

Support `--days` on:

- `gsc-cli analytics top-pages`
- `gsc-cli analytics top-queries`
- `gsc-cli report overview`

Rules:

- allowed values: `7`, `28`, `90`
- default: `28`
- `--days` cannot be combined with `--start` or `--end`

### Overview report

Add:

- `gsc-cli report overview`

Inputs:

- common flags
- `--days`
- `--start`
- `--end`
- `--limit`
- `--inspect-url` optional

If `--inspect-url` is omitted:

- for `sc-domain:<domain>` infer `https://<domain>/`
- for URL-prefix properties use the property URL itself

## 4. Acceptance Criteria

- [x] AC-1: `analytics top-pages` and `analytics top-queries` support `--days 7|28|90`
- [x] AC-2: invalid `--days` values are rejected deterministically
- [x] AC-3: `report overview` returns top pages, top queries, sitemap rows, and one inspection result
- [x] AC-4: `report overview` supports both `json` and `table`
- [x] AC-5: docs in `docs/` and `README.md` reflect the new workflow
- [x] AC-6: live validation succeeds against `futurelab.studio`

## 5. Implementation Notes

Implemented:

- `--days 7|28|90` on:
  - `gsc-cli analytics top-pages`
  - `gsc-cli analytics top-queries`
  - `gsc-cli report overview`
- `gsc-cli report overview`
- docs index:
  - `docs/README.md`
- operator guide:
  - `docs/guides/operator-workflows.md`

Validation completed on `2026-03-15`:

- `gsc-cli analytics top-pages --site sc-domain:futurelab.studio --days 7 --output table`
- `gsc-cli analytics top-queries --site sc-domain:futurelab.studio --days 90 --output table`
- `gsc-cli report overview --site sc-domain:futurelab.studio --days 28 --limit 5 --output table`

Observed live results:

- top pages returned production rows for `futurelab.studio`
- top queries returned live query rows
- overview returned:
  - date range summary
  - top pages
  - top queries
  - sitemap status with `0` warnings and `0` errors
  - homepage inspection with `PASS`
