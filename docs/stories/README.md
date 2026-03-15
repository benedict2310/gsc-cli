# gsc-cli — Story Map

Standalone Go CLI for read-only Google Search Console analysis.

**Project goal:** local operator tool for querying Search Console safely with a dedicated service account. No OAuth desktop flow, no token persistence, no website runtime, no MCP in v1.

---

## Phase 1 — CLI Foundation

**Goal:** bootstrap a small, auditable Go CLI that can authenticate with a service account and query core Search Console data for a configured property.

| # | Story | Status |
|---|-------|--------|
| P1-S1 | Bootstrap read-only service-account CLI foundation | Done |
| P1-S2 | Add operator reporting shortcuts and overview workflow | Done |
| P1-S3 | Fix query flag parsing and document system install | Done |

**Done when:** `gsc-cli` can list sites, query search analytics, list sitemaps, and inspect a URL using service-account credentials with deterministic CLI output.
