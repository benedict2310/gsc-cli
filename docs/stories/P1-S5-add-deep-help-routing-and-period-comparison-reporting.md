# P1-S5 - Add Deep Help Routing And Period Comparison Reporting

**Epic:** Phase 1 — CLI Foundation  
**Status:** Done  
**Priority:** P1

## 1. Summary

Extend `gsc-cli` with:

- `help <topic>` and `help <topic> <subcommand>`
- `report compare`

This makes the tool more self-explanatory to a cold-start agent and more useful
for repeated SEO review loops.

## 2. Acceptance Criteria

- [x] AC-1: `gsc-cli help analytics query` and similar deep help paths work
- [x] AC-2: `report compare` compares the current period against the previous equal period
- [x] AC-3: `report compare` supports `page` and `query` dimensions
- [x] AC-4: docs in `README.md` and `docs/` explain deep help and compare workflows
- [x] AC-5: live validation succeeds against `futurelab.studio`

## 3. Implementation Notes

Implemented:

- `gsc-cli help <topic>`
- `gsc-cli help <topic> <subcommand>`
- `gsc-cli report compare`

`report compare` currently supports:

- `--dimension page`
- `--dimension query`

Validation completed on `2026-03-15`:

- `./gsc-cli help analytics query`
- `./gsc-cli help report compare`
- `./gsc-cli report compare --site sc-domain:futurelab.studio --days 28 --dimension page --limit 10 --output table`

Observed live result for `futurelab.studio`:

- current 28-day page impressions: `34`
- previous equal window page impressions: `0`
- impression delta: `+34`
