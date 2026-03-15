# P1-S3 - Fix Query Flag Parsing And Document System Install

**Epic:** Phase 1 — CLI Foundation  
**Status:** Done  
**Priority:** P1  
**Estimated Effort:** <0.5 day  
**Dependencies:** P1-S1 complete

## 1. Summary

Fix the `analytics query` subcommand so it correctly parses the first flag in its
argument list, then document and perform a system-wide local install.

## 2. Acceptance Criteria

- [x] AC-1: `gsc-cli analytics query` no longer drops the first flag
- [x] AC-2: regression coverage exists for the parsing bug
- [x] AC-3: `README.md` documents the system install path
- [x] AC-4: docs in `docs/` document the install workflow
- [x] AC-5: the binary is installed and runnable from the machine's writable global bin path

## 3. Implementation Notes

Observed on this host:

- `/usr/local/bin` is root-owned and not writable
- `/opt/homebrew/bin` is writable and already first in `PATH`

The install workflow was therefore normalized to:

```bash
install -m 755 ./gsc-cli \"$(brew --prefix)/bin/gsc-cli\"
```

Validation completed on `2026-03-15`:

- `gsc-cli analytics query --site sc-domain:futurelab.studio --start 2026-02-15 --end 2026-03-14 --dimensions query --limit 5 --output table`
- `command -v gsc-cli`
- `gsc-cli --help`
