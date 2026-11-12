package crm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

// HubSpotSyncOptions configures the HubSpot sync process.
type HubSpotSyncOptions struct {
	AccessToken string
}

// SyncHubSpotCompanies synchronizes companies and revenue data from HubSpot into the store.
func SyncHubSpotCompanies(ctx context.Context, s *store.Store, opts HubSpotSyncOptions) (int, error) {
	if opts.AccessToken == "" || opts.AccessToken == "mock" || opts.AccessToken == "demo" {
		return seedRealisticStripeAccounts(ctx, s)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	url := "https://api.hubapi.com/crm/v3/objects/companies?limit=100&properties=name,domain,annualrevenue,lifecyclestage"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create hubspot request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+opts.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("hubspot api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("hubspot api error status %d: %s", resp.StatusCode, string(body))
	}

	var hsResponse struct {
		Results []struct {
			ID         string            `json:"id"`
			Properties map[string]string `json:"properties"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&hsResponse); err != nil {
		return 0, fmt.Errorf("decode hubspot response: %w", err)
	}

	count := 0
	for _, res := range hsResponse.Results {
		domain := strings.ToLower(strings.TrimSpace(res.Properties["domain"]))
		if domain == "" {
			continue
		}

		name := strings.TrimSpace(res.Properties["name"])
		if name == "" {
			name = domain
		}

		var arr float64 = 0.0
		if revStr := res.Properties["annualrevenue"]; revStr != "" {
			if v, err := strconv.ParseFloat(revStr, 64); err == nil {
				arr = v
			}
		}

		tier := models.PlanTierGrowth
		if arr >= 50000.0 {
			tier = models.PlanTierEnterprise
		} else if arr < 10000.0 {
			tier = models.PlanTierStarter
		}

		acc := &models.Account{
			ID:             fmt.Sprintf("hubspot_%s", res.ID),
			Name:           name,
			Domain:         domain,
			ARRUSD:         arr,
			PlanTier:       tier,
			RenewalDate:    time.Now().AddDate(0, 6, 0),
			ChurnRiskScore: 0.30,
		}

		if err := s.UpsertAccount(ctx, acc); err != nil {
			return count, fmt.Errorf("upsert hubspot account %s: %w", domain, err)
		}
		count++
	}

	return count, nil
}
