package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration populated from environment variables.
type Config struct {
	// Port
	PortRegion       string
	PortClientID     string
	PortClientSecret string
	PortAccessToken  string
	UseWebhook       bool
	WebhookSecret    string

	WebhookSeatsURL   string
	WebhookM365SumURL string
	WebhookM365UsrURL string

	// GitHub
	GitHubOrg     string
	GitHubToken   string
	GitHubAPIBase string
	GitHubAPIVer  string

	// Microsoft Graph
	MSTenantID     string
	MSClientID     string
	MSClientSecret string
	GraphAPIBase   string
	M365Skus       []string

	// Behavior
	PeriodDays     int
	SeatsActiveD14 int

	// Feature toggles
	EnableGitHub bool
	EnableM365   bool
}

// Load parses environment variables into Config with defaults.
func Load() Config {
	period := 30
	if s := strings.TrimSpace(os.Getenv("PERIOD_DAYS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			period = n
		}
	}
	active14 := 14
	if s := strings.TrimSpace(os.Getenv("SEATS_ACTIVE_WINDOW_DAYS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			active14 = n
		}
	}
	var skus []string
	if v := strings.TrimSpace(os.Getenv("M365_COPILOT_SKUS")); v != "" {
		for _, p := range strings.Split(v, ",") {
			if t := strings.TrimSpace(p); t != "" {
				skus = append(skus, t)
			}
		}
	}
	enableGitHub := boolEnv("INGEST_GITHUB", true)
	enableM365 := boolEnv("INGEST_M365", true)
	if !enableGitHub && !enableM365 {
		log.Fatal("set INGEST_GITHUB=true and/or INGEST_M365=true to ingest at least one source")
	}
	return Config{
		PortRegion:        getOr("PORT_REGION", "eu"),
		PortClientID:      os.Getenv("PORT_CLIENT_ID"),
		PortClientSecret:  os.Getenv("PORT_CLIENT_SECRET"),
		PortAccessToken:   os.Getenv("PORT_ACCESS_TOKEN"),
		UseWebhook:        strings.EqualFold(os.Getenv("USE_PORT_WEBHOOK"), "true"),
		WebhookSecret:     os.Getenv("PORT_WEBHOOK_SECRET"),
		WebhookSeatsURL:   os.Getenv("PORT_WEBHOOK_SEATS_URL"),
		WebhookM365SumURL: os.Getenv("PORT_WEBHOOK_M365_SUMMARY_URL"),
		WebhookM365UsrURL: os.Getenv("PORT_WEBHOOK_M365_USERS_URL"),
		GitHubOrg:         mustEnv("GITHUB_ORG", !enableGitHub),
		GitHubToken:       mustEnv("GITHUB_TOKEN", !enableGitHub),
		GitHubAPIBase:     getOr("GITHUB_API_BASE", "https://api.github.com"),
		GitHubAPIVer:      getOr("GITHUB_API_VERSION", "2022-11-28"),
		MSTenantID:        mustEnv("MS_TENANT_ID", !enableM365),
		MSClientID:        mustEnv("MS_CLIENT_ID", !enableM365),
		MSClientSecret:    mustEnv("MS_CLIENT_SECRET", !enableM365),
		GraphAPIBase:      getOr("GRAPH_API_BASE", "https://graph.microsoft.com"),
		M365Skus:          skus,
		PeriodDays:        period,
		SeatsActiveD14:    active14,
		EnableGitHub:      enableGitHub,
		EnableM365:        enableM365,
	}
}

func mustEnv(key string, optional bool) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" && !optional {
		log.Fatalf("missing required env: %s", key)
	}
	return v
}

func getOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func boolEnv(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Fatalf("invalid boolean for %s: %v", key, err)
	}
	return b
}
