# Port AI Ops Toolkit

This repo houses a growing catalog of Go-based workers that normalize AI usage signals (Copilot today, Cursor/Windsurf/etc. tomorrow) and ship them into Port with batteries-included dashboards and automations.

## Layout
- `configs/blueprints/` — Port blueprints to create once.
- `configs/mappings/` — webhook payload + mapping definitions.
- `workers/` — one folder per worker module (currently `copilot-worker`).
- `deploy/` — GitHub Action, CronJob YAML, and Helm charts for every worker.
- `docs/` — bite-sized guides (`runbook`, `tokens`, `dashboard`, `workers`, etc.).

## Read me next
1. `docs/runbook.md` — ten-minute flow from blueprints to dashboards.
2. `docs/tokens-and-permissions.md` — which secrets to mint.
3. `docs/workers.md` — how to add more worker modules.
4. `docs/dashboard.md` + `docs/validation-and-guardrails.md` — observe + harden.

## Quick start (Copilot worker)
1. Apply the JSON in `configs/blueprints/`, then load the webhook mapping files from `configs/mappings/`.
2. Copy `workers/copilot-worker/copilot.config.example.env` to `.env`, fill secrets (GitHub PAT, Graph app credentials, Port tokens, webhook URLs), and run `go build -o copilot-worker ./workers/copilot-worker/...`.
3. Deploy via GitHub Actions (`deploy/github-actions.yaml`) or the Kubernetes CronJob/Helm chart under `deploy/`.
4. Build the dashboard + guardrails described in the docs.

More workers = repeat the same pattern under `workers/<name>` and point a new chart/deployment at the new binary.
