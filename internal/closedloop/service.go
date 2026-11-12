package closedloop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
	"github.com/google/uuid"
)

// Service manages closed-loop customer communication workflows.
type Service struct {
	store *store.Store
}

// NewService initializes a new closed loop service.
func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

// GenerateNotifications creates personalized update drafts for all accounts that requested a feature.
func (s *Service) GenerateNotifications(ctx context.Context, oppID string) ([]models.ClosedLoopNotification, error) {
	opp, err := s.store.GetOpportunityByID(ctx, oppID)
	if err != nil {
		return nil, fmt.Errorf("get opportunity: %w", err)
	}

	quotes, err := s.store.ListEnrichedFeedbackForOpportunity(ctx, oppID)
	if err != nil {
		return nil, fmt.Errorf("list feedback: %w", err)
	}

	// Deduplicate by recipient email
	seenEmails := make(map[string]bool)
	var generated []models.ClosedLoopNotification

	for _, q := range quotes {
		email := strings.TrimSpace(q.CustomerEmail)
		if email == "" || seenEmails[email] {
			continue
		}
		seenEmails[email] = true

		name := q.CustomerName
		if name == "" {
			name = "there"
		}

		companyName := q.AccountName
		if companyName == "" {
			companyName = q.AccountDomain
		}
		if companyName == "" {
			companyName = "your team"
		}

		subject := fmt.Sprintf("Now Live: %s", opp.Title)

		body := fmt.Sprintf(`Hi %s,

Thank you for your valuable product feedback. You previously reached out regarding %s:
"%s"

We are pleased to let you know that our engineering team has completed this capability and it is now available for %s.

What is new:
%s

You can enable and configure this feature immediately in your workspace settings under Security & Integrations.

If you have any questions or need setup assistance, reply directly to this email and our team will assist you.

Best regards,
The Product Team
`, name, opp.Title, strings.TrimSpace(q.RawText), companyName, opp.Description)

		notif := models.ClosedLoopNotification{
			ID:            fmt.Sprintf("notif_%s", uuid.New().String()[:8]),
			OpportunityID: oppID,
			AccountID:     q.AccountID,
			CustomerEmail: email,
			CustomerName:  name,
			Subject:       subject,
			Body:          body,
			Status:        "DRAFT",
			CreatedAt:     time.Now().UTC(),
		}

		if err := s.store.InsertNotification(ctx, &notif); err != nil {
			return nil, fmt.Errorf("insert notification: %w", err)
		}

		generated = append(generated, notif)
	}

	return generated, nil
}

// DispatchNotification sends or simulates sending a notification.
func (s *Service) DispatchNotification(ctx context.Context, notificationID string) error {
	now := time.Now().UTC()
	return s.store.UpdateNotificationStatus(ctx, notificationID, "SENT", &now)
}

// DispatchAllPending sends all draft notifications for an opportunity.
func (s *Service) DispatchAllPending(ctx context.Context, oppID string) (int, error) {
	notifs, err := s.store.ListNotificationsByOpportunity(ctx, oppID)
	if err != nil {
		return 0, err
	}

	sentCount := 0
	for _, n := range notifs {
		if n.Status == "DRAFT" || n.Status == "QUEUED" {
			if err := s.DispatchNotification(ctx, n.ID); err == nil {
				sentCount++
			}
		}
	}

	return sentCount, nil
}
