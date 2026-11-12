package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
)

func TestVectorEncodeDecode(t *testing.T) {
	vec := []float32{0.123, -0.456, 0.789, 1.0, -1.0}
	encoded := EncodeFloat32Slice(vec)
	decoded := DecodeFloat32Slice(encoded)

	if len(decoded) != len(vec) {
		t.Fatalf("expected len %d, got %d", len(vec), len(decoded))
	}
	for i := range vec {
		diff := decoded[i] - vec[i]
		if diff < -0.0001 || diff > 0.0001 {
			t.Errorf("at index %d: expected %f, got %f", i, vec[i], decoded[i])
		}
	}
}

func TestStoreOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "feedbackloom_store_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	// 1. Account upsert and retrieval
	renewal := time.Now().Add(45 * 24 * time.Hour)
	acc := &models.Account{
		ID:             "acc_101",
		Name:           "Acme Technologies",
		Domain:         "acme.com",
		ARRUSD:         120000.0,
		PlanTier:       models.PlanTierEnterprise,
		RenewalDate:    renewal,
		ChurnRiskScore: 0.65,
	}

	if err := s.UpsertAccount(ctx, acc); err != nil {
		t.Fatalf("upsert account error: %v", err)
	}

	fetchedByDomain, err := s.GetAccountByDomain(ctx, "acme.com")
	if err != nil {
		t.Fatalf("get account by domain error: %v", err)
	}
	if fetchedByDomain.Name != "Acme Technologies" || fetchedByDomain.ARRUSD != 120000.0 {
		t.Errorf("unexpected account data: %+v", fetchedByDomain)
	}

	// 2. Feedback insertion
	fb := &models.FeedbackItem{
		ID:                 "fb_001",
		Source:             models.SourceZendesk,
		AccountID:          acc.ID,
		CustomerEmail:      "cto@acme.com",
		CustomerName:       "Jane Doe",
		RawText:            "We need Okta SAML SSO urgently. If we can't enforce SSO next month, our security team will block renewal.",
		ExtractedPainPoint: "Lack of SAML 2.0 / Okta single sign-on",
		ExtractedOutcome:   "Enterprise SAML SSO integration",
		SentimentScore:     -0.8,
		UrgencyScore:       0.95,
		Embedding:          []float32{0.5, 0.2, -0.1},
	}
	if err := s.InsertFeedbackItem(ctx, fb); err != nil {
		t.Fatalf("insert feedback error: %v", err)
	}

	fetchedFB, err := s.GetFeedbackItemByID(ctx, "fb_001")
	if err != nil {
		t.Fatalf("get feedback item error: %v", err)
	}
	if fetchedFB.CustomerEmail != "cto@acme.com" || len(fetchedFB.Embedding) != 3 {
		t.Errorf("unexpected feedback data: %+v", fetchedFB)
	}

	// 3. Opportunity creation and linking
	opp := &models.Opportunity{
		ID:            "opp_sso",
		Title:         "Enterprise SAML 2.0 & Okta SSO",
		Description:   "Provide native Okta and Azure AD single sign-on authentication.",
		Category:      models.CategorySecurity,
		TotalARRUSD:   120000.0,
		WeightedScore: 120000.0 * (1.0 + 1.0/45.0) * 0.65,
		AccountCount:  1,
		FeedbackCount: 1,
		Status:        models.StatusDiscovered,
	}
	if err := s.UpsertOpportunity(ctx, opp); err != nil {
		t.Fatalf("upsert opportunity error: %v", err)
	}

	if err := s.LinkFeedbackOpportunity(ctx, fb.ID, opp.ID, 0.98); err != nil {
		t.Fatalf("link feedback opportunity error: %v", err)
	}

	// 4. Enriched feedback query
	enriched, err := s.ListEnrichedFeedbackForOpportunity(ctx, opp.ID)
	if err != nil {
		t.Fatalf("list enriched feedback error: %v", err)
	}
	if len(enriched) != 1 {
		t.Fatalf("expected 1 enriched feedback item, got %d", len(enriched))
	}
	if enriched[0].AccountName != "Acme Technologies" || enriched[0].AccountARRUSD != 120000.0 {
		t.Errorf("unexpected enriched feedback data: %+v", enriched[0])
	}

	// 5. Update PRD
	if err := s.UpdateOpportunityPRD(ctx, opp.ID, "# PRD for SAML SSO\nDetailed spec...", `{"blocks":[]}`); err != nil {
		t.Fatalf("update PRD error: %v", err)
	}
	updatedOpp, err := s.GetOpportunityByID(ctx, opp.ID)
	if err != nil {
		t.Fatalf("get opp error: %v", err)
	}
	if updatedOpp.Status != models.StatusInPRD || updatedOpp.GeneratedPRDMarkdown == "" {
		t.Errorf("expected status IN_PRD and non-empty markdown, got status %s", updatedOpp.Status)
	}

	// 6. Notifications
	notif := &models.ClosedLoopNotification{
		ID:            "notif_001",
		OpportunityID: opp.ID,
		AccountID:     acc.ID,
		CustomerEmail: "cto@acme.com",
		CustomerName:  "Jane Doe",
		Subject:       "Update on Okta SAML SSO",
		Body:          "Hi Jane, we have completed the SAML SSO specification.",
		Status:        "DRAFT",
	}
	if err := s.InsertNotification(ctx, notif); err != nil {
		t.Fatalf("insert notification error: %v", err)
	}

	notifs, err := s.ListNotificationsByOpportunity(ctx, opp.ID)
	if err != nil {
		t.Fatalf("list notifications error: %v", err)
	}
	if len(notifs) != 1 || notifs[0].Subject != "Update on Okta SAML SSO" {
		t.Errorf("unexpected notifications: %+v", notifs)
	}

	// 7. Dashboard stats
	stats, err := s.GetDashboardStats(ctx)
	if err != nil {
		t.Fatalf("get dashboard stats error: %v", err)
	}
	if stats["total_feedback"].(int) != 1 || stats["total_accounts"].(int) != 1 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}
