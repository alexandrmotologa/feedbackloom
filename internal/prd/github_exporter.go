package prd

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

// GenerateGitHubIssue formats an opportunity into a GitHub Issue with Gherkin task lists.
func GenerateGitHubIssue(opp *models.Opportunity, quotes []models.EnrichedFeedbackItem) models.GitHubIssuePayload {
	var sb strings.Builder

	// Header Summary Callout
	sb.WriteString("### Business Value & Revenue Context\n\n")
	sb.WriteString(fmt.Sprintf(
		"> **Total ARR at Stake**: $%.2f USD across %d requesting customer accounts.\n",
		opp.TotalARRUSD, opp.AccountCount,
	))
	sb.WriteString(fmt.Sprintf(
		"> **Weighted Revenue Score (WRS)**: %.2f | **Estimated Effort**: %d Sprints | **ROI Index**: $%.2f ARR/sprint\n\n",
		opp.WeightedScore, opp.EffortSprints, opp.ROIIndex,
	))

	// Problem description
	sb.WriteString("### Problem Statement\n\n")
	sb.WriteString(opp.Description)
	sb.WriteString("\n\n")

	// Customer evidence table
	sb.WriteString("### Verified Customer Evidence Quotes\n\n")
	sb.WriteString("| Customer Account | Plan Tier | ARR (USD) | Days to Renewal | Verbatim Quote |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")

	for i, q := range quotes {
		if i >= 5 {
			break
		}
		name := q.AccountName
		if name == "" {
			name = q.AccountDomain
		}
		cleanQuote := strings.ReplaceAll(strings.TrimSpace(q.RawText), "\n", " ")
		cleanQuote = strings.ReplaceAll(cleanQuote, "|", "/")
		if len(cleanQuote) > 100 {
			cleanQuote = cleanQuote[:100] + "..."
		}
		sb.WriteString(fmt.Sprintf(
			"| **%s** | `%s` | $%.0f | %d days | \"%s\" |\n",
			name, q.PlanTier, q.AccountARRUSD, q.DaysToRenewal, cleanQuote,
		))
	}
	sb.WriteString("\n")

	// User stories and Gherkin task checklist
	sb.WriteString("### Engineering Tasks & Acceptance Criteria Checklist\n\n")
	sb.WriteString("- [ ] **Architecture Design & Threat Model Review**\n")
	sb.WriteString("- [ ] **Core API & Service Layer Implementation**\n")

	switch opp.Category {
	case models.CategorySecurity:
		if strings.Contains(strings.ToLower(opp.Title), "sso") {
			sb.WriteString("- [ ] **Acceptance Test**: Valid SAML 2.0 metadata URL imports certificate successfully\n")
			sb.WriteString("- [ ] **Acceptance Test**: Enforcing SSO redirects domain email users to Okta sign-in\n")
			sb.WriteString("- [ ] **Acceptance Test**: SCIM directory provisioning creates and disables workspace seats\n")
		} else {
			sb.WriteString("- [ ] **Acceptance Test**: Streaming audit webhook dispatches signed JSON events\n")
			sb.WriteString("- [ ] **Acceptance Test**: Actor ID, IP address, and timestamp are immutably logged\n")
		}
	case models.CategoryPerformance:
		sb.WriteString("- [ ] **Acceptance Test**: Async export returns HTTP 202 Accepted with unique job ID\n")
		sb.WriteString("- [ ] **Acceptance Test**: Presigned cloud download link generates and delivers webhook\n")
	default:
		sb.WriteString("- [ ] **Acceptance Test**: Core user story end-to-end integration test passes\n")
	}

	sb.WriteString("- [ ] **Security & InfoSec Compliance Sign-off**\n")
	sb.WriteString("- [ ] **Automated Closed-Loop Customer Email Broadcast Triggered**\n\n")

	// Guardrails
	sb.WriteString("### Out of Scope Guardrails\n\n")
	sb.WriteString("- Legacy proprietary protocols and unencrypted transmission are strictly out of scope.\n")
	sb.WriteString("- Custom on-premises hardware security integrations are deferred to subsequent releases.\n")

	// Labels
	labels := []string{
		"feedbackloom",
		fmt.Sprintf("category:%s", strings.ToLower(opp.Category)),
		fmt.Sprintf("effort:%dsprints", opp.EffortSprints),
	}
	if opp.TotalARRUSD >= 100000.0 {
		labels = append(labels, "tier:enterprise-blocker")
	}
	if opp.UrgentRenewalAlert {
		labels = append(labels, "priority:renewal-danger")
	}

	return models.GitHubIssuePayload{
		Title:  fmt.Sprintf("[Feature] %s ($%.0fk ARR)", opp.Title, opp.TotalARRUSD/1000.0),
		Body:   sb.String(),
		Labels: labels,
	}
}
