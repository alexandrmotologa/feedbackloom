package ranking

import (
	"math"
	"sort"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

// OpportunityMetrics contains all calculated financial and urgency figures for an opportunity.
type OpportunityMetrics struct {
	TotalARRUSD      float64
	WeightedScore    float64
	AccountCount     int
	FeedbackCount    int
	EnterpriseARRUSD float64
	GrowthARRUSD     float64
	StarterARRUSD    float64
	AvgChurnRisk     float64
	MinDaysToRenewal int
	TopAccounts      []string
}

// CalculateOpportunityMetrics evaluates the financial ARR, urgency, and renewal impact of an opportunity cluster.
func CalculateOpportunityMetrics(items []models.EnrichedFeedbackItem) OpportunityMetrics {
	metrics := OpportunityMetrics{
		MinDaysToRenewal: 365,
		TopAccounts:      make([]string, 0),
	}

	if len(items) == 0 {
		return metrics
	}

	metrics.FeedbackCount = len(items)

	// Deduplicate by account ID to ensure each unique account's ARR is summed only once per opportunity
	uniqueAccounts := make(map[string]models.EnrichedFeedbackItem)
	anonymousCount := 0

	for _, item := range items {
		if item.AccountID != "" {
			if _, exists := uniqueAccounts[item.AccountID]; !exists {
				uniqueAccounts[item.AccountID] = item
			}
		} else {
			anonymousCount++
		}
	}

	metrics.AccountCount = len(uniqueAccounts)
	if metrics.AccountCount == 0 && anonymousCount > 0 {
		metrics.AccountCount = anonymousCount
	}

	var sumChurnRisk float64
	type accTuple struct {
		Name   string
		ARRUSD float64
	}
	var accList []accTuple

	for _, accItem := range uniqueAccounts {
		arr := accItem.AccountARRUSD
		metrics.TotalARRUSD += arr

		// Breakdown by tier
		if arr >= 50000.0 {
			metrics.EnterpriseARRUSD += arr
		} else if arr >= 10000.0 {
			metrics.GrowthARRUSD += arr
		} else {
			metrics.StarterARRUSD += arr
		}

		days := accItem.DaysToRenewal
		if days < 0 {
			days = 0
		}
		if days < metrics.MinDaysToRenewal {
			metrics.MinDaysToRenewal = days
		}

		churn := accItem.ChurnRiskScore
		if churn <= 0.0 {
			churn = 0.2
		}
		sumChurnRisk += churn

		// Formula: ARR * (1 + 1 / max(1, DaysToRenewal)) * ChurnRisk
		urgencyMultiplier := 1.0 + (1.0 / float64(int(math.Max(1.0, float64(days)))))
		accountWRS := arr * urgencyMultiplier * churn
		metrics.WeightedScore += accountWRS

		name := accItem.AccountName
		if name == "" {
			name = accItem.AccountDomain
		}
		if name != "" {
			accList = append(accList, accTuple{Name: name, ARRUSD: arr})
		}
	}

	if len(uniqueAccounts) > 0 {
		metrics.AvgChurnRisk = sumChurnRisk / float64(len(uniqueAccounts))
	}

	// Sort accounts by ARR descending to highlight top enterprise demand
	sort.Slice(accList, func(i, j int) bool {
		return accList[i].ARRUSD > accList[j].ARRUSD
	})

	for i := 0; i < len(accList) && i < 5; i++ {
		metrics.TopAccounts = append(metrics.TopAccounts, accList[i].Name)
	}

	return metrics
}
