package portapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/port-labs/port-ai-ops-toolkit/pkg/common/httpx"
)

// Client handles direct Port entity upserts when webhooks aren't used.
type Client struct {
	base   string
	token  string
	client httpx.Doer
}

func NewClient(ctx context.Context, hc httpx.Doer, region, accessToken, clientID, clientSecret string) (*Client, error) {
	base := map[string]string{"eu": "https://api.getport.io", "us": "https://api.us.getport.io"}[strings.ToLower(region)]
	if base == "" {
		base = "https://api.getport.io"
	}
	tok := strings.TrimSpace(accessToken)
	if tok == "" {
		ep := base + "/v1/auth/access_token"
		body := map[string]string{"clientId": clientID, "clientSecret": clientSecret}
		if clientID == "" || clientSecret == "" {
			return nil, errors.New("PORT_CLIENT_ID/PORT_CLIENT_SECRET or PORT_ACCESS_TOKEN required")
		}
		b, _ := json.Marshal(body)
		req, _ := http.NewRequestWithContext(ctx, "POST", ep, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		httpx.SetUserAgent(req)
		resp, err := httpx.DoWithRetry(ctx, hc, req, 3)
		if err != nil {
			return nil, fmt.Errorf("port auth: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			all, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("port auth failed: %s %s", resp.Status, string(all))
		}
		var out struct {
			AccessToken string `json:"accessToken"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
		tok = out.AccessToken
	}
	return &Client{base: base, token: tok, client: hc}, nil
}

func (p *Client) UpsertEntity(ctx context.Context, blueprint string, entity any) error {
	ep := fmt.Sprintf("%s/v1/blueprints/%s/entities?upsert=true&merge=true", p.base, url.PathEscape(blueprint))
	b, _ := json.Marshal(entity)
	req, _ := http.NewRequestWithContext(ctx, "POST", ep, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")
	httpx.SetUserAgent(req)
	resp, err := httpx.DoWithRetry(ctx, p.client, req, 3)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		all, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("port upsert failed: %s %s", resp.Status, all)
	}
	return nil
}
