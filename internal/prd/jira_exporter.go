package prd

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

// GenerateJiraEpic formats an opportunity into a Jira Epic description with native Jira markup.
func GenerateJiraEpic(opp *models.Opportunity, quotes []models.EnrichedFeedbackItem) models.JiraEpicPayload {
	var sb strings.Builder

	sb.WriteString("h1. Business Summary & Financial Weighting\n\n")
	sb.WriteString(fmt.Sprintf(
		"*Total ARR at Stake:* *USD $%.2f* across *%d* verified customer accounts.\n",
		opp.TotalARRUSD, opp.AccountCount,
	))
	sb.WriteString(fmt.Sprintf(
		"*Weighted Revenue Score (WRS):* %.2f | *Estimated Effort:* %d Sprints | *ROI Index:* $%.2f ARR/sprint\n\n",
		opp.WeightedScore, opp.EffortSprints, opp.ROIIndex,
	))

	sb.WriteString("h2. Problem Statement\n\n")
	sb.WriteString(opp.Description + "\n\n")

	sb.WriteString("h2. Customer Evidence Quotes\n\n")
	sb.WriteString("|| Account || Plan || ARR || Renewal || Quote ||\n")

	for i, q := range quotes {
		if i >= 5 {
			break
		}
		name := q.AccountName
		if name == "" {
			name = q.AccountDomain
		}
		clean := strings.ReplaceAll(strings.TrimSpace(q.RawText), "\n", " ")
		clean = strings.ReplaceAll(clean, "|", "/")
		if len(clean) > 80 {
			clean = clean[:80] + "..."
		}
		sb.WriteString(fmt.Sprintf(
			"| %s | %s | $%.0f | %d days | \"%s\" |\n",
			name, q.PlanTier, q.AccountARRUSD, q.DaysToRenewal, clean,
		))
	}
	sb.WriteString("\n")

	sb.WriteString("h2. User Stories & Gherkin Acceptance Criteria\n\n")
	sb.WriteString("{code:gherkin}\n")
	sb.WriteString(fmt.Sprintf("Feature: %s\n\n", opp.Title))
	sb.WriteString("  Scenario: Standard verified customer execution\n")
	sb.WriteString("    Given an authenticated enterprise account\n")
	sb.WriteString("    When the capability is configured in workspace settings\n")
	sb.WriteString("    Then security validation succeeds with 99.95% availability\n")
	sb.WriteString("{code}\n\n")

	sb.WriteString("h2. Technical Guardrails\n\n")
	sb.WriteString("* All credentials encrypted at rest with AES-256-GCM and in transit via TLS 1.3.\n")
	sb.WriteString("* Immutable audit log emitted for every administrative configuration update.\n")

	priority := "Medium"
	if opp.TotalARRUSD >= 200000.0 || opp.UrgentRenewalAlert {
		priority = "Highest"
	} else if opp.TotalARRUSD >= 50000.0 {
		priority = "High"
	}

	return models.JiraEpicPayload{
		Summary:     fmt.Sprintf("[Epic] %s ($%.0fk ARR)", opp.Title, opp.TotalARRUSD/1000.0),
		Description: sb.String(),
		IssueType:   "Epic",
		Priority:    priority,
	}
}
