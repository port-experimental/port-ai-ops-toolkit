# Differences vs Port’s existing GitHub Copilot integration

| Area | Port default behavior | This package |
|---|---|---|
| **Data source** | Built-in GitHub Copilot integration ingests **metrics** (org/team). | Keep it for metrics; **add seats** via a Webhook (or direct API) to track license utilization. |
| **Mapping** | Default mapping calculates totals and `acceptance_rate`. | Adds `editor_top`, `language_top`, chat fields (`total_chat_turns`, `total_active_chat_users`, `total_chat_acceptances`). |
| **Seats/licensing** | Not included in metrics. | New blueprint `github_copilot_seats` + daily snapshot via Go worker. |
| **Seat utilization** | Not available. | Optional calc in dashboards (active users vs seats snapshot); optional property `seat_utilization_rate` if you enrich usage entities. |
| **M365 Copilot** | No built-in integration. | **New**: `m365_copilot_usage_summary` + `m365_copilot_user` via Graph. |
| **Privacy** | Not applicable. | UPNs hashed if reports are de-identified; only store hashes when needed. |
