# P1-S4 - Improve Cold-Start Agent Discoverability And Command Ergonomics

**Epic:** Phase 1 — CLI Foundation  
**Status:** Done  
**Priority:** P1

## 1. Summary

Harden `gsc-cli` for cold-start use by an AI agent or operator with minimal prior
context.

This pass focuses on:

- non-error help flows
- clearer root guidance
- actionable config errors
- removing the `--days 0` trap
- adding `doctor`
- adding `version`
- making docs generic instead of machine-specific

## 2. Acceptance Criteria

- [x] AC-1: `--help` on subcommands exits successfully and is not reported as an error
- [x] AC-2: root help contains a recommended starting command and command descriptions
- [x] AC-3: config errors explain the relevant flag or env var to set
- [x] AC-4: explicit `--start` and `--end` override `--days` without requiring `--days 0`
- [x] AC-5: `gsc-cli doctor` exists and reports credential and site readiness
- [x] AC-6: `gsc-cli version` exists
- [x] AC-7: `README.md` and `docs/` are generic and agent-friendly

## 3. Implementation Notes

Implemented:

- non-error subcommand help handling via `flag.ErrHelp` passthrough
- stronger root help with:
  - start-here recommendation
  - command descriptions
- actionable config errors for missing site and credentials
- `--days` parsing changed to an optional flag so explicit date ranges no longer
  require `--days 0`
- `gsc-cli doctor`
- `gsc-cli version`
- generic docs and examples using `example.com` placeholders

Validation completed on `2026-03-15`:

- `./gsc-cli --help`
- `./gsc-cli analytics --help`
- `./gsc-cli analytics query --help`
- `./gsc-cli doctor --help`
- `./gsc-cli version --output table`
- `./gsc-cli analytics top-pages --site sc-domain:futurelab.studio --start 2026-02-15 --end 2026-03-14 --limit 5 --output table`
- `./gsc-cli doctor --site sc-domain:futurelab.studio --output table`
