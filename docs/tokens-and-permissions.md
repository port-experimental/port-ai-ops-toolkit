# Tokens & Permissions (exact requirements)

## Port
- Create **Client ID / Secret** in Port → Settings → Credentials.
- Either:
  - Use **client credentials** to mint short-lived tokens at runtime (`/v1/auth/access_token`), _or_
  - Generate a **personal API token** and set `PORT_ACCESS_TOKEN`.

## GitHub (Copilot)
- Use a **classic PAT** (recommended for today) with scopes:
  - `manage_billing:copilot` (required for Copilot seat endpoints)
  - `read:org` (org metrics) or `read:enterprise` (if fetching at enterprise scope)
- Ensure **Copilot metrics access policy** is enabled at the org/enterprise level.

## Microsoft Graph (M365 Copilot)
- Create an **Entra ID app registration**.
- Grant **Application** permissions:
  - `Reports.Read.All` (Copilot usage reports)
  - For license counts via `/subscribedSkus`, grant read permissions for directory/organization (e.g., `Directory.Read.All`) if required by your tenant policies.
- **Admin consent** the app.
- Token flow: **client credentials** to `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token` with scope `https://graph.microsoft.com/.default`.

## Secrets Handling
- Store all secrets as environment variables (see `workers/copilot-worker/copilot.config.example.env`).
- Don’t log tokens. Rotate quarterly or per policy; revoke on role changes.
