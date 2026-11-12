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

// IntercomWebhookPayload models incoming Intercom conversation webhooks.
type IntercomWebhookPayload struct {
	Type string `json:"type"`
	Data struct {
		Item struct {
			ID                  string `json:"id"`
			ConversationMessage struct {
				Body   string `json:"body"`
				Author struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"author"`
			} `json:"conversation_message"`
			Source struct {
				Body   string `json:"body"`
				Author struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"author"`
			} `json:"source"`
		} `json:"item"`
	} `json:"data"`
}

// IngestIntercomPayload parses and persists an Intercom conversation webhook.
func IngestIntercomPayload(ctx context.Context, s *store.Store, resolver *crm.Resolver, body []byte) (*models.FeedbackItem, error) {
	var payload IntercomWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal intercom payload: %w", err)
	}

	rawText := strings.TrimSpace(payload.Data.Item.ConversationMessage.Body)
	authorName := strings.TrimSpace(payload.Data.Item.ConversationMessage.Author.Name)
	authorEmail := strings.TrimSpace(payload.Data.Item.ConversationMessage.Author.Email)

	if rawText == "" {
		rawText = strings.TrimSpace(payload.Data.Item.Source.Body)
	}
	if authorName == "" {
		authorName = strings.TrimSpace(payload.Data.Item.Source.Author.Name)
	}
	if authorEmail == "" {
		authorEmail = strings.TrimSpace(payload.Data.Item.Source.Author.Email)
	}

	if rawText == "" {
		return nil, fmt.Errorf("intercom message body is empty")
	}

	// Strip basic HTML tags if Intercom sent rich HTML
	rawText = stripHTML(rawText)

	accountID := ""
	if resolver != nil && authorEmail != "" {
		if acc, _ := resolver.ResolveAccount(ctx, authorEmail); acc != nil {
			accountID = acc.ID
		}
	}

	id := fmt.Sprintf("ic_%s", payload.Data.Item.ID)
	if payload.Data.Item.ID == "" {
		id = fmt.Sprintf("ic_%s", uuid.New().String()[:8])
	}

	item := &models.FeedbackItem{
		ID:            id,
		Source:        models.SourceIntercom,
		AccountID:     accountID,
		CustomerEmail: authorEmail,
		CustomerName:  authorName,
		RawText:       rawText,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.InsertFeedbackItem(ctx, item); err != nil {
		return nil, fmt.Errorf("persist intercom item: %w", err)
	}

	return item, nil
}

// stripHTML removes standard HTML tags from rich text message bodies.
func stripHTML(input string) string {
	var b strings.Builder
	inTag := false
	for _, r := range input {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
