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

// GongCallPayload models Gong call transcript exports and webhook events.
type GongCallPayload struct {
	CallID    string `json:"call_id"`
	Title     string `json:"title"`
	Customer  struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Company string `json:"company"`
	} `json:"customer"`
	Transcripts []struct {
		SpeakerName string `json:"speaker_name"`
		IsCustomer  bool   `json:"is_customer"`
		Text        string `json:"text"`
	} `json:"transcripts"`
}

// IngestGongCall processes a Gong call transcript, extracting customer statements and objections.
func IngestGongCall(ctx context.Context, s *store.Store, resolver *crm.Resolver, body []byte) ([]models.FeedbackItem, error) {
	var payload GongCallPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal gong payload: %w", err)
	}

	email := strings.TrimSpace(payload.Customer.Email)
	name := strings.TrimSpace(payload.Customer.Name)
	if name == "" && payload.Customer.Company != "" {
		name = payload.Customer.Company
	}

	accountID := ""
	if resolver != nil && email != "" {
		if acc, _ := resolver.ResolveAccount(ctx, email); acc != nil {
			accountID = acc.ID
		}
	}

	var results []models.FeedbackItem

	// If individual transcripts exist, extract customer remarks
	if len(payload.Transcripts) > 0 {
		for i, seg := range payload.Transcripts {
			text := strings.TrimSpace(seg.Text)
			if text == "" {
				continue
			}

			// Filter to customer speakers or statements with feature indicators
			isCustomer := seg.IsCustomer || !strings.Contains(strings.ToLower(seg.SpeakerName), "sales")
			if !isCustomer && !containsFeatureIndicator(text) {
				continue
			}

			speaker := seg.SpeakerName
			if speaker == "" {
				speaker = name
			}

			id := fmt.Sprintf("gong_%s_%d", payload.CallID, i)
			item := models.FeedbackItem{
				ID:            id,
				Source:        models.SourceGong,
				AccountID:     accountID,
				CustomerEmail: email,
				CustomerName:  speaker,
				RawText:       text,
				CreatedAt:     time.Now().UTC(),
			}

			if err := s.InsertFeedbackItem(ctx, &item); err != nil {
				return results, fmt.Errorf("persist gong item: %w", err)
			}
			results = append(results, item)
		}
	} else if payload.Title != "" {
		// Single high-level call note
		id := fmt.Sprintf("gong_%s", payload.CallID)
		if payload.CallID == "" {
			id = fmt.Sprintf("gong_%s", uuid.New().String()[:8])
		}
		item := models.FeedbackItem{
			ID:            id,
			Source:        models.SourceGong,
			AccountID:     accountID,
			CustomerEmail: email,
			CustomerName:  name,
			RawText:       payload.Title,
			CreatedAt:     time.Now().UTC(),
		}
		if err := s.InsertFeedbackItem(ctx, &item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// containsFeatureIndicator checks if a call segment expresses a requirement or gap.
func containsFeatureIndicator(text string) bool {
	lower := strings.ToLower(text)
	indicators := []string{
		"need", "require", "must have", "dealbreaker", "blocker", "missing",
		"when will you support", "can we get", "roadmap", "security requirement",
		"compliance", "sso", "audit log", "rbac", "webhook", "export",
	}
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}
	return false
}
