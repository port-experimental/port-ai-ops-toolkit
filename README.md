# AI Consumption in Port — Go Ingestion (GitHub Copilot + Microsoft 365 Copilot)

This repository bundles everything needed to ingest GitHub Copilot seats + usage and Microsoft 365 Copilot usage into Port with a Go worker and opinionated dashboards.

## Repository layout
- `docs/` — plain-language guides (runbook, tokens, dashboard, validation, references).
- `configs/blueprints/` — Port blueprint JSON files to apply via Builder/API.
- `configs/mappings/` — mapping override + webhook payload definitions for seats and M365.
- `worker/go-ingestor/` — Go implementation (`main.go`, `go.mod`, `config.example.env`, Dockerfile, compiled helper binary).
- `deploy/` — automation examples for GitHub Actions and a Kubernetes CronJob.

## Recommended reading order
1. `docs/runbook.md` — step-by-step flow (blueprints → mappings → secrets → deploy).
2. `docs/tokens-and-permissions.md` — scopes, roles, token generation.
3. `worker/go-ingestor/` — configure `.env`, run locally, or build the container image.
4. `deploy/` — schedule the worker (GitHub Actions or CronJob).
5. `docs/dashboard.md` — Port dashboard widgets for AI consumption insights.
6. `docs/differences-vs-port-default.md` — how this package extends Port’s defaults.
7. `docs/validation-and-guardrails.md` — validation, privacy, ops guardrails.
8. `docs/references.md` — API references for GitHub, Microsoft Graph, and Port.

## Quick start
1. Apply the JSON in `configs/blueprints/`, then wire up Port data sources with the files in `configs/mappings/`.
2. Copy `worker/go-ingestor/config.example.env` to `.env`, populate secrets (GitHub PAT, Graph app credentials, Port tokens, webhook URLs), toggle `INGEST_GITHUB` / `INGEST_M365` as needed, and run `go build -o ingest ./worker/go-ingestor/...`.
3. Choose a deployment target from `deploy/` and plug in the same environment variables for scheduled runs.
4. Finish with the widgets in `docs/dashboard.md` and the validation checklist in `docs/validation-and-guardrails.md`.

Everything under `docs/` can be read independently, but the sequence page keeps you in the right order.
