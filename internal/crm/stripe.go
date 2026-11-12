package crm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

// StripeSyncOptions configures the Stripe sync process.
type StripeSyncOptions struct {
	APIKey string
}

// SyncStripeCustomers synchronizes customers and active subscriptions from Stripe into the store.
func SyncStripeCustomers(ctx context.Context, s *store.Store, opts StripeSyncOptions) (int, error) {
	if opts.APIKey == "" || opts.APIKey == "mock" || opts.APIKey == "demo" {
		return seedRealisticStripeAccounts(ctx, s)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.stripe.com/v1/customers?limit=100&expand[]=data.subscriptions", nil)
	if err != nil {
		return 0, fmt.Errorf("create stripe request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+opts.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("stripe api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("stripe api error status %d: %s", resp.StatusCode, string(body))
	}

	var stripeResponse struct {
		Data []struct {
			ID            string            `json:"id"`
			Name          string            `json:"name"`
			Email         string            `json:"email"`
			Metadata      map[string]string `json:"metadata"`
			Subscriptions struct {
				Data []struct {
					Status             string `json:"status"`
					CurrentPeriodEnd   int64  `json:"current_period_end"`
					Items              struct {
						Data []struct {
							Price struct {
								UnitAmount int64  `json:"unit_amount"` // In cents
								Recurring  struct {
									Interval string `json:"interval"` // month, year, week
								} `json:"recurring"`
							} `json:"price"`
							Quantity int64 `json:"quantity"`
						} `json:"data"`
					} `json:"items"`
				} `json:"data"`
			} `json:"subscriptions"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&stripeResponse); err != nil {
		return 0, fmt.Errorf("decode stripe response: %w", err)
	}

	count := 0
	for _, cust := range stripeResponse.Data {
		domain := ""
		if cust.Metadata != nil && cust.Metadata["domain"] != "" {
			domain = strings.ToLower(strings.TrimSpace(cust.Metadata["domain"]))
		}
		if domain == "" && cust.Email != "" {
			domain = ExtractDomain(cust.Email)
		}
		if domain == "" {
			continue
		}

		name := cust.Name
		if name == "" {
			name = domain
		}

		var totalARRUSD float64 = 0.0
		var latestRenewal time.Time

		for _, sub := range cust.Subscriptions.Data {
			if sub.Status != "active" && sub.Status != "trialing" {
				continue
			}

			if sub.CurrentPeriodEnd > 0 {
				renewal := time.Unix(sub.CurrentPeriodEnd, 0)
				if latestRenewal.IsZero() || renewal.After(latestRenewal) {
					latestRenewal = renewal
				}
			}

			for _, item := range sub.Items.Data {
				qty := item.Quantity
				if qty <= 0 {
					qty = 1
				}
				amountUSD := float64(item.Price.UnitAmount*qty) / 100.0

				switch item.Price.Recurring.Interval {
				case "month":
					totalARRUSD += amountUSD * 12.0
				case "year":
					totalARRUSD += amountUSD
				case "week":
					totalARRUSD += amountUSD * 52.0
				default:
					totalARRUSD += amountUSD * 12.0
				}
			}
		}

		tier := models.PlanTierGrowth
		if totalARRUSD >= 50000.0 {
			tier = models.PlanTierEnterprise
		} else if totalARRUSD < 10000.0 {
			tier = models.PlanTierStarter
		}

		if latestRenewal.IsZero() {
			latestRenewal = time.Now().AddDate(0, 6, 0)
		}

		acc := &models.Account{
			ID:             fmt.Sprintf("stripe_%s", cust.ID),
			Name:           name,
			Domain:         domain,
			ARRUSD:         totalARRUSD,
			PlanTier:       tier,
			RenewalDate:    latestRenewal,
			ChurnRiskScore: 0.25,
		}

		if err := s.UpsertAccount(ctx, acc); err != nil {
			return count, fmt.Errorf("upsert stripe account %s: %w", domain, err)
		}
		count++
	}

	return count, nil
}

// seedRealisticStripeAccounts seeds benchmark customer accounts for local development and demos.
func seedRealisticStripeAccounts(ctx context.Context, s *store.Store) (int, error) {
	now := time.Now()
	sampleAccounts := []models.Account{
		{
			ID:             "acc_cyberdyne",
			Name:           "Cyberdyne Systems",
			Domain:         "cyberdyne.com",
			ARRUSD:         285000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 1, 15), // 45 days
			ChurnRiskScore: 0.75,
		},
		{
			ID:             "acc_acme",
			Name:           "Acme Global Tech",
			Domain:         "acme.com",
			ARRUSD:         195000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 2, 5), // 65 days
			ChurnRiskScore: 0.60,
		},
		{
			ID:             "acc_globex",
			Name:           "Globex Corporation",
			Domain:         "globex.corp",
			ARRUSD:         160000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 0, 20), // 20 days (critical renewal!)
			ChurnRiskScore: 0.85,
		},
		{
			ID:             "acc_soylent",
			Name:           "Soylent Financial Corp",
			Domain:         "soylent.io",
			ARRUSD:         145000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 3, 10),
			ChurnRiskScore: 0.35,
		},
		{
			ID:             "acc_massive_dyn",
			Name:           "Massive Dynamic Labs",
			Domain:         "massivedynamic.org",
			ARRUSD:         110000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 1, 5), // 35 days
			ChurnRiskScore: 0.70,
		},
		{
			ID:             "acc_initech",
			Name:           "Initech Software",
			Domain:         "initech.com",
			ARRUSD:         48000.0,
			PlanTier:       models.PlanTierGrowth,
			RenewalDate:    now.AddDate(0, 4, 12),
			ChurnRiskScore: 0.30,
		},
		{
			ID:             "acc_stark",
			Name:           "Stark Industries AI",
			Domain:         "stark.tech",
			ARRUSD:         320000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 5, 0),
			ChurnRiskScore: 0.20,
		},
		{
			ID:             "acc_wayne",
			Name:           "Wayne Security Ops",
			Domain:         "waynesec.com",
			ARRUSD:         225000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 0, 28), // 28 days
			ChurnRiskScore: 0.80,
		},
		{
			ID:             "acc_umbrella",
			Name:           "Umbrella Health",
			Domain:         "umbrella.bio",
			ARRUSD:         38000.0,
			PlanTier:       models.PlanTierGrowth,
			RenewalDate:    now.AddDate(0, 6, 14),
			ChurnRiskScore: 0.25,
		},
		{
			ID:             "acc_piedpiper",
			Name:           "Pied Piper Storage",
			Domain:         "piedpiper.com",
			ARRUSD:         28000.0,
			PlanTier:       models.PlanTierGrowth,
			RenewalDate:    now.AddDate(0, 2, 20),
			ChurnRiskScore: 0.45,
		},
		{
			ID:             "acc_hooli",
			Name:           "Hooli Cloud",
			Domain:         "hooli.xyz",
			ARRUSD:         175000.0,
			PlanTier:       models.PlanTierEnterprise,
			RenewalDate:    now.AddDate(0, 1, 10), // 40 days
			ChurnRiskScore: 0.65,
		},
		{
			ID:             "acc_dundermifflin",
			Name:           "Dunder Mifflin Direct",
			Domain:         "dundermifflin.com",
			ARRUSD:         7500.0,
			PlanTier:       models.PlanTierStarter,
			RenewalDate:    now.AddDate(0, 8, 0),
			ChurnRiskScore: 0.15,
		},
	}

	count := 0
	for _, acc := range sampleAccounts {
		a := acc
		if err := s.UpsertAccount(ctx, &a); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
