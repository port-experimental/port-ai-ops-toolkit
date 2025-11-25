package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/port-labs/port-ai-ops-toolkit/pkg/common/httpx"
)

// Seat is the subset of the Copilot seats response we care about.
type Seat struct {
	AssignedAt     *time.Time `json:"assigned_at"`
	LastActivityAt *time.Time `json:"last_activity_at"`
}

// FetchSeats pages through Copilot seat assignments for an org.
func FetchSeats(ctx context.Context, hc httpx.Doer, base, apiVer, token, org string) ([]Seat, error) {
	var seats []Seat
	page := 1
	for {
		ep := fmt.Sprintf("%s/orgs/%s/copilot/billing/seats?per_page=100&page=%d",
			strings.TrimRight(base, "/"), url.PathEscape(org), page)
		req, _ := http.NewRequestWithContext(ctx, "GET", ep, nil)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", apiVer)
		req.Header.Set("Authorization", "Bearer "+token)
		httpx.SetUserAgent(req)
		resp, err := httpx.DoWithRetry(ctx, hc, req, 3)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 404 {
			_ = resp.Body.Close()
			return seats, nil
		}
		if resp.StatusCode >= 300 {
			all, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return nil, fmt.Errorf("gh seats: %s %s", resp.Status, all)
		}
		var out struct {
			Seats []Seat `json:"seats"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			_ = resp.Body.Close()
			return nil, err
		}
		_ = resp.Body.Close()
		seats = append(seats, out.Seats...)
		if len(out.Seats) < 100 {
			break
		}
		page++
	}
	return seats, nil
}
