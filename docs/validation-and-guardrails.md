# Validation & guardrails

## Functional checks
- **Entities present:** after first run, you should see:
  - one `github_copilot_seats` entity (latest `record_date`),
  - one `m365_copilot_usage_summary` entity per run,
  - many `m365_copilot_user` entities (or none if de-identified and blocked by policy).
- **Idempotency:** rerun the worker; entities should **upsert** (no duplicates).

## Data parity
- Compare Port metrics vs source admin portals:
  - GitHub Copilot org metrics page (yesterday’s data) for `total_active_users`.
  - M365 admin reports for Copilot (summary vs user detail).

## Rate limits & retries
- The worker backs off on `429` and `5xx`. Keep schedules daily (02:00 UTC).

## Privacy
- If M365 de-identifies users, `user_principal_name` will be blank; we store a `user_hash` instead.

## Security
- Secrets via env only; **never** log tokens.
- Rotate PAT and app secrets per policy; revoke promptly.
- For Webhooks: use HMAC signature; rotate `PORT_WEBHOOK_SECRET` quarterly.

## Operations
- Alerts: page on job failures or if seat utilization remains < 40% for 14 days.
- Housekeeping: keep only 180 days of `m365_copilot_user` entities if storage limits bite; summaries are compact.
