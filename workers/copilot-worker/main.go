// Ingest GitHub Copilot seats and Microsoft 365 Copilot usage into Port.
// - Uses Port Webhooks (recommended) or Port Entities API directly.
// - Safe by default: timeouts, retries with backoff, and no secret logging.
// - PII: hashes UPN if de-identified reports hide user names.
//
// Build: go build -o copilot-worker ./...
// Env: see copilot.config.example.env
package main

import (
	"context"
	"log"
	"time"

	"github.com/port-labs/port-ai-ops-toolkit/pkg/common/httpx"
	"github.com/port-labs/port-ai-ops-toolkit/pkg/common/portapi"
	"github.com/port-labs/port-ai-ops-toolkit/workers/copilot-worker/internal/config"
	"github.com/port-labs/port-ai-ops-toolkit/workers/copilot-worker/internal/ingest"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[copilot-worker] ")
	cfg := config.Load()

	// Fast sanity: require either Webhook URLs (UseWebhook=true) OR Port credentials.
	if cfg.UseWebhook {
		if cfg.EnableGitHub && cfg.WebhookSeatsURL == "" {
			log.Fatal("USE_PORT_WEBHOOK=true but PORT_WEBHOOK_SEATS_URL is missing while INGEST_GITHUB=true")
		}
		if cfg.EnableM365 && (cfg.WebhookM365SumURL == "" || cfg.WebhookM365UsrURL == "") {
			log.Fatal("USE_PORT_WEBHOOK=true but M365 webhook URLs are missing while INGEST_M365=true")
		}
	} else {
		if cfg.PortAccessToken == "" && (cfg.PortClientID == "" || cfg.PortClientSecret == "") {
			log.Fatal("Provide PORT_ACCESS_TOKEN or PORT_CLIENT_ID/PORT_CLIENT_SECRET")
		}
	}

	hc := httpx.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create Port client only if needed
	var (
		pcli *portapi.Client
		err  error
	)
	if !cfg.UseWebhook {
		pcli, err = portapi.NewClient(ctx, hc, cfg.PortRegion, cfg.PortAccessToken, cfg.PortClientID, cfg.PortClientSecret)
		if err != nil {
			log.Fatalf("port client: %v", err)
		}
	}

	recordDate := time.Now().UTC().Format(time.RFC3339)

	if cfg.EnableGitHub {
		ingest.GitHubSeats(ctx, cfg, hc, pcli, recordDate)
	}

	if cfg.EnableM365 {
		ingest.M365(ctx, cfg, hc, pcli, recordDate)
	}

	log.Println("ingestion completed")
}
