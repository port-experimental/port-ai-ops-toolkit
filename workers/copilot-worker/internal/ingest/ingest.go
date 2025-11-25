package ingest

import (
	"context"
	"log"
	"time"

	"github.com/port-labs/port-ai-ops-toolkit/pkg/common/httpx"
	"github.com/port-labs/port-ai-ops-toolkit/pkg/common/portapi"
	"github.com/port-labs/port-ai-ops-toolkit/workers/copilot-worker/internal/config"
	"github.com/port-labs/port-ai-ops-toolkit/workers/copilot-worker/internal/githubapi"
	"github.com/port-labs/port-ai-ops-toolkit/workers/copilot-worker/internal/graphapi"
)

// GitHubSeats ingests GitHub Copilot seat snapshots via webhook or Port API.
func GitHubSeats(ctx context.Context, cfg config.Config, hc httpx.Doer, pcli *portapi.Client, recordDate string) {
	seats, err := githubapi.FetchSeats(ctx, hc, cfg.GitHubAPIBase, cfg.GitHubAPIVer, cfg.GitHubToken, cfg.GitHubOrg)
	if err != nil {
		log.Printf("warn: gh seats: %v", err)
		return
	}
	seatsTotal := len(seats)
	cut14 := time.Now().AddDate(0, 0, -cfg.SeatsActiveD14)
	cut30 := time.Now().AddDate(0, 0, -30)
	var seatsActive14, seatsActive30 int
	for _, s := range seats {
		if s.LastActivityAt != nil {
			if s.LastActivityAt.After(cut14) {
				seatsActive14++
			}
			if s.LastActivityAt.After(cut30) {
				seatsActive30++
			}
		}
	}
	if cfg.UseWebhook {
		payload := map[string]any{
			"kind": "gh-copilot-seats",
			"record": map[string]any{
				"record_date":      recordDate,
				"seats_total":      seatsTotal,
				"seats_active_14d": seatsActive14,
				"seats_active_30d": seatsActive30,
			},
		}
		if err := postWebhook(ctx, hc, cfg.WebhookSeatsURL, cfg.WebhookSecret, payload); err != nil {
			log.Printf("warn: seats webhook: %v", err)
		}
		return
	}

	ent := map[string]any{
		"identifier": recordDate,
		"properties": map[string]any{
			"record_date":      recordDate,
			"seats_total":      seatsTotal,
			"seats_active_14d": seatsActive14,
			"seats_active_30d": seatsActive30,
		},
	}
	if err := pcli.UpsertEntity(ctx, "github_copilot_seats", ent); err != nil {
		log.Printf("warn: seats upsert: %v", err)
	}
}

// M365 ingests Microsoft 365 Copilot summary + user details.
func M365(ctx context.Context, cfg config.Config, hc httpx.Doer, pcli *portapi.Client, recordDate string) {
	period := periodToken(cfg.PeriodDays)
	gTok, err := graphapi.Token(ctx, hc, cfg.MSTenantID, cfg.MSClientID, cfg.MSClientSecret)
	if err != nil {
		log.Fatalf("graph token: %v", err)
	}

	summary, err := graphapi.CopilotSummary(ctx, hc, cfg.GraphAPIBase, gTok, period)
	if err != nil {
		log.Fatalf("graph summary: %v", err)
	}
	enabled := intFrom(summary, "enabledUserCount")
	active := intFrom(summary, "activeUserCount")

	var skuTotal int
	if len(cfg.M365Skus) > 0 {
		skus, err := graphapi.SubscribedSkus(ctx, hc, cfg.GraphAPIBase, gTok)
		if err != nil {
			log.Printf("warn: graph skus: %v", err)
		} else {
			for _, v := range skus {
				part := str(v["skuPartNumber"])
				if containsFold(cfg.M365Skus, part) {
					if pu, ok := v["prepaidUnits"].(map[string]any); ok {
						if en, ok2 := pu["enabled"].(float64); ok2 {
							skuTotal += int(en)
						}
					}
				}
			}
		}
	}

	if cfg.UseWebhook {
		payload := map[string]any{
			"kind": "m365-copilot-summary",
			"record": map[string]any{
				"period":             period,
				"report_date":        recordDate,
				"enabled_user_count": enabled,
				"active_user_count":  active,
				"sku_total":          skuTotal,
			},
		}
		if err := postWebhook(ctx, hc, cfg.WebhookM365SumURL, cfg.WebhookSecret, payload); err != nil {
			log.Printf("warn: m365 summary webhook: %v", err)
		}
	} else {
		ent := map[string]any{
			"identifier": period + "@" + recordDate,
			"properties": map[string]any{
				"period":             period,
				"report_date":        recordDate,
				"enabled_user_count": enabled,
				"active_user_count":  active,
				"sku_total":          skuTotal,
			},
		}
		if err := pcli.UpsertEntity(ctx, "m365_copilot_usage_summary", ent); err != nil {
			log.Printf("warn: m365 summary upsert: %v", err)
		}
	}

	users, err := graphapi.CopilotUserDetail(ctx, hc, cfg.GraphAPIBase, gTok, period)
	if err != nil {
		log.Printf("warn: graph user detail: %v", err)
		users = nil
	}
	const maxUsersPerRun = 5000
	count := 0
	for _, u := range users {
		if count >= maxUsersPerRun {
			break
		}
		upn := str(u["userPrincipalName"])
		hash := upn
		if hash == "" {
			if dn := str(u["displayName"]); dn != "" {
				hash = dn
			} else {
				hash = stableMapFingerprint(u)
			}
		}
		userProps := map[string]any{
			"period":                        period,
			"report_date":                   recordDate,
			"user_principal_name":           upn,
			"user_hash":                     sha256Hex(hash),
			"last_activity_date":            u["lastActivityDate"],
			"teams_copilot_last_activity":   u["microsoftTeamsCopilotLastActivityDate"],
			"word_copilot_last_activity":    u["wordCopilotLastActivityDate"],
			"excel_copilot_last_activity":   u["excelCopilotLastActivityDate"],
			"ppt_copilot_last_activity":     u["powerPointCopilotLastActivityDate"],
			"outlook_copilot_last_activity": u["outlookCopilotLastActivityDate"],
			"onenote_copilot_last_activity": u["oneNoteCopilotLastActivityDate"],
			"loop_copilot_last_activity":    u["loopCopilotLastActivityDate"],
			"chat_last_activity":            u["copilotChatLastActivityDate"],
		}
		if cfg.UseWebhook {
			payload := map[string]any{"kind": "m365-copilot-users", "user": userProps}
			if err := postWebhook(ctx, hc, cfg.WebhookM365UsrURL, cfg.WebhookSecret, payload); err != nil {
				log.Printf("warn: m365 users webhook: %v", err)
			}
		} else {
			ent := map[string]any{
				"identifier": userProps["user_hash"],
				"properties": userProps,
			}
			if err := pcli.UpsertEntity(ctx, "m365_copilot_user", ent); err != nil {
				log.Printf("warn: m365 user upsert: %v", err)
			}
		}
		count++
	}
	if len(users) > maxUsersPerRun {
		log.Printf("warn: m365 user detail truncated: processed %d of %d rows", maxUsersPerRun, len(users))
	}
}
