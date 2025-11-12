package ingest

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/port-experimental/copilot-usage-ingestor/internal/httpx"
)

func periodToken(days int) string {
	switch {
	case days <= 7:
		return "D7"
	case days <= 30:
		return "D30"
	case days <= 90:
		return "D90"
	case days <= 180:
		return "D180"
	default:
		return "ALL"
	}
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func intFrom(m map[string]any, key string) int {
	if f, ok := m[key].(float64); ok {
		return int(f)
	}
	if i, ok := m[key].(int); ok {
		return i
	}
	return 0
}

func containsFold(list []string, val string) bool {
	for _, x := range list {
		if strings.EqualFold(strings.TrimSpace(x), strings.TrimSpace(val)) {
			return true
		}
	}
	return false
}

func postWebhook(ctx context.Context, hc httpx.Doer, urlStr, secret string, payload any) error {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("X-Signature", signBodySHA256(secret, b))
	}
	resp, err := httpx.DoWithRetry(ctx, hc, req, 3)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		all, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook POST failed: %s %s", resp.Status, string(all))
	}
	return nil
}

func signBodySHA256(secret string, body []byte) string {
	if secret == "" {
		return ""
	}
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}
