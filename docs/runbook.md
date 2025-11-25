# Runbook — Copilot Worker (10‑minute version)

**Outcome:** ingest GitHub Copilot seats/usage plus M365 Copilot summary + user detail into Port, then surface them on a dashboard.

## 1. Port assets
1. Upload `configs/blueprints/*.json` (seats, usage, m365 summary, m365 user) via Builder → **Edit JSON** or the Port API.
2. Drop the webhook mapping files from `configs/mappings/` into Port:
   - `webhook_github_seats.json`
   - `webhook_m365_summary.json`
   - `webhook_m365_users.json`
   - Optional: apply `github_copilot_mapping_override.yaml` if you already pull GitHub usage via the built-in integration.

> Builder flow: Data Sources → New → Webhook → paste JSON → set a random `security.secret`. Capture each resulting URL + the secret for `.env`.

## 2. Secrets checklist
- Port client credentials _or_ personal API token.
- GitHub PAT with `manage_billing:copilot` + `read:org`.
- Entra ID app (`Reports.Read.All`, admin-consented) for Microsoft Graph.
- Webhook URLs and shared secret from step 1.
- Store everything in a secret manager; reference the names inside your values/CI/CD tool.

## 3. Configure the worker
```bash
cp workers/copilot-worker/copilot.config.example.env .env
# fill env vars, then run locally:
cd workers/copilot-worker
go build -o copilot-worker ./...
./copilot-worker
```
Expect: one GitHub seats snapshot, one M365 summary entity per run, and as many M365 user entities as licenses.

## 4. Schedule it
- **GitHub Actions** → `deploy/github-actions.yaml` (runs daily at 03:30 UTC).
- **Kubernetes CronJob** → `deploy/k8s-cronjob.yaml` or the Helm chart in `deploy/helm/copilot-worker`.

## 5. Observe + harden
- Build the KPI widgets listed in `docs/dashboard.md`.
- Review `docs/validation-and-guardrails.md` after first ingest (parity checks, rate limits, retention, alerting).

Add more workers by repeating this pattern under `workers/<new-worker>`.
