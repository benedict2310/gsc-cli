# P1-S1 - Bootstrap Read-Only Service-Account CLI Foundation

**Epic:** Phase 1 — CLI Foundation  
**Status:** Done  
**Priority:** P1  
**Estimated Effort:** 1-2 days  
**Dependencies:** Google Search Console API enabled; dedicated service account created and added to the `futurelab.studio` property  
**Target:** `/Users/bene/Dev-Source-NoBackup/gsc-cli`  
**Reference Inputs:**  
- `surendranb/google-search-console-mcp` architecture pattern (service-account only, thin API layer)  
- Google Search Console API via `google.golang.org/api/webmasters/v3`

---

## 1. Summary

Create a new standalone Go project named `gsc-cli` that provides a local, read-only command-line interface for Google Search Console. The tool must avoid browser OAuth, token persistence, and MCP in v1. Authentication must use a dedicated service account JSON key only.

## 2. Architecture Context and Reuse Guidance

This project is intentionally **not** part of `htmlctl` and **not** an `htmlctl` extension.

Architectural constraints:

- **Standalone project:** separate module, separate binary, separate docs.
- **CLI-first:** expose a normal command-line interface first; do not introduce MCP protocol support in v1.
- **Read-only only:** restrict to Search Console read access. No sitemap submission, site add/remove, or other write operations.
- **Service-account only:** no interactive OAuth, no embedded OAuth client IDs, no refresh tokens, no keychain integration, no token cache.
- **Small surface area:** keep the code easy to audit; prefer a thin wrapper over the official Google API client.

The main external architecture worth copying from existing tools is:

- environment or flag-supplied service-account JSON path
- flag-supplied or env-supplied site/property
- thin client wrapper over the Search Console API

The following external patterns should **not** be copied:

- browser-based OAuth flows
- local token encryption/keyring layers
- multi-account vaults
- MCP server transports
- write-capable Search Console operations

## 3. Proposed Changes and Architecture Improvements

### Core shape

Create a single Go module with:

- one CLI binary: `gsc-cli`
- one internal config package
- one internal Search Console client package
- one small output layer for deterministic JSON/table responses

### Commands for v1

Implement only these commands:

1. `gsc-cli sites list`
2. `gsc-cli analytics query`
3. `gsc-cli sitemaps list`
4. `gsc-cli inspect url`

### Authentication model

Support:

- `--credentials-file`
- `--site`

with env fallbacks:

- `GSC_CLI_CREDENTIALS_FILE`
- `GSC_CLI_SITE_URL`

Authentication must load a Google service-account JSON file and use only the read-only Search Console scope.

### Output model

Support:

- `--output json`
- `--output table`

`json` should be the safest/default machine-readable mode. Table output may be a human-friendly formatter on top of the same normalized structs.

### Error handling

Errors must be:

- concise
- deterministic
- safe to print

Do not dump credential paths, raw Google error payloads, or stack traces unless an explicit debug flag is later added.

### Explicit non-goals for v1

- no MCP server mode
- no Bing
- no GA4
- no service-account creation helpers
- no config file persistence
- no write actions

## 4. File Touch List

Create these files:

- `go.mod`
  - module bootstrap and dependency declaration
- `cmd/gsc-cli/main.go`
  - command parsing and top-level process orchestration
- `internal/config/config.go`
  - flags/env loading and validation
- `internal/config/config_test.go`
  - config validation tests
- `internal/searchconsole/client.go`
  - Search Console client construction and API wrappers
- `internal/searchconsole/client_test.go`
  - request-shaping and error-handling tests with mocked transport
- `internal/output/output.go`
  - shared JSON/table output formatting
- `internal/output/output_test.go`
  - deterministic output tests
- `README.md`
  - operator setup, service-account configuration, example commands

Optional if needed during implementation:

- `internal/searchconsole/types.go`
  - normalized result structs for CLI output
- `internal/searchconsole/normalize.go`
  - mapping Google API responses into stable CLI shapes

## 5. Implementation Steps

1. Initialize the Go module and add the official Google API dependency.
2. Implement config loading:
   - require credentials file for all commands
   - require site for all site-scoped commands
   - allow `sites list` without a default site
3. Build the Search Console client wrapper using service-account credentials and the read-only scope.
4. Implement `sites list`.
5. Implement `analytics query` with:
   - date range
   - dimension list
   - row limit
   - search type validation
6. Implement `sitemaps list`.
7. Implement `inspect url`.
8. Normalize results into stable structs before printing.
9. Add `json` and `table` output modes.
10. Document setup and usage in the project `README.md`.

## 6. Tests and Validation

### Automated

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`

### Required test coverage

- config:
  - missing credentials file rejected
  - missing site rejected where required
  - unsupported output mode rejected
- analytics query:
  - invalid dimension rejected
  - invalid search type rejected
  - invalid date format rejected
  - row limit bounds enforced
- client:
  - service account file loading works
  - API request parameters are shaped correctly
  - upstream errors are sanitized into stable CLI errors
- output:
  - JSON output deterministic
  - table output deterministic

### Manual validation

With a real Futurelab service account:

- `gsc-cli sites list`
- `gsc-cli analytics query --site sc-domain:futurelab.studio --start <date> --end <date>`
- `gsc-cli sitemaps list --site sc-domain:futurelab.studio`
- `gsc-cli inspect url --site sc-domain:futurelab.studio --url https://futurelab.studio/`

## 7. Acceptance Criteria

- [x] AC-1: `gsc-cli` builds as a standalone Go binary in its own repository.
- [x] AC-2: Authentication uses service-account credentials only; no OAuth desktop flow or token persistence exists in v1.
- [x] AC-3: The CLI supports `sites list`, `analytics query`, `sitemaps list`, and `inspect url`.
- [x] AC-4: Output is deterministic and supports at least `json` and `table`.
- [x] AC-5: The tool uses read-only Search Console access only.
- [x] AC-6: Real manual validation succeeds against the Futurelab Search Console property.
- [x] AC-7: Project documentation explains service-account setup and safe usage clearly.

## 9. Implementation Notes

Implemented in `/Users/bene/Dev-Source-NoBackup/gsc-cli`:

- `cmd/gsc-cli/main.go`
- `internal/config/config.go`
- `internal/searchconsole/client.go`
- `internal/searchconsole/types.go`
- `internal/output/output.go`
- `README.md`

Convenience commands added on top of the base story:

- `gsc-cli analytics top-pages`
- `gsc-cli analytics top-queries`

Test coverage added:

- `internal/config/config_test.go`
- `internal/searchconsole/client_test.go`
- `internal/output/output_test.go`

Local validation completed:

- `go build ./...`
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`

Real manual validation completed on `2026-03-15` against `sc-domain:futurelab.studio`:

- `sites list`
  - property visible with `siteFullUser`
- `analytics query --site sc-domain:futurelab.studio --start 2026-03-01 --end 2026-03-14 --dimensions page --limit 10`
  - returned production page impression rows for `futurelab.studio`
- `sitemaps list --site sc-domain:futurelab.studio`
  - returned `https://futurelab.studio/sitemap.xml` with `0` warnings and `0` errors
- `inspect url --site sc-domain:futurelab.studio --url https://futurelab.studio/`
  - returned `PASS`
  - `Submitted and indexed`
  - canonical and sitemap data present

## 8. Risks and Open Questions

### Risks

- **Risk:** service-account permissions are configured too broadly.  
  **Mitigation:** document least-privilege setup and use a dedicated Futurelab service account only.

- **Risk:** Google API error payloads leak noisy or path-sensitive details into CLI output.  
  **Mitigation:** normalize and sanitize errors before printing.

- **Risk:** v1 scope expands into a generic SEO Swiss-army knife.  
  **Mitigation:** keep the project tightly bounded to read-only Search Console primitives.

### Open Questions

- Should `inspect url` ship in v1 if it requires different response normalization complexity than the other commands?
- Should v1 prefer domain properties (`sc-domain:futurelab.studio`) over URL-prefix properties in docs/examples, while still supporting both?
