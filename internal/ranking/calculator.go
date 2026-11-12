package ranking

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

// OpportunityMetrics contains all calculated financial, competitive, effort, and urgency figures.
type OpportunityMetrics struct {
	TotalARRUSD        float64
	WeightedScore      float64
	AccountCount       int
	FeedbackCount      int
	EnterpriseARRUSD   float64
	GrowthARRUSD       float64
	StarterARRUSD      float64
	AvgChurnRisk       float64
	MinDaysToRenewal   int
	TopAccounts        []string
	CompetitorMentions []string
	LostPipelineARR    float64
	EffortSprints      int
	ROIIndex           float64
	UrgentRenewalAlert bool
	VelocityGrowthPct  float64
}

// CalculateOpportunityMetrics evaluates the financial ARR, competitive loss, urgency, and ROI impact.
func CalculateOpportunityMetrics(items []models.EnrichedFeedbackItem) OpportunityMetrics {
	metrics := OpportunityMetrics{
		MinDaysToRenewal:   365,
		TopAccounts:        make([]string, 0),
		CompetitorMentions: make([]string, 0),
		EffortSprints:      2, // Default 2 sprints (Medium)
	}

	if len(items) == 0 {
		return metrics
	}

	metrics.FeedbackCount = len(items)

	// Deduplicate by account ID to ensure each unique account's ARR is summed only once per opportunity
	uniqueAccounts := make(map[string]models.EnrichedFeedbackItem)
	anonymousCount := 0

	// Track competitor mentions and velocity
	competitorMap := make(map[string]bool)
	recentItemsCount := 0
	cutoff := time.Now().Add(-14 * 24 * time.Hour)

	for _, item := range items {
		if item.AccountID != "" {
			if _, exists := uniqueAccounts[item.AccountID]; !exists {
				uniqueAccounts[item.AccountID] = item
			}
		} else {
			anonymousCount++
		}

		for _, comp := range item.CompetitorMentions {
			if !competitorMap[comp] {
				competitorMap[comp] = true
				metrics.CompetitorMentions = append(metrics.CompetitorMentions, comp)
			}
		}

		if item.CreatedAt.After(cutoff) {
			recentItemsCount++
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

		// Churn risk >= 0.6 counts towards lost/at-risk pipeline
		if churn >= 0.6 {
			metrics.LostPipelineARR += arr
		}

		// Urgent renewal alert if enterprise account renewal <= 45 days
		if arr >= 100000.0 && days <= 45 {
			metrics.UrgentRenewalAlert = true
		}

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

	// Infer engineering effort in sprints based on scope keywords
	textCorpus := ""
	for _, it := range items {
		textCorpus += " " + strings.ToLower(it.RawText)
	}

	if strings.Contains(textCorpus, "sso") || strings.Contains(textCorpus, "saml") {
		metrics.EffortSprints = 4 // Large (4 sprints)
	} else if strings.Contains(textCorpus, "audit") || strings.Contains(textCorpus, "siem") {
		metrics.EffortSprints = 2 // Medium (2 sprints)
	} else if strings.Contains(textCorpus, "rbac") || strings.Contains(textCorpus, "permission") {
		metrics.EffortSprints = 3 // Medium-Large (3 sprints)
	} else if strings.Contains(textCorpus, "export") || strings.Contains(textCorpus, "bulk") {
		metrics.EffortSprints = 2 // Medium (2 sprints)
	} else {
		metrics.EffortSprints = 1 // Small (1 sprint)
	}

	// Calculate ROI Index: ARR per Engineering Sprint
	metrics.ROIIndex = metrics.TotalARRUSD / float64(int(math.Max(1.0, float64(metrics.EffortSprints))))

	// Velocity growth calculation (% recent vs baseline)
	if len(items) > 0 {
		metrics.VelocityGrowthPct = float64(recentItemsCount) / float64(len(items)) * 100.0
	}

	return metrics
}
