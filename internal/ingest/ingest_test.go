package ingest

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

func setupTestStore(t *testing.T) (*store.Store, *crm.Resolver, func()) {
	tmpDir, err := os.MkdirTemp("", "feedbackloom_ingest_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "ingest_test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store new error: %v", err)
	}

	ctx := context.Background()
	_, err = crm.SyncStripeCustomers(ctx, s, crm.StripeSyncOptions{APIKey: "mock"})
	if err != nil {
		t.Fatalf("seed accounts error: %v", err)
	}

	resolver := crm.NewResolver(s)
	cleanup := func() {
		s.Close()
		os.RemoveAll(tmpDir)
	}

	return s, resolver, cleanup
}

func TestCSVLoader(t *testing.T) {
	s, resolver, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	csvData := `source,customer_email,customer_name,raw_text
ZENDESK,cto@cyberdyne.com,Miles Dyson,We must have Okta SAML SSO support.
SLACK,lucius.fox@waynesec.com,Lucius Fox,Please stream SIEM audit logs.
`
	count, err := ImportFeedbackFromReader(ctx, s, resolver, strings.NewReader(csvData), models.SourceCSV)
	if err != nil {
		t.Fatalf("import csv error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 items, got %d", count)
	}

	items, err := s.ListFeedbackItems(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list items error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 stored items, got %d", len(items))
	}
	for _, it := range items {
		if it.AccountID == "" {
			t.Errorf("expected resolved AccountID for email %s, got empty", it.CustomerEmail)
		}
	}
}

func TestZendeskWebhook(t *testing.T) {
	s, resolver, cleanup := setupTestStore(t)
	defer cleanup()

	payload := `{
		"ticket": {
			"id": 48291,
			"subject": "Missing Okta SSO",
			"description": "Our team cannot use password logins anymore.",
			"requester": {
				"name": "Sarah Connor",
				"email": "sarah@cyberdyne.com"
			}
		}
	}`

	item, err := IngestZendeskPayload(context.Background(), s, resolver, []byte(payload))
	if err != nil {
		t.Fatalf("zendesk ingest error: %v", err)
	}
	if item.Source != models.SourceZendesk || item.AccountID != "acc_cyberdyne" {
		t.Errorf("unexpected item: %+v", item)
	}
}

func TestIntercomWebhook(t *testing.T) {
	s, resolver, cleanup := setupTestStore(t)
	defer cleanup()

	payload := `{
		"type": "conversation",
		"data": {
			"item": {
				"id": "conv_999",
				"conversation_message": {
					"body": "<p>We need <b>custom roles</b> for our clinical auditors.</p>",
					"author": {
						"name": "Albert Wesker",
						"email": "albert.wesker@umbrella.bio"
					}
				}
			}
		}
	}`

	item, err := IngestIntercomPayload(context.Background(), s, resolver, []byte(payload))
	if err != nil {
		t.Fatalf("intercom ingest error: %v", err)
	}
	if item.RawText != "We need custom roles for our clinical auditors." {
		t.Errorf("HTML stripping failed, got: %s", item.RawText)
	}
	if item.AccountID != "acc_umbrella" {
		t.Errorf("expected acc_umbrella, got %s", item.AccountID)
	}
}

func TestGongCallIngest(t *testing.T) {
	s, resolver, cleanup := setupTestStore(t)
	defer cleanup()

	payload := `{
		"call_id": "call_12345",
		"title": "Quarterly Enterprise Renewal Sync",
		"customer": {
			"name": "Pepper Potts",
			"email": "jarvis-admin@stark.tech",
			"company": "Stark Industries"
		},
		"transcripts": [
			{
				"speaker_name": "Sales Rep",
				"is_customer": false,
				"text": "How is your deployment going?"
			},
			{
				"speaker_name": "Pepper Potts",
				"is_customer": true,
				"text": "Our CISO has a critical compliance blocker: we need SIEM audit log exports for SOC 2 Type II."
			}
		]
	}`

	items, err := IngestGongCall(context.Background(), s, resolver, []byte(payload))
	if err != nil {
		t.Fatalf("gong ingest error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 extracted customer feedback item, got %d", len(items))
	}
	if items[0].AccountID != "acc_stark" {
		t.Errorf("expected acc_stark, got %s", items[0].AccountID)
	}
}

func TestSlackPayload(t *testing.T) {
	s, resolver, cleanup := setupTestStore(t)
	defer cleanup()

	payload := `{
		"type": "event_callback",
		"event": {
			"type": "message",
			"user": "U12345",
			"user_email": "richard.hendricks@piedpiper.com",
			"username": "Richard Hendricks",
			"text": "The REST API times out on 50k rows. We need a bulk export endpoint.",
			"channel": "C998877",
			"ts": "1710000000.1234"
		}
	}`

	item, err := IngestSlackPayload(context.Background(), s, resolver, []byte(payload))
	if err != nil {
		t.Fatalf("slack ingest error: %v", err)
	}
	if item.AccountID != "acc_piedpiper" || item.Source != models.SourceSlack {
		t.Errorf("unexpected slack item: %+v", item)
	}
}

func TestSeedRealisticFeedback(t *testing.T) {
	s, resolver, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	count, err := SeedRealisticFeedback(ctx, s, resolver)
	if err != nil {
		t.Fatalf("seed feedback error: %v", err)
	}
	if count < 10 {
		t.Fatalf("expected at least 10 seeded feedback items, got %d", count)
	}

	items, err := s.ListFeedbackItems(ctx, 100, 0)
	if err != nil {
		t.Fatalf("list items error: %v", err)
	}
	for _, it := range items {
		if it.AccountID == "" {
			t.Errorf("seeded feedback item %s has no linked account", it.ID)
		}
	}
}
