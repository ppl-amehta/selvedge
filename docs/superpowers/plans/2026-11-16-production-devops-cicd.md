# Production DevOps, CI/CD, Containerization & Serverless Hosting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform FPCC Selvedge ERP from a local developer prototype into a hardened, frontier production system with full defense-in-depth CI/CD, multi-stage zero-CVE Dockerization, hybrid high-speed SQLite/LibSQL persistence, and automated GitOps deployment to a free serverless platform.

**Architecture:** A three-tier pipeline: (1) Pre-commit and CI defense-in-depth gates (static analysis, concurrency race detection, vulnerability scanning, Svelte typechecking, and OpenCode audit trail); (2) Multi-stage hardened Alpine container compiling a statically linked CGO SQLite binary with embedded Svelte 5 frontend and non-root execution; (3) Frontier hybrid persistence combining local NVMe SQLite WAL speed with pluggable LibSQL/Turso embedded replica support and Litestream streaming replication, orchestrated via GitOps push-to-main deployment with scale-to-zero serverless machines.

**Tech Stack:** Go 1.27, Svelte 5.57+, Vite 8.3+, Chi v5, SQLite 3 (WAL mode), LibSQL / Turso, Litestream, Docker (multi-stage Alpine 3.21), GitHub Actions, GHCR (GitHub Container Registry), Fly.io (Apps v2), Golangci-lint, Govulncheck, Trivy.

**Spec:** Direct user directive for extreme end-to-end CI/CD hyper-optimization, checking gates, frontier DevOps/GitOps, Dockerization, free serverless hosting, and comprehensive modern framework/API upgrades.

## Global Constraints

- Must maintain 100% backward compatibility for all existing factory floor stations (Order Taking, Inward 1A/1B/1C, Order Prep & Planning, Cutting Station).
- Must preserve sub-millisecond local SQLite WAL and graph query performance (`_txlock=immediate`, `_synchronous=NORMAL`).
- Container image must run as non-root (`appuser` UID 10001) with zero HIGH/CRITICAL CVEs and strict healthcheck integration.
- CI/CD must run with `-race` concurrency detection and fail on any data race or dependency vulnerability.
- All OpenCode backend policy requirements (AGENTS.md) must be satisfied with GLM 5.3 Flash dual-validation and pre-completion code review.

## Review Focus

1. CGO musl static linking for `go-sqlite3`: verify the static binary has zero shared glibc dependencies and runs cleanly on pure Alpine.
2. SQLite file locking on persistent volumes: verify `_txlock=immediate` prevents database locks during concurrent outbox processing.
3. Cold-start latency on scale-to-zero: verify the embedded Svelte 5 SPA and SQLite WAL boot in under 200ms.
4. CI pipeline execution time: verify GitHub Actions jobs run concurrently with layer caching (`type=gha`) to finish in under 3 minutes.
5. Container vulnerability baseline: verify Trivy scan reports zero CRITICAL and zero HIGH vulnerabilities.

---

### Task 1: Framework, Engine & Dependency Upgrades

**Files:**
- Modify: `web/package.json`
- Create: `web/svelte.config.js`
- Create: `web/jsconfig.json`
- Modify: `go.mod`
- Modify: `internal/platform/router/router.go`
- Test: `web/src/App.svelte`, `internal/platform/router/router_test.go`

- [ ] **Step 1.1: Upgrade frontend dependencies & configure Svelte typechecking**
  Update `web/package.json` with latest Svelte 5, Vite, and add `svelte-check` for CI static typing.
  Create `web/svelte.config.js` with `vitePreprocess`.
  Create `web/jsconfig.json` configuring ECMAScript 2024 and strict DOM typechecking.
  Run `npm install` and `npm run check`.

- [ ] **Step 1.2: Upgrade Go dependencies and ensure toolchain alignment**
  Update `go.mod` with latest patch releases (`go-sqlite3` v1.14.53, etc.).
  Run `go mod tidy` and verify all modules compile cleanly.

- [ ] **Step 1.3: Implement robust `/api/health` production probe**
  Enhance `internal/platform/router/router.go` to provide a dedicated health endpoint `/api/health` reporting:
  - System status (`healthy` / `degraded`)
  - SQLite database ping and WAL readiness
  - Memory allocs and goroutine count
  - System uptime and engine version
  Write unit test in `internal/platform/router/router_test.go` verifying 200 OK and JSON schema.

- [ ] **Step 1.4: Run test suite**
  Execute `go test -v ./...` and `cd web && npm run build` to confirm zero regressions.

---

### Task 2: Frontier Hybrid Persistence Engine (SQLite WAL + LibSQL/Turso + Litestream)

**Files:**
- Modify: `internal/platform/db/db.go`
- Create: `deploy/litestream.yml`
- Test: `internal/platform/db/db_test.go`

- [ ] **Step 2.1: Add hybrid SQLite / LibSQL connection abstraction**
  In `internal/platform/db/db.go`, implement environment-aware connection logic:
  - Default: Native SQLite with WAL pragmas (`_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_synchronous=NORMAL&_txlock=immediate`).
  - Remote/Embedded Cloud: If `TURSO_DATABASE_URL` is set, support LibSQL protocol or embedded replica path with fallback.
  - Add connection pool limits and busy-handler guards.

- [ ] **Step 2.2: Add Litestream continuous replication configuration**
  Create `deploy/litestream.yml` configuring streaming replication from `/data/selvedge.db` to any S3-compatible bucket (e.g. Cloudflare R2 / AWS S3) with 1-second snapshot frequency for zero-data-loss disaster recovery.

- [ ] **Step 2.3: Write database connection and migration tests**
  Write tests in `internal/platform/db/db_test.go` verifying WAL pragmas, transaction safety, and migration idempotency.

---

### Task 3: Production Hardened Multi-Stage Dockerfile & Container Architecture

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`
- Create: `docker-compose.yml`
- Create: `docker-compose.prod.yml`

- [ ] **Step 3.1: Create multi-stage hardened Dockerfile**
  - **Stage 1 (Frontend Builder)**: `node:24-alpine` building Svelte 5 assets into `web/dist`.
  - **Stage 2 (Go Compiler)**: `golang:1.24-alpine` with `musl-dev`, `gcc`, `git`, compiling a static CGO binary (`-ldflags="-s -w -extldflags '-static'"`).
  - **Stage 3 (Runtime Environment)**: Minimal `alpine:3.21` with `ca-certificates`, `tzdata`, `wget`.
  - **Security**: Non-root user `appuser:appgroup` (UID/GID 10001:10001), no root permissions.
  - **Storage**: `VOLUME ["/data"]`, `ENV DB_PATH=/data/selvedge.db`.
  - **Healthcheck**: Native `HEALTHCHECK` running `wget --spider -q http://127.0.0.1:4000/api/health || exit 1`.

- [ ] **Step 3.2: Create optimized `.dockerignore`**
  Ignore `.git`, `node_modules`, `.live`, tests, and binary artifacts to ensure minimal context transfer (< 5MB) and optimal layer caching.

- [ ] **Step 3.3: Create production `docker-compose.yml`**
  Configure service with healthcheck, automatic restart (`unless-stopped`), resource limits (`cpus: '1.0'`, `memory: 512M`), and persistent volume `selvedge-data`.

- [ ] **Step 3.4: Test Docker build locally**
  Build image: `docker build -t selvedge:local .` and verify size is ~25MB and binary runs cleanly.

---

### Task 4: Extreme CI/CD Checking Gates (Full Defense-in-Depth)

**Files:**
- Create: `.golangci.yml`
- Create: `.github/workflows/ci.yml`

- [ ] **Step 4.1: Configure modern Golangci-lint**
  Create `.golangci.yml` enabling modern linters:
  - `govet`, `errcheck`, `staticcheck`, `unused`, `gosec`, `revive`, `misspell`, `gocyclo`, `gocritic`.
  - Strict cyclomatic complexity and security checks.

- [ ] **Step 4.2: Author comprehensive GitHub Actions CI pipeline (`ci.yml`)**
  Configure parallel jobs running on every pull request and push:
  - **Job 1: `lint-and-typecheck`**:
    - Go: `golangci-lint` runner.
    - Svelte: `npm run check` and `npm run build`.
  - **Job 2: `concurrency-and-tests`**:
    - `go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...` ensuring zero concurrency race hazards across outbox, graph, and orders.
  - **Job 3: `security-scan`**:
    - `govulncheck` on Go code.
    - Aqua Security `trivy-action` scanning repository and container image for zero HIGH/CRITICAL CVEs.
  - **Job 4: `opencode-audit`**:
    - Validates `.agents/logs/opencode-mcp.log` and surfaces the verification badge.

- [ ] **Step 4.3: Validate CI workflows locally**
  Run linter and tests locally with race detection: `go test -race ./...`.

---

### Task 5: GitOps Deployment & Free Serverless Hosting Automation

**Files:**
- Create: `fly.toml`
- Create: `deploy/koyeb.yaml`
- Create: `deploy/render.yaml`
- Create: `.github/workflows/deploy.yml`

- [ ] **Step 5.1: Create modern Fly.io V2 Machine configuration (`fly.toml`)**
  - Application configuration for Apps v2 architecture.
  - Mount persistent volume `selvedge_data` on `/data` (3GB free tier).
  - Enable true serverless scale-to-zero:
    - `auto_stop_machines = "stop"`
    - `auto_start_machines = true`
    - `min_machines_running = 0`
  - Configure HTTP service on port 4000 with healthcheck on `/api/health`.

- [ ] **Step 5.2: Create alternative free-tier manifests (`koyeb.yaml`, `render.yaml`)**
  Provide declarative configuration for Koyeb and Render free container tiers with persistent disk attachments.

- [ ] **Step 5.3: Create GitOps Push-to-Main CD Pipeline (`deploy.yml`)**
  - Triggers automatically on push to `main` branch.
  - Builds multi-arch container (`linux/amd64`, `linux/arm64`) using GitHub Actions cache (`type=gha`).
  - Pushes image to GitHub Container Registry (`ghcr.io/ppl-amehta/selvedge`).
  - Automatically deploys to Fly.io via `flyctl deploy --remote-only`.
  - Runs post-deployment verification probe against public URL.

---

### Task 6: Pre-Completion Code Review, Verification & Git Synchronization

**Files:**
- All created and modified files

- [ ] **Step 6.1: Run OpenCode MCP Pre-Completion Code Review**
  Invoke `opencode_code_review` (model: `glm-5.3-flash`) across new CI/CD, Dockerfile, and persistence code. Address any critical findings.
- [ ] **Step 6.2: Run full local build, race tests, and container checks**
  Execute `go test -race -v ./...`, `npm run build`, and `npm run check`.
- [ ] **Step 6.3: Commit and push to GitHub**
  Commit all modernized production assets with a structured conventional commit and push to `origin main`.
