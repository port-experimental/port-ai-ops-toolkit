# Worker modules (build more ingestion bots fast)

The repository is organized as `workers/<name>` so you can ship additional binaries, containers, and charts without cloning the codebase again. Each worker keeps its own `go.mod`, Dockerfile, Helm values, and docs snippet.

## Anatomy of a worker
| Piece | Location | Notes |
| --- | --- | --- |
| Source | `workers/<name>/*.go` | Stick to the same `internal/` package layout so shared helpers stay familiar. |
| Config template | `workers/<name>/<name>.config.example.env` | Document every environment variable once. |
| Container | `workers/<name>/Dockerfile` | Build a static binary and copy it into a distroless base. |
| Automation | `deploy/helm/<name>-chart`, `deploy/<name>-*.yaml` | Keep chart values mirrored with the config template. |

## Add a new worker (TL;DR)
1. `cp -R workers/copilot-worker workers/<new-worker>` and rename files/binaries inside.
2. Update module path: `module github.com/port-labs/port-ai-ops-toolkit/workers/<new-worker>`.
3. Rewrite `internal/` packages to hit the new AI tool’s APIs.
4. Add a new Helm chart and deployment manifests under `deploy/`.
5. Document the worker in `README.md` + `docs/` so operators know when to use it.

Keep binaries laser-focused: one worker per AI tool (Copilot, Cursor, Windsurf, etc.). This repo then becomes the shared playbook for running them side by side.
