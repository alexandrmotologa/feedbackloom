package cluster

import (
	"context"
	"strings"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

// ExtractionResult holds NLP extraction details from raw feedback text.
type ExtractionResult struct {
	PainPoint      string
	DesiredOutcome string
	Category       string
	SentimentScore float64 // -1.0 to 1.0
	UrgencyScore   float64 // 0.0 to 1.0
}

// FeedbackExtractor extracts actionable insights from customer quotes.
type FeedbackExtractor struct{}

// NewFeedbackExtractor initializes a new extractor.
func NewFeedbackExtractor() *FeedbackExtractor {
	return &FeedbackExtractor{}
}

// Extract processes a raw text quote and extracts structured pain points and urgency.
func (e *FeedbackExtractor) Extract(ctx context.Context, rawText string) ExtractionResult {
	lower := strings.ToLower(rawText)

	res := ExtractionResult{
		SentimentScore: -0.2, // Default neutral-frustrated
		UrgencyScore:   0.3,  // Baseline urgency
		Category:       models.CategoryUsability,
	}

	// 1. Urgency calculation
	if strings.Contains(lower, "cancel") || strings.Contains(lower, "block renewal") || strings.Contains(lower, "cannot renew") {
		res.UrgencyScore = 0.95
		res.SentimentScore = -0.9
	} else if strings.Contains(lower, "urgent") || strings.Contains(lower, "asap") || strings.Contains(lower, "dealbreaker") || strings.Contains(lower, "mandate") {
		res.UrgencyScore = 0.85
		res.SentimentScore = -0.7
	} else if strings.Contains(lower, "blocker") || strings.Contains(lower, "must have") || strings.Contains(lower, "security audit") || strings.Contains(lower, "compliance") {
		res.UrgencyScore = 0.75
		res.SentimentScore = -0.5
	} else if strings.Contains(lower, "could we get") || strings.Contains(lower, "would be great") || strings.Contains(lower, "nice to have") {
		res.UrgencyScore = 0.25
		res.SentimentScore = 0.3
	}

	// 2. Category & Pain Point detection
	switch {
	case strings.Contains(lower, "sso") || strings.Contains(lower, "saml") || strings.Contains(lower, "okta") || strings.Contains(lower, "azure ad") || strings.Contains(lower, "login"):
		res.Category = models.CategorySecurity
		res.PainPoint = "Absence of enterprise SAML 2.0 / Okta single sign-on authentication"
		res.DesiredOutcome = "Native SAML 2.0 identity provider integration and SCIM user provisioning"

	case strings.Contains(lower, "audit") || strings.Contains(lower, "siem") || strings.Contains(lower, "soc 2") || strings.Contains(lower, "splunk") || strings.Contains(lower, "datadog"):
		res.Category = models.CategorySecurity
		res.PainPoint = "Missing automated SIEM audit logs and immutable compliance event tracking"
		res.DesiredOutcome = "Real-time streaming audit webhooks and SOC 2 Type II audit trail export"

	case strings.Contains(lower, "rbac") || strings.Contains(lower, "role") || strings.Contains(lower, "permission") || strings.Contains(lower, "scope") || strings.Contains(lower, "admin"):
		res.Category = models.CategorySecurity
		res.PainPoint = "Overly broad administrative permissions without granular role-based scopes"
		res.DesiredOutcome = "Configurable custom role definitions with granular read/write permission scopes"

	case strings.Contains(lower, "export") || strings.Contains(lower, "bulk") || strings.Contains(lower, "timeout") || strings.Contains(lower, "csv") || strings.Contains(lower, "sftp"):
		res.Category = models.CategoryPerformance
		res.PainPoint = "API timeouts and failure when exporting large data volumes"
		res.DesiredOutcome = "Asynchronous bulk export API with signed cloud download URLs and scheduled deliveries"

	case strings.Contains(lower, "webhook") || strings.Contains(lower, "integration") || strings.Contains(lower, "slack") || strings.Contains(lower, "teams"):
		res.Category = models.CategoryIntegrations
		res.PainPoint = "Limited third-party ecosystem integrations and real-time webhook delivery"
		res.DesiredOutcome = "Configurable webhook subscriptions with retry mechanisms and Slack/Teams alerts"

	default:
		res.Category = models.CategoryUsability
		res.PainPoint = cleanSentence(rawText, 120)
		res.DesiredOutcome = "Product enhancement addressing customer friction"
	}

	return res
}

func cleanSentence(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
