# Dashboard cheatsheet

Spin up a dashboard page in Port, filter to the last 30 days, and drop in the widgets below. Built once, it works for every Copilot worker run.

## KPI row
| KPI | Blueprint | Metric |
| --- | --- | --- |
| GitHub Acceptance Rate | `github_copilot_usage` | `acceptance_rate` (average) |
| M365 License Utilization | `m365_copilot_usage_summary` | `license_utilization_rate` (latest) |
| Seats Active (14d) | `github_copilot_seats` | `seats_active_14d` (latest) |
| Chat Adoption % | `github_copilot_usage` | `chat_adoption_rate` (average) |

## Scorecards (leaders glance here)
- **GitHub Active Seats %:** `seats_active_14d / seats_total`, thresholds 70/40.
- **GitHub Suggestions / Active Dev:** `total_suggestions / total_active_users`.
- **M365 License Utilization %:** `active_user_count / sku_total`, thresholds 60/35.
- **M365 Weekly Depth:** % of `m365_copilot_user` entities with `days_since_last_activity <= 7`.

## Trends
- GitHub active users → line chart (`record_date`, `total_active_users`).
- GitHub chat turns → line chart (`record_date`, `total_chat_turns`).
- M365 active users → line chart (`report_date`, `active_user_count`).

## Breakdowns & tables
- Editors vs languages (pie charts using `editor_top`, `language_top`).
- App mix for M365 (stacked bar by recent activity columns).
- Dormant M365 users (table sorted by `days_since_last_activity` ≥ 30).
- Top teams by acceptance (table sorted by `acceptance_rate` with min suggestions filter).
