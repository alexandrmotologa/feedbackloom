package prd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

func TestPRDGeneration(t *testing.T) {
	gen := NewPRDGenerator()
	ctx := context.Background()

	opp := &models.Opportunity{
		ID:            "opp_test_sso",
		Title:         "Enterprise SAML 2.0 & Okta SSO Support",
		Description:   "Enterprise customer accounts demand native SAML 2.0 and Okta/Azure AD single sign-on integration.",
		Category:      models.CategorySecurity,
		TotalARRUSD:   480000.0,
		WeightedScore: 520000.0,
		AccountCount:  2,
		FeedbackCount: 2,
		Status:        models.StatusDiscovered,
	}

	quotes := []models.EnrichedFeedbackItem{
		{
			FeedbackItem: models.FeedbackItem{
				ID:            "fb_1",
				Source:        models.SourceZendesk,
				CustomerEmail: "cto@cyberdyne.com",
				CustomerName:  "Miles Dyson",
				RawText:       "We cannot proceed with our enterprise rollout without Okta SAML 2.0 single sign-on.",
				UrgencyScore:  0.95,
				CreatedAt:     time.Now().Add(-24 * time.Hour),
			},
			AccountName:    "Cyberdyne Systems",
			AccountDomain:  "cyberdyne.com",
			AccountARRUSD:  285000.0,
			PlanTier:       models.PlanTierEnterprise,
			DaysToRenewal:  45,
			ChurnRiskScore: 0.75,
		},
		{
			FeedbackItem: models.FeedbackItem{
				ID:            "fb_2",
				Source:        models.SourceIntercom,
				CustomerEmail: "rachel@acme.com",
				CustomerName:  "Rachel Green",
				RawText:       "SAML SSO is required for Azure AD compliance.",
				UrgencyScore:  0.80,
				CreatedAt:     time.Now().Add(-12 * time.Hour),
			},
			AccountName:    "Acme Global Tech",
			AccountDomain:  "acme.com",
			AccountARRUSD:  195000.0,
			PlanTier:       models.PlanTierEnterprise,
			DaysToRenewal:  65,
			ChurnRiskScore: 0.60,
		},
	}

	md, notionJSON, err := gen.GeneratePRD(ctx, opp, quotes)
	if err != nil {
		t.Fatalf("generate PRD error: %v", err)
	}

	// Verify required sections
	requiredSections := []string{
		"# PRD: Enterprise SAML 2.0 & Okta SSO Support",
		"## 1. Executive Summary and Business Context",
		"## 2. Problem Statement",
		"## 3. Verified Customer Evidence Quotes",
		"Cyberdyne Systems",
		"$285000 ARR",
		"## 4. Target Personas",
		"## 5. User Stories and Gherkin Acceptance Criteria",
		"```gherkin",
		"Feature: SAML 2.0 Identity Provider Configuration",
		"## 6. Technical and Non-Functional Requirements",
		"## 7. Out of Scope Guardrails",
		"## 8. Rollout Plan and Success Metrics",
	}

	for _, sec := range requiredSections {
		if !strings.Contains(md, sec) {
			t.Errorf("expected markdown to contain %q", sec)
		}
	}

	// Humanizer check: no em dashes or en dashes
	if strings.Contains(md, "—") || strings.Contains(md, "–") {
		t.Errorf("markdown contains em/en dashes, violating humanizer discipline")
	}

	// Notion JSON verification
	var parsedNotion struct {
		Title    string        `json:"title"`
		Children []interface{} `json:"children"`
	}
	if err := json.Unmarshal([]byte(notionJSON), &parsedNotion); err != nil {
		t.Fatalf("notion JSON unmarshal error: %v", err)
	}
	if len(parsedNotion.Children) < 5 {
		t.Errorf("expected at least 5 notion blocks, got %d", len(parsedNotion.Children))
	}

	// Verify Notion export mock
	url, err := ExportToNotionAPI(ctx, opp.Title, notionJSON, "mock", "page_123")
	if err != nil {
		t.Fatalf("notion mock export error: %v", err)
	}
	if !strings.HasPrefix(url, "https://notion.so/") {
		t.Errorf("unexpected notion URL: %s", url)
	}
}
