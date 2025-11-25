package graphapi

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/port-labs/port-ai-ops-toolkit/pkg/common/httpx"
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
	httpx.SetUserAgent(req)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("graph summary: %s %s", resp.Status, body)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err == nil && len(out) > 0 {
		return out, nil
	}
	summary, err := parseSummaryCSV(body)
	if err != nil {
		return nil, err
	}
	return summary, nil
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("graph user detail: %s %s", resp.Status, body)
	}
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err == nil && arr != nil {
		return arr, nil
	}
	return parseUserDetailCSV(body)
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
	httpx.SetUserAgent(req)
	return httpx.DoWithRetry(ctx, hc, req, 3)
}

func parseSummaryCSV(body []byte) (map[string]any, error) {
	rows, err := parseCSVRows(body)
	if err != nil {
		return nil, fmt.Errorf("decode summary csv: %w", err)
	}
	if len(rows) == 0 {
		return nil, errors.New("graph summary: empty report")
	}
	row := normalizeRow(rows[0])
	summary := map[string]any{}
	summary["enabledUserCount"] = atoiAny(row, []string{
		"microsoft365_copilot_enabled_user_count",
		"enabled_user_count",
	})
	summary["activeUserCount"] = atoiAny(row, []string{
		"microsoft365_copilot_active_user_count",
		"active_user_count",
	})
	if rd := firstNonEmpty(row, []string{"report_refresh_date", "report_date"}); rd != "" {
		summary["reportDate"] = rd
	}
	return summary, nil
}

func parseUserDetailCSV(body []byte) ([]map[string]any, error) {
	rows, err := parseCSVRows(body)
	if err != nil {
		return nil, fmt.Errorf("decode user detail csv: %w", err)
	}
	var result []map[string]any
	for _, raw := range rows {
		n := normalizeRow(raw)
		user := map[string]any{}
		for key, target := range userFieldTargets() {
			if val := n[key]; val != "" {
				user[target] = val
			}
		}
		if len(user) == 0 {
			continue
		}
		result = append(result, user)
	}
	return result, nil
}

func parseCSVRows(body []byte) ([]map[string]string, error) {
	body = bytes.TrimPrefix(body, []byte("\ufeff"))
	r := csv.NewReader(bytes.NewReader(body))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) <= 1 {
		return nil, nil
	}
	headers := records[0]
	var rows []map[string]string
	for _, rec := range records[1:] {
		if len(rec) == 1 && strings.TrimSpace(rec[0]) == "" {
			continue
		}
		row := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(rec) {
				row[h] = rec[i]
			} else {
				row[h] = ""
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func normalizeRow(row map[string]string) map[string]string {
	out := make(map[string]string, len(row))
	for k, v := range row {
		out[normalizeKey(k)] = strings.TrimSpace(v)
	}
	return out
}

func normalizeKey(s string) string {
	var b strings.Builder
	lastUnderscore := false
	prevAlphaNum := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if unicode.IsUpper(r) && prevAlphaNum && !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
			b.WriteRune(unicode.ToLower(r))
			prevAlphaNum = true
			lastUnderscore = false
		default:
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
			prevAlphaNum = false
		}
	}
	res := b.String()
	res = strings.Trim(res, "_")
	res = strings.ReplaceAll(res, "__", "_")
	return res
}

func atoiAny(row map[string]string, keys []string) int {
	for _, k := range keys {
		if v := strings.TrimSpace(row[k]); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return 0
}

func firstNonEmpty(row map[string]string, keys []string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(row[k]); v != "" {
			return v
		}
	}
	return ""
}

func userFieldTargets() map[string]string {
	return map[string]string{
		"user_principal_name": "userPrincipalName",
		"userprincipalname":   "userPrincipalName",
		"display_name":        "displayName",
		"displayname":         "displayName",
		"last_activity_date":  "lastActivityDate",
		"lastactivitydate":    "lastActivityDate",
		"microsoft_teams_copilot_last_activity_date": "microsoftTeamsCopilotLastActivityDate",
		"microsoftteamscopilotlastactivitydate":      "microsoftTeamsCopilotLastActivityDate",
		"word_copilot_last_activity_date":            "wordCopilotLastActivityDate",
		"wordcopilotlastactivitydate":                "wordCopilotLastActivityDate",
		"excel_copilot_last_activity_date":           "excelCopilotLastActivityDate",
		"excelcopilotlastactivitydate":               "excelCopilotLastActivityDate",
		"power_point_copilot_last_activity_date":     "powerPointCopilotLastActivityDate",
		"powerpointcopilotlastactivitydate":          "powerPointCopilotLastActivityDate",
		"outlook_copilot_last_activity_date":         "outlookCopilotLastActivityDate",
		"outlookcopilotlastactivitydate":             "outlookCopilotLastActivityDate",
		"one_note_copilot_last_activity_date":        "oneNoteCopilotLastActivityDate",
		"onenotecopilotlastactivitydate":             "oneNoteCopilotLastActivityDate",
		"loop_copilot_last_activity_date":            "loopCopilotLastActivityDate",
		"loopcopilotlastactivitydate":                "loopCopilotLastActivityDate",
		"copilot_chat_last_activity_date":            "copilotChatLastActivityDate",
		"copilotchatlastactivitydate":                "copilotChatLastActivityDate",
	}
}
