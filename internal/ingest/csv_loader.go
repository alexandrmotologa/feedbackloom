package ingest

import (
	"context"
	"encoding/csv"
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

// ImportFeedbackFromCSV imports customer feedback records from a CSV file.
func ImportFeedbackFromCSV(ctx context.Context, s *store.Store, resolver *crm.Resolver, filePath string, defaultSource string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open feedback csv: %w", err)
	}
	defer file.Close()

	return ImportFeedbackFromReader(ctx, s, resolver, file, defaultSource)
}

// ImportFeedbackFromReader reads feedback CSV records, enriches with CRM accounts, and persists to store.
func ImportFeedbackFromReader(ctx context.Context, s *store.Store, resolver *crm.Resolver, r io.Reader, defaultSource string) (int, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}

	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	textIdx, ok := headerMap["raw_text"]
	if !ok {
		textIdx, ok = headerMap["text"]
	}
	if !ok {
		textIdx, ok = headerMap["feedback"]
	}
	if !ok {
		textIdx, ok = headerMap["message"]
	}
	if !ok {
		return 0, fmt.Errorf("missing text column in csv (raw_text, text, feedback, or message)")
	}

	emailIdx, hasEmail := headerMap["customer_email"]
	if !hasEmail {
		emailIdx, hasEmail = headerMap["email"]
	}
	nameIdx, hasName := headerMap["customer_name"]
	if !hasName {
		nameIdx, hasName = headerMap["name"]
	}
	sourceIdx, hasSource := headerMap["source"]
	createdIdx, hasCreated := headerMap["created_at"]
	idIdx, hasID := headerMap["id"]

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("read csv record: %w", err)
		}

		rawText := strings.TrimSpace(record[textIdx])
		if rawText == "" {
			continue
		}

		email := ""
		if hasEmail && emailIdx < len(record) {
			email = strings.TrimSpace(record[emailIdx])
		}

		name := ""
		if hasName && nameIdx < len(record) {
			name = strings.TrimSpace(record[nameIdx])
		}

		source := defaultSource
		if source == "" {
			source = models.SourceCSV
		}
		if hasSource && sourceIdx < len(record) && strings.TrimSpace(record[sourceIdx]) != "" {
			source = strings.ToUpper(strings.TrimSpace(record[sourceIdx]))
		}

		createdAt := time.Now().UTC()
		if hasCreated && createdIdx < len(record) && strings.TrimSpace(record[createdIdx]) != "" {
			val := strings.TrimSpace(record[createdIdx])
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				createdAt = t
			} else if t, err := time.Parse("2006-01-02", val); err == nil {
				createdAt = t
			}
		}

		id := fmt.Sprintf("fb_%s", uuid.New().String()[:8])
		if hasID && idIdx < len(record) && strings.TrimSpace(record[idIdx]) != "" {
			id = strings.TrimSpace(record[idIdx])
		}

		// Resolve account from customer email
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
			RawText:       rawText,
			CreatedAt:     createdAt,
		}

		if err := s.InsertFeedbackItem(ctx, item); err != nil {
			return count, fmt.Errorf("insert feedback: %w", err)
		}
		count++
	}

	return count, nil
}
