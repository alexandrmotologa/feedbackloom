package crm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"cto@acme.com", "acme.com"},
		{"John Doe <john@cyberdyne.com>", "cyberdyne.com"},
		{"sarah@sub.corp.stripe.com", "stripe.com"},
		{"user@gmail.com", ""}, // consumer webmail rejected
		{"user@yahoo.com", ""},
		{"user@proton.me", ""},
		{"alice@initech.co.uk", "initech.co.uk"},
		{"invalid-email", ""},
		{"", ""},
	}

	for _, tc := range tests {
		got := ExtractDomain(tc.input)
		if got != tc.expected {
			t.Errorf("ExtractDomain(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestCSVImportAndResolver(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "feedbackloom_crm_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "crm_test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store new error: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	csvContent := `name,domain,arr_usd,plan_tier,renewal_date,churn_risk_score
Cyberdyne Systems,cyberdyne.com,250000,ENTERPRISE,2026-12-31,0.70
Initech Software,initech.com,45000,GROWTH,2027-01-15,0.30
`
	imported, err := ImportAccountsFromReader(ctx, s, strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("import accounts error: %v", err)
	}
	if imported != 2 {
		t.Fatalf("expected 2 accounts imported, got %d", imported)
	}

	resolver := NewResolver(s)
	acc, err := resolver.ResolveAccount(ctx, "cto@cyberdyne.com")
	if err != nil || acc == nil {
		t.Fatalf("expected account for cyberdyne, got err: %v, acc: %+v", err, acc)
	}
	if acc.ARRUSD != 250000 || acc.PlanTier != models.PlanTierEnterprise {
		t.Errorf("unexpected account data: %+v", acc)
	}

	// Consumer email should resolve to nil
	webmailAcc, err := resolver.ResolveAccount(ctx, "someone@gmail.com")
	if err != nil || webmailAcc != nil {
		t.Errorf("expected nil for consumer email, got %+v", webmailAcc)
	}
}

func TestSeedRealisticStripeAccounts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "feedbackloom_stripe_seed_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "stripe_test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store new error: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	count, err := SyncStripeCustomers(ctx, s, StripeSyncOptions{APIKey: "mock"})
	if err != nil {
		t.Fatalf("sync stripe customers error: %v", err)
	}
	if count < 10 {
		t.Fatalf("expected at least 10 seeded accounts, got %d", count)
	}

	accs, err := s.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("list accounts error: %v", err)
	}
	if len(accs) != count {
		t.Errorf("expected %d accounts, got %d", count, len(accs))
	}

	// Verify days to renewal calculation
	cyberdyne, err := s.GetAccountByDomain(ctx, "cyberdyne.com")
	if err != nil {
		t.Fatalf("get cyberdyne error: %v", err)
	}
	days := cyberdyne.DaysToRenewal()
	if days <= 0 || days > 365 {
		t.Errorf("expected reasonable days to renewal, got %d", days)
	}
	_ = time.Now()
}
