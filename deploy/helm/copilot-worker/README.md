# Copilot Worker Helm Chart

This chart packages the CronJob that runs the Go worker from `workers/copilot-worker`. It exposes every configuration flag the binary reads, grouping sensitive values under a Kubernetes Secret while keeping the rest inlined as environment variables.

## Installing

```bash
helm upgrade --install copilot-worker ./deploy/helm/copilot-worker \
  --namespace ai-ingestion --create-namespace \
  -f my-values.yaml
```

Override `my-values.yaml` with the correct schedule, image tag, and runtime configuration pulled from `workers/copilot-worker/copilot.config.example.env`.

## Configuration

Key sections in `values.yaml`:

- `env`: non-sensitive environment variables such as toggles, regions, API hosts, and M365 tenant identifiers.
- `secret.data`: sensitive values (client IDs, secrets, PATs, webhook URLs). When `secret.create=true`, the chart renders a Secret named `<release>-copilot-worker-secret`. Set `secret.nameOverride` and `secret.create=false` to reuse an externally managed Secret.
- `cronJob`: schedule, history limits, restart policy, deadlines, and annotations for the CronJob/job template.
- `resources`, `nodeSelector`, `affinity`, `tolerations`, `imagePullSecrets`: standard pod controls.

All environment variables supported by the worker are already declared in `values.yaml`; fill them in and remove the ones you do not need. Leave secrets blank in your Git-managed values and populate them during deployment (for example via `helm install ... --set-file secret.data.GITHUB_TOKEN=token.txt`).
