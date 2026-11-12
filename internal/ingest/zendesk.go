package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
	"github.com/google/uuid"
)

// ZendeskWebhookPayload models the incoming webhook structure from Zendesk Support.
type ZendeskWebhookPayload struct {
	Ticket struct {
		ID          int64  `json:"id"`
		Subject     string `json:"subject"`
		Description string `json:"description"`
		Requester   struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"requester"`
		CreatedAt string `json:"created_at"`
	} `json:"ticket"`
}

// IngestZendeskPayload parses and persists a Zendesk webhook payload.
func IngestZendeskPayload(ctx context.Context, s *store.Store, resolver *crm.Resolver, body []byte) (*models.FeedbackItem, error) {
	var payload ZendeskWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal zendesk payload: %w", err)
	}

	rawText := strings.TrimSpace(payload.Ticket.Description)
	if rawText == "" {
		rawText = strings.TrimSpace(payload.Ticket.Subject)
	}
	if rawText == "" {
		return nil, fmt.Errorf("zendesk ticket contains no subject or description")
	}

	email := strings.TrimSpace(payload.Ticket.Requester.Email)
	name := strings.TrimSpace(payload.Ticket.Requester.Name)

	accountID := ""
	if resolver != nil && email != "" {
		if acc, _ := resolver.ResolveAccount(ctx, email); acc != nil {
			accountID = acc.ID
		}
	}

	id := fmt.Sprintf("zd_%d", payload.Ticket.ID)
	if payload.Ticket.ID == 0 {
		id = fmt.Sprintf("zd_%s", uuid.New().String()[:8])
	}

	item := &models.FeedbackItem{
		ID:            id,
		Source:        models.SourceZendesk,
		AccountID:     accountID,
		CustomerEmail: email,
		CustomerName:  name,
		RawText:       rawText,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.InsertFeedbackItem(ctx, item); err != nil {
		return nil, fmt.Errorf("persist zendesk item: %w", err)
	}

	return item, nil
}
