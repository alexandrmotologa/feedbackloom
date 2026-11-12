package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
	"github.com/google/uuid"
)

// JSONFeedbackPayload represents an incoming JSON record for customer feedback.
type JSONFeedbackPayload struct {
	ID            string `json:"id,omitempty"`
	Source        string `json:"source,omitempty"`
	CustomerEmail string `json:"customer_email,omitempty"`
	CustomerName  string `json:"customer_name,omitempty"`
	RawText       string `json:"raw_text"`
	CreatedAt     string `json:"created_at,omitempty"`
}

// ImportFeedbackFromJSON imports feedback items from a JSON array file.
func ImportFeedbackFromJSON(ctx context.Context, s *store.Store, resolver *crm.Resolver, filePath string, defaultSource string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open json file: %w", err)
	}
	defer file.Close()

	return ImportFeedbackFromJSONReader(ctx, s, resolver, file, defaultSource)
}

// ImportFeedbackFromJSONReader imports feedback items from an io.Reader.
func ImportFeedbackFromJSONReader(ctx context.Context, s *store.Store, resolver *crm.Resolver, r io.Reader, defaultSource string) (int, error) {
	var items []JSONFeedbackPayload
	if err := json.NewDecoder(r).Decode(&items); err != nil {
		return 0, fmt.Errorf("decode json payload: %w", err)
	}

	count := 0
	for _, raw := range items {
		text := strings.TrimSpace(raw.RawText)
		if text == "" {
			continue
		}

		source := defaultSource
		if raw.Source != "" {
			source = strings.ToUpper(raw.Source)
		}
		if source == "" {
			source = models.SourceAPI
		}

		email := strings.TrimSpace(raw.CustomerEmail)
		name := strings.TrimSpace(raw.CustomerName)

		createdAt := time.Now().UTC()
		if raw.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, raw.CreatedAt); err == nil {
				createdAt = t
			}
		}

		id := raw.ID
		if id == "" {
			id = fmt.Sprintf("fb_%s", uuid.New().String()[:8])
		}

		accountID := ""
		if resolver != nil && email != "" {
			if acc, _ := resolver.ResolveAccount(ctx, email); acc != nil {
				accountID = acc.ID
			}
		}

		item := &models.FeedbackItem{
			ID:            id,
			Source:        source,
			AccountID:     accountID,
			CustomerEmail: email,
			CustomerName:  name,
			RawText:       text,
			CreatedAt:     createdAt,
		}

		if err := s.InsertFeedbackItem(ctx, item); err != nil {
			return count, fmt.Errorf("insert feedback: %w", err)
		}
		count++
	}

	return count, nil
}
