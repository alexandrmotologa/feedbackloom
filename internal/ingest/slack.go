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

// SlackEventPayload represents standard Slack event and webhook structures.
type SlackEventPayload struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge,omitempty"` // for Slack URL verification
	Event     struct {
		Type      string `json:"type"`
		User      string `json:"user"`
		UserEmail string `json:"user_email,omitempty"`
		UserName  string `json:"username,omitempty"`
		Text      string `json:"text"`
		Channel   string `json:"channel"`
		Ts        string `json:"ts"`
	} `json:"event"`
	// Also support direct slash command / incoming webhook payloads
	Text      string `json:"text,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
}

// IngestSlackPayload parses and persists Slack message feedback.
func IngestSlackPayload(ctx context.Context, s *store.Store, resolver *crm.Resolver, body []byte) (*models.FeedbackItem, error) {
	var payload SlackEventPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal slack payload: %w", err)
	}

	text := strings.TrimSpace(payload.Event.Text)
	if text == "" {
		text = strings.TrimSpace(payload.Text)
	}
	if text == "" {
		return nil, fmt.Errorf("slack message text is empty")
	}

	email := strings.TrimSpace(payload.Event.UserEmail)
	if email == "" {
		email = strings.TrimSpace(payload.UserEmail)
	}

	name := strings.TrimSpace(payload.Event.UserName)
	if name == "" {
		name = strings.TrimSpace(payload.UserName)
	}

	accountID := ""
	if resolver != nil && email != "" {
		if acc, _ := resolver.ResolveAccount(ctx, email); acc != nil {
			accountID = acc.ID
		}
	}

	id := fmt.Sprintf("slack_%s", uuid.New().String()[:8])
	if payload.Event.Ts != "" {
		cleanTs := strings.ReplaceAll(payload.Event.Ts, ".", "")
		id = fmt.Sprintf("slack_%s", cleanTs)
	}

	item := &models.FeedbackItem{
		ID:            id,
		Source:        models.SourceSlack,
		AccountID:     accountID,
		CustomerEmail: email,
		CustomerName:  name,
		RawText:       text,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.InsertFeedbackItem(ctx, item); err != nil {
		return nil, fmt.Errorf("persist slack item: %w", err)
	}

	return item, nil
}
