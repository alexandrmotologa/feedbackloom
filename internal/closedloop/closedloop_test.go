package closedloop

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

func TestClosedLoopWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "feedbackloom_closedloop_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "closedloop.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store new error: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	// 1. Account and feedback setup
	acc := &models.Account{
		ID:       "acc_cyberdyne",
		Name:     "Cyberdyne Systems",
		Domain:   "cyberdyne.com",
		ARRUSD:   285000.0,
		PlanTier: models.PlanTierEnterprise,
	}
	_ = s.UpsertAccount(ctx, acc)

	fb := &models.FeedbackItem{
		ID:            "fb_sso",
		Source:        models.SourceZendesk,
		AccountID:     acc.ID,
		CustomerEmail: "cto@cyberdyne.com",
		CustomerName:  "Miles Dyson",
		RawText:       "We cannot renew without Okta SAML SSO support.",
	}
	_ = s.InsertFeedbackItem(ctx, fb)

	opp := &models.Opportunity{
		ID:          "opp_sso",
		Title:       "Enterprise SAML 2.0 & Okta SSO Support",
		Description: "Native SAML 2.0 identity provider integration and Okta authentication.",
		Category:    models.CategorySecurity,
		TotalARRUSD: 285000.0,
		Status:      models.StatusShipped,
	}
	_ = s.UpsertOpportunity(ctx, opp)
	_ = s.LinkFeedbackOpportunity(ctx, fb.ID, opp.ID, 0.99)

	svc := NewService(s)

	// 2. Generate notifications
	notifs, err := svc.GenerateNotifications(ctx, opp.ID)
	if err != nil {
		t.Fatalf("generate notifications error: %v", err)
	}
	if len(notifs) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifs))
	}

	n := notifs[0]
	if n.CustomerEmail != "cto@cyberdyne.com" || n.CustomerName != "Miles Dyson" {
		t.Errorf("unexpected recipient: %+v", n)
	}
	if !strings.Contains(n.Body, "Miles Dyson") || !strings.Contains(n.Body, "Cyberdyne Systems") || !strings.Contains(n.Body, "Okta SAML SSO") {
		t.Errorf("expected personalized body, got:\n%s", n.Body)
	}

	// 3. Dispatch
	sent, err := svc.DispatchAllPending(ctx, opp.ID)
	if err != nil {
		t.Fatalf("dispatch error: %v", err)
	}
	if sent != 1 {
		t.Errorf("expected 1 sent notification, got %d", sent)
	}

	storedNotifs, _ := s.ListNotificationsByOpportunity(ctx, opp.ID)
	if storedNotifs[0].Status != "SENT" || storedNotifs[0].SentAt == nil {
		t.Errorf("expected status SENT with timestamp, got %+v", storedNotifs[0])
	}
}
