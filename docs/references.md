# References (keep handy)

**GitHub Copilot**
- Org metrics: `GET /orgs/{org}/copilot/metrics`
- Team metrics: `GET /orgs/{org}/team/{team_slug}/copilot/metrics`
- Seats: `GET /orgs/{org}/copilot/billing/seats`
- API version header: `X-GitHub-Api-Version: 2022-11-28`

**Microsoft Graph**
- Summary: `GET /beta/reports/getMicrosoft365CopilotUserCountSummary(period='D7|D30|D90|D180|ALL')?$format=application/json`
- User detail: `GET /beta/reports/getMicrosoft365CopilotUsageUserDetail(period='D30')?$format=application/json`
- Licenses: `GET /v1.0/subscribedSkus`
- OAuth2: `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token` (client credentials)

**Port**
- API base: EU `https://api.getport.io` · US `https://api.us.getport.io`
- Auth: `POST /v1/auth/access_token` (client ID/secret → access token)
- Entities: `POST /v1/blueprints/{blueprint}/entities?upsert=true`
- Webhooks: create “Webhook” data source and paste mappings from `configs/mappings/`

Tip: when in doubt, use **webhook ingestion** for quick wins; you can always switch to direct API later.
