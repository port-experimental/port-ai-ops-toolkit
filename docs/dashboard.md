# Dashboard (AI consumption)

Create a **Dashboard** page in Port and add these widgets. Filter by last 30 days where applicable.

## KPI Row
- **GitHub Acceptance Rate (avg 30d)** — blueprint: `github_copilot_usage`, property: `acceptance_rate` (avg).
- **M365 License Utilization %** — blueprint: `m365_copilot_usage_summary`, property: `license_utilization_rate` (latest).
- **Seats Active (14d)** — blueprint: `github_copilot_seats`, property: `seats_active_14d` (latest).
- **Chat Adoption %** — blueprint: `github_copilot_usage`, property: `chat_adoption_rate` (avg).

## Trends (Line)
- **Active Users (GitHub)** — x:`record_date`, y:`total_active_users`.
- **Chat Turns (GitHub)** — x:`record_date`, y:`total_chat_turns`.
- **Active Users (M365)** — x:`report_date`, y:`active_user_count` per period entity.

## Breakdowns
- **Top Editors (GitHub)** — pie by `editor_top`.
- **Top Languages (GitHub)** — pie by `language_top`.
- **App Mix (M365)** — create stacked bars from user detail using `*_last_activity` recency (optional).

## Tables
- **Dormant M365 Users (≥30d)** — from `m365_copilot_user`, compute `DATEDIFF(now, last_activity_date)` and sort.
- **Top Teams by Acceptance Rate** — if you ingest team-level usage, sort `acceptance_rate` with a minimum suggestions threshold.
