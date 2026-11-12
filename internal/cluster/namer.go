package cluster

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

// SynthesizedTheme holds the generated title, description, and category for a cluster.
type SynthesizedTheme struct {
	Title       string
	Description string
	Category    string
}

// SynthesizeOpportunityTheme generates an actionable product opportunity name and description
// based on the underlying customer quotes in the cluster.
func SynthesizeOpportunityTheme(items []models.EnrichedFeedbackItem) SynthesizedTheme {
	if len(items) == 0 {
		return SynthesizedTheme{
			Title:       "General Feedback Theme",
			Description: "Customer feedback requiring product evaluation.",
			Category:    models.CategoryUsability,
		}
	}

	// Tally keyword occurrences across all items in cluster
	ssoCount := 0
	auditCount := 0
	rbacCount := 0
	exportCount := 0
	integrationCount := 0

	totalUrgentAccounts := 0
	for _, it := range items {
		lower := strings.ToLower(it.RawText + " " + it.ExtractedPainPoint)
		if strings.Contains(lower, "sso") || strings.Contains(lower, "saml") || strings.Contains(lower, "okta") || strings.Contains(lower, "azure ad") {
			ssoCount++
		}
		if strings.Contains(lower, "audit") || strings.Contains(lower, "siem") || strings.Contains(lower, "soc 2") || strings.Contains(lower, "splunk") || strings.Contains(lower, "datadog") {
			auditCount++
		}
		if strings.Contains(lower, "rbac") || strings.Contains(lower, "role") || strings.Contains(lower, "permission") || strings.Contains(lower, "scope") {
			rbacCount++
		}
		if strings.Contains(lower, "export") || strings.Contains(lower, "bulk") || strings.Contains(lower, "csv") || strings.Contains(lower, "timeout") {
			exportCount++
		}
		if strings.Contains(lower, "webhook") || strings.Contains(lower, "slack") || strings.Contains(lower, "teams") {
			integrationCount++
		}
		if it.UrgencyScore >= 0.7 {
			totalUrgentAccounts++
		}
	}

	switch {
	case ssoCount >= auditCount && ssoCount >= rbacCount && ssoCount >= exportCount && ssoCount > 0:
		return SynthesizedTheme{
			Title:       "Enterprise SAML 2.0 & Okta SSO Support",
			Description: fmt.Sprintf("Enterprise customer accounts demand native SAML 2.0 and Okta/Azure AD single sign-on integration. Enterprise security mandates require centralized identity management to unblock contract renewals across %d accounts.", len(items)),
			Category:    models.CategorySecurity,
		}

	case auditCount >= rbacCount && auditCount >= exportCount && auditCount > 0:
		return SynthesizedTheme{
			Title:       "Real-Time SIEM Audit Log Streaming & Security Webhooks",
			Description: fmt.Sprintf("Customer compliance teams require automated, immutable audit log streaming to Datadog, Splunk, and cloud security lakes to satisfy annual SOC 2 Type II audits. Demanded by %d customer accounts.", len(items)),
			Category:    models.CategorySecurity,
		}

	case rbacCount >= exportCount && rbacCount > 0:
		return SynthesizedTheme{
			Title:       "Granular Role-Based Access Control (RBAC) & Custom Scopes",
			Description: fmt.Sprintf("Customers require custom role permissions allowing external contractors and departmental managers restricted view-only or workflow approval rights without full administrative access. Requested across %d accounts.", len(items)),
			Category:    models.CategorySecurity,
		}

	case exportCount > 0:
		return SynthesizedTheme{
			Title:       "High-Performance Asynchronous Bulk Data Export API",
			Description: fmt.Sprintf("Customers experience timeout errors when querying large datasets via standard REST endpoints. They require an asynchronous batch export API with signed S3 download links and scheduled deliveries. Requested across %d accounts.", len(items)),
			Category:    models.CategoryPerformance,
		}

	case integrationCount > 0:
		return SynthesizedTheme{
			Title:       "Third-Party Ecosystem Webhooks & Alerting",
			Description: fmt.Sprintf("Automated real-time event notifications to Slack, Microsoft Teams, and custom webhooks with retry semantics. Demanded by %d customer accounts.", len(items)),
			Category:    models.CategoryIntegrations,
		}

	default:
		// Synthesize from the highest ARR account's extracted pain point
		bestPainPoint := items[0].ExtractedPainPoint
		if bestPainPoint == "" {
			bestPainPoint = cleanSentence(items[0].RawText, 80)
		}
		return SynthesizedTheme{
			Title:       fmt.Sprintf("Product Initiative: %s", bestPainPoint),
			Description: fmt.Sprintf("Discovered theme reflecting feedback from %d customer accounts requesting capabilities around %s.", len(items), bestPainPoint),
			Category:    models.CategoryUsability,
		}
	}
}
