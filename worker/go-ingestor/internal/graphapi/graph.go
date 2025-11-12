package graphapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/port-experimental/copilot-usage-ingestor/internal/httpx"
)

// Token performs client-credentials OAuth2 against Entra ID for Graph.
func Token(ctx context.Context, hc httpx.Doer, tenant, clientID, clientSecret string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", "https://graph.microsoft.com/.default")
	ep := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenant)
	req, _ := http.NewRequestWithContext(ctx, "POST", ep, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpx.DoWithRetry(ctx, hc, req, 3)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		all, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("graph token: %s %s", resp.Status, all)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

// CopilotSummary fetches tenant-level summary for a chosen period (beta).
func CopilotSummary(ctx context.Context, hc httpx.Doer, base, token, period string) (map[string]any, error) {
	q := url.Values{}
	q.Set("$format", "application/json")
	path := fmt.Sprintf("/beta/reports/getMicrosoft365CopilotUserCountSummary(period='%s')", period)
	resp, err := graphGet(ctx, hc, base, token, path, q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		all, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph summary: %s %s", resp.Status, all)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// CopilotUserDetail returns per-user last-activity fields.
func CopilotUserDetail(ctx context.Context, hc httpx.Doer, base, token, period string) ([]map[string]any, error) {
	q := url.Values{}
	q.Set("$format", "application/json")
	path := fmt.Sprintf("/beta/reports/getMicrosoft365CopilotUsageUserDetail(period='%s')", period)
	resp, err := graphGet(ctx, hc, base, token, path, q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		all, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph user detail: %s %s", resp.Status, all)
	}
	var arr []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// SubscribedSkus lists SKUs; we count configured Copilot skuPartNumbers only.
func SubscribedSkus(ctx context.Context, hc httpx.Doer, base, token string) ([]map[string]any, error) {
	resp, err := graphGet(ctx, hc, base, token, "/v1.0/subscribedSkus", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		all, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph skus: %s %s", resp.Status, all)
	}
	var out struct {
		Value []map[string]any `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

func graphGet(ctx context.Context, hc httpx.Doer, base, token, path string, q url.Values) (*http.Response, error) {
	u := strings.TrimRight(base, "/") + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return httpx.DoWithRetry(ctx, hc, req, 3)
}
