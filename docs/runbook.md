# Runbook — ship the ingestion stack

Goal: ingest **GitHub Copilot** (usage + seats) and **Microsoft 365 Copilot** usage into **Port** with a Go worker; ship a dashboard and guardrails.

## Apply the Port blueprints
Upload the JSON files in `configs/blueprints/` via Port Builder (Edit JSON) or the API:
- `github_copilot_seats.json`
- `github_copilot_usage.json` (extends Port defaults)
- `m365_copilot_usage_summary.json`
- `m365_copilot_user.json`

## Wire up data sources & mappings
- **GitHub Copilot metrics** — keep Port’s built-in integration and add `configs/mappings/github_copilot_mapping_override.yaml` to enrich editor/language/chat fields.
- **GitHub seats snapshot** — create a Webhook data source, paste `configs/mappings/webhook_github_seats.json`, then save the ingestion URL + secret as env vars.
- **M365 Copilot** — create two Webhook data sources using `configs/mappings/webhook_m365_summary.json` and `configs/mappings/webhook_m365_users.json`; store URLs + secret.

> Prefer Webhooks for simplicity. You can also upsert directly through Port’s Entities API if needed.

## Gather tokens and secrets
Follow `docs/tokens-and-permissions.md` to mint:
- **Port** access token (or client credentials).
- **GitHub** PAT (classic) with `manage_billing:copilot` plus `read:org` / `read:enterprise`.
- **Microsoft Graph** app credentials with `Reports.Read.All` and license-read permissions, plus admin consent.

## Configure the worker
- Copy `worker/go-ingestor/config.example.env` to `.env` and fill every variable.
- Toggle `INGEST_GITHUB` / `INGEST_M365` if you want to run one source at a time.
- When using webhooks, set `USE_PORT_WEBHOOK=true` and provide the seats + M365 ingestion URLs with a shared secret.

## Run locally for the first backfill
```bash
cd worker/go-ingestor
go build -o ingest ./...
./ingest
```
Expected outcome: a GitHub seats snapshot, one M365 summary entity, and many M365 user entities per execution.

## Deploy on a schedule
Pick whichever scheduler fits best:
- Kubernetes CronJob → `deploy/k8s-cronjob.yaml` (02:00 UTC daily).
- GitHub Actions → `deploy/github-actions.yaml` (03:30 UTC daily).

## Build the dashboard
`docs/dashboard.md` walks through KPI, trend, breakdown, and table widgets so Port surfaces the new data immediately.

## Validate and lock down
`docs/validation-and-guardrails.md` covers parity checks, rate limits, privacy, and secret rotation. Run through it after deployment and whenever the data model changes.
