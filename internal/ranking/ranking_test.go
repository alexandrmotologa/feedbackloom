package ranking

import (
	"testing"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

func TestOpportunityMetricsCalculation(t *testing.T) {
	items := []models.EnrichedFeedbackItem{
		{
			FeedbackItem: models.FeedbackItem{
				ID:        "fb_1",
				AccountID: "acc_enterprise",
			},
			AccountName:    "Enterprise Corp",
			AccountARRUSD:  200000.0,
			PlanTier:       models.PlanTierEnterprise,
			DaysToRenewal:  20,
			ChurnRiskScore: 0.8,
		},
		{
			// Another feedback from the same account - must not double-count ARR
			FeedbackItem: models.FeedbackItem{
				ID:        "fb_2",
				AccountID: "acc_enterprise",
			},
			AccountName:    "Enterprise Corp",
			AccountARRUSD:  200000.0,
			PlanTier:       models.PlanTierEnterprise,
			DaysToRenewal:  20,
			ChurnRiskScore: 0.8,
		},
		{
			FeedbackItem: models.FeedbackItem{
				ID:        "fb_3",
				AccountID: "acc_growth",
			},
			AccountName:    "Growth Startup",
			AccountARRUSD:  30000.0,
			PlanTier:       models.PlanTierGrowth,
			DaysToRenewal:  90,
			ChurnRiskScore: 0.3,
		},
	}

	metrics := CalculateOpportunityMetrics(items)

	// Unique accounts should be 2, total feedback 3
	if metrics.AccountCount != 2 {
		t.Errorf("expected 2 unique accounts, got %d", metrics.AccountCount)
	}
	if metrics.FeedbackCount != 3 {
		t.Errorf("expected 3 feedback items, got %d", metrics.FeedbackCount)
	}

	// Total ARR = 200,000 + 30,000 = 230,000
	if metrics.TotalARRUSD != 230000.0 {
		t.Errorf("expected Total ARR 230000.0, got %f", metrics.TotalARRUSD)
	}
	if metrics.EnterpriseARRUSD != 200000.0 {
		t.Errorf("expected Enterprise ARR 200000.0, got %f", metrics.EnterpriseARRUSD)
	}
	if metrics.GrowthARRUSD != 30000.0 {
		t.Errorf("expected Growth ARR 30000.0, got %f", metrics.GrowthARRUSD)
	}

	// Min days to renewal should be 20
	if metrics.MinDaysToRenewal != 20 {
		t.Errorf("expected min days 20, got %d", metrics.MinDaysToRenewal)
	}

	// Weighted score should be > 0 and reflect the higher urgency of Enterprise Corp
	if metrics.WeightedScore <= 0 {
		t.Errorf("expected positive weighted score, got %f", metrics.WeightedScore)
	}
	if len(metrics.TopAccounts) != 2 || metrics.TopAccounts[0] != "Enterprise Corp" {
		t.Errorf("unexpected top accounts: %+v", metrics.TopAccounts)
	}
}
