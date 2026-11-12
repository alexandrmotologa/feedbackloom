package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	_ "modernc.org/sqlite"
)

// Store encapsulates database operations on SQLite.
type Store struct {
	db *sql.DB
}

// New opens a connection to SQLite, runs migrations, and configures WAL mode.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	// Optimize connection pooling for SQLite.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &Store{db: db}, nil
}

// Close terminates database connections.
func (s *Store) Close() error {
	return s.db.Close()
}

// UpsertAccount inserts or updates a customer account.
func (s *Store) UpsertAccount(ctx context.Context, acc *models.Account) error {
	query := `
		INSERT INTO accounts (id, name, domain, arr_usd, plan_tier, renewal_date, churn_risk_score, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(domain) DO UPDATE SET
			name = excluded.name,
			arr_usd = excluded.arr_usd,
			plan_tier = excluded.plan_tier,
			renewal_date = excluded.renewal_date,
			churn_risk_score = excluded.churn_risk_score,
			updated_at = CURRENT_TIMESTAMP;
	`
	now := time.Now().UTC()
	if acc.CreatedAt.IsZero() {
		acc.CreatedAt = now
	}
	acc.UpdatedAt = now

	renewalStr := ""
	if !acc.RenewalDate.IsZero() {
		renewalStr = acc.RenewalDate.Format("2006-01-02")
	}

	_, err := s.db.ExecContext(ctx, query,
		acc.ID,
		acc.Name,
		acc.Domain,
		acc.ARRUSD,
		acc.PlanTier,
		renewalStr,
		acc.ChurnRiskScore,
		acc.CreatedAt.Format(time.RFC3339),
		acc.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("upsert account: %w", err)
	}
	return nil
}

// GetAccountByID retrieves an account by ID.
func (s *Store) GetAccountByID(ctx context.Context, id string) (*models.Account, error) {
	query := `SELECT id, name, domain, arr_usd, plan_tier, renewal_date, churn_risk_score, created_at, updated_at FROM accounts WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var a models.Account
	var renewalStr, createdStr, updatedStr string

	err := row.Scan(&a.ID, &a.Name, &a.Domain, &a.ARRUSD, &a.PlanTier, &renewalStr, &a.ChurnRiskScore, &createdStr, &updatedStr)
	if err != nil {
		return nil, err
	}

	if renewalStr != "" {
		a.RenewalDate, _ = time.Parse("2006-01-02", renewalStr)
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	return &a, nil
}

// GetAccountByDomain retrieves an account by its unique email domain.
func (s *Store) GetAccountByDomain(ctx context.Context, domain string) (*models.Account, error) {
	query := `SELECT id, name, domain, arr_usd, plan_tier, renewal_date, churn_risk_score, created_at, updated_at FROM accounts WHERE domain = ?`
	row := s.db.QueryRowContext(ctx, query, domain)

	var a models.Account
	var renewalStr, createdStr, updatedStr string

	err := row.Scan(&a.ID, &a.Name, &a.Domain, &a.ARRUSD, &a.PlanTier, &renewalStr, &a.ChurnRiskScore, &createdStr, &updatedStr)
	if err != nil {
		return nil, err
	}

	if renewalStr != "" {
		a.RenewalDate, _ = time.Parse("2006-01-02", renewalStr)
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	return &a, nil
}

// ListAccounts returns all accounts sorted by ARR descending.
func (s *Store) ListAccounts(ctx context.Context) ([]models.Account, error) {
	query := `SELECT id, name, domain, arr_usd, plan_tier, renewal_date, churn_risk_score, created_at, updated_at FROM accounts ORDER BY arr_usd DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Account
	for rows.Next() {
		var a models.Account
		var renewalStr, createdStr, updatedStr string

		if err := rows.Scan(&a.ID, &a.Name, &a.Domain, &a.ARRUSD, &a.PlanTier, &renewalStr, &a.ChurnRiskScore, &createdStr, &updatedStr); err != nil {
			return nil, err
		}
		if renewalStr != "" {
			a.RenewalDate, _ = time.Parse("2006-01-02", renewalStr)
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		result = append(result, a)
	}
	return result, rows.Err()
}

// InsertFeedbackItem persists a new feedback item.
func (s *Store) InsertFeedbackItem(ctx context.Context, item *models.FeedbackItem) error {
	query := `
		INSERT INTO feedback_items (id, source, account_id, customer_email, customer_name, raw_text, extracted_pain_point, extracted_outcome, sentiment_score, urgency_score, competitor_mentions, embedding, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			account_id = excluded.account_id,
			customer_name = excluded.customer_name,
			extracted_pain_point = excluded.extracted_pain_point,
			extracted_outcome = excluded.extracted_outcome,
			sentiment_score = excluded.sentiment_score,
			urgency_score = excluded.urgency_score,
			competitor_mentions = excluded.competitor_mentions,
			embedding = excluded.embedding;
	`
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}

	var embBytes []byte
	if len(item.Embedding) > 0 {
		embBytes = EncodeFloat32Slice(item.Embedding)
	}

	var accountIDVal interface{}
	if item.AccountID != "" {
		accountIDVal = item.AccountID
	}

	compJSON, _ := json.Marshal(item.CompetitorMentions)

	_, err := s.db.ExecContext(ctx, query,
		item.ID,
		item.Source,
		accountIDVal,
		item.CustomerEmail,
		item.CustomerName,
		item.RawText,
		item.ExtractedPainPoint,
		item.ExtractedOutcome,
		item.SentimentScore,
		item.UrgencyScore,
		string(compJSON),
		embBytes,
		item.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert feedback item: %w", err)
	}
	return nil
}

// GetFeedbackItemByID fetches a single feedback entry.
func (s *Store) GetFeedbackItemByID(ctx context.Context, id string) (*models.FeedbackItem, error) {
	query := `SELECT id, source, account_id, customer_email, customer_name, raw_text, extracted_pain_point, extracted_outcome, sentiment_score, urgency_score, competitor_mentions, embedding, created_at FROM feedback_items WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var item models.FeedbackItem
	var accountID, custName, painPoint, outcome, compStr sql.NullString
	var embBytes []byte
	var createdStr string

	err := row.Scan(&item.ID, &item.Source, &accountID, &item.CustomerEmail, &custName, &item.RawText, &painPoint, &outcome, &item.SentimentScore, &item.UrgencyScore, &compStr, &embBytes, &createdStr)
	if err != nil {
		return nil, err
	}

	if accountID.Valid {
		item.AccountID = accountID.String
	}
	if custName.Valid {
		item.CustomerName = custName.String
	}
	if painPoint.Valid {
		item.ExtractedPainPoint = painPoint.String
	}
	if outcome.Valid {
		item.ExtractedOutcome = outcome.String
	}
	if compStr.Valid && compStr.String != "" {
		_ = json.Unmarshal([]byte(compStr.String), &item.CompetitorMentions)
	}
	if len(embBytes) > 0 {
		item.Embedding = DecodeFloat32Slice(embBytes)
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)

	return &item, nil
}

// ListFeedbackItems returns a paginated list of feedback items.
func (s *Store) ListFeedbackItems(ctx context.Context, limit, offset int) ([]models.FeedbackItem, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, source, account_id, customer_email, customer_name, raw_text, extracted_pain_point, extracted_outcome, sentiment_score, urgency_score, competitor_mentions, embedding, created_at
		FROM feedback_items
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.FeedbackItem
	for rows.Next() {
		var item models.FeedbackItem
		var accountID, custName, painPoint, outcome, compStr sql.NullString
		var embBytes []byte
		var createdStr string

		if err := rows.Scan(&item.ID, &item.Source, &accountID, &item.CustomerEmail, &custName, &item.RawText, &painPoint, &outcome, &item.SentimentScore, &item.UrgencyScore, &compStr, &embBytes, &createdStr); err != nil {
			return nil, err
		}
		if accountID.Valid {
			item.AccountID = accountID.String
		}
		if custName.Valid {
			item.CustomerName = custName.String
		}
		if painPoint.Valid {
			item.ExtractedPainPoint = painPoint.String
		}
		if outcome.Valid {
			item.ExtractedOutcome = outcome.String
		}
		if compStr.Valid && compStr.String != "" {
			_ = json.Unmarshal([]byte(compStr.String), &item.CompetitorMentions)
		}
		if len(embBytes) > 0 {
			item.Embedding = DecodeFloat32Slice(embBytes)
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		result = append(result, item)
	}
	return result, rows.Err()
}

// ListAllFeedbackItems returns all feedback items.
func (s *Store) ListAllFeedbackItems(ctx context.Context) ([]models.FeedbackItem, error) {
	return s.ListFeedbackItems(ctx, 100000, 0)
}

// UpdateFeedbackExtraction updates NLP extraction fields.
func (s *Store) UpdateFeedbackExtraction(ctx context.Context, id string, painPoint, outcome string, sentiment, urgency float64, competitors []string) error {
	compJSON, _ := json.Marshal(competitors)
	query := `UPDATE feedback_items SET extracted_pain_point = ?, extracted_outcome = ?, sentiment_score = ?, urgency_score = ?, competitor_mentions = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, painPoint, outcome, sentiment, urgency, string(compJSON), id)
	return err
}

// UpdateFeedbackEmbedding updates vector embedding bytes.
func (s *Store) UpdateFeedbackEmbedding(ctx context.Context, id string, embedding []float32) error {
	query := `UPDATE feedback_items SET embedding = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, EncodeFloat32Slice(embedding), id)
	return err
}

// UpdateFeedbackAccount links a feedback item to a customer account.
func (s *Store) UpdateFeedbackAccount(ctx context.Context, feedbackID, accountID string) error {
	query := `UPDATE feedback_items SET account_id = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, accountID, feedbackID)
	return err
}

// UpsertOpportunity creates or updates a product opportunity.
func (s *Store) UpsertOpportunity(ctx context.Context, opp *models.Opportunity) error {
	query := `
		INSERT INTO opportunities (id, title, description, category, total_arr_usd, weighted_score, account_count, feedback_count, status, competitor_mentions, lost_pipeline_arr, effort_sprints, roi_index, urgent_renewal_alert, velocity_growth_pct, generated_prd_markdown, generated_prd_notion_json, shipped_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			category = excluded.category,
			total_arr_usd = excluded.total_arr_usd,
			weighted_score = excluded.weighted_score,
			account_count = excluded.account_count,
			feedback_count = excluded.feedback_count,
			status = excluded.status,
			competitor_mentions = excluded.competitor_mentions,
			lost_pipeline_arr = excluded.lost_pipeline_arr,
			effort_sprints = excluded.effort_sprints,
			roi_index = excluded.roi_index,
			urgent_renewal_alert = excluded.urgent_renewal_alert,
			velocity_growth_pct = excluded.velocity_growth_pct,
			generated_prd_markdown = excluded.generated_prd_markdown,
			generated_prd_notion_json = excluded.generated_prd_notion_json,
			shipped_at = excluded.shipped_at,
			updated_at = CURRENT_TIMESTAMP;
	`
	now := time.Now().UTC()
	if opp.CreatedAt.IsZero() {
		opp.CreatedAt = now
	}
	opp.UpdatedAt = now

	var shippedStr interface{}
	if opp.ShippedAt != nil {
		shippedStr = opp.ShippedAt.Format(time.RFC3339)
	}

	compJSON, _ := json.Marshal(opp.CompetitorMentions)
	urgentVal := 0
	if opp.UrgentRenewalAlert {
		urgentVal = 1
	}

	if opp.EffortSprints <= 0 {
		opp.EffortSprints = 2
	}
	if opp.ROIIndex <= 0 {
		opp.ROIIndex = opp.TotalARRUSD / float64(opp.EffortSprints)
	}

	_, err := s.db.ExecContext(ctx, query,
		opp.ID,
		opp.Title,
		opp.Description,
		opp.Category,
		opp.TotalARRUSD,
		opp.WeightedScore,
		opp.AccountCount,
		opp.FeedbackCount,
		opp.Status,
		string(compJSON),
		opp.LostPipelineARR,
		opp.EffortSprints,
		opp.ROIIndex,
		urgentVal,
		opp.VelocityGrowthPct,
		opp.GeneratedPRDMarkdown,
		opp.GeneratedPRDNotionJSON,
		shippedStr,
		opp.CreatedAt.Format(time.RFC3339),
		opp.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetOpportunityByID fetches an opportunity by ID.
func (s *Store) GetOpportunityByID(ctx context.Context, id string) (*models.Opportunity, error) {
	query := `SELECT id, title, description, category, total_arr_usd, weighted_score, account_count, feedback_count, status, competitor_mentions, lost_pipeline_arr, effort_sprints, roi_index, urgent_renewal_alert, velocity_growth_pct, generated_prd_markdown, generated_prd_notion_json, shipped_at, created_at, updated_at FROM opportunities WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var opp models.Opportunity
	var prdMD, prdNotion, shippedStr, compStr sql.NullString
	var urgentVal int
	var createdStr, updatedStr string

	err := row.Scan(
		&opp.ID, &opp.Title, &opp.Description, &opp.Category, &opp.TotalARRUSD, &opp.WeightedScore,
		&opp.AccountCount, &opp.FeedbackCount, &opp.Status, &compStr, &opp.LostPipelineARR,
		&opp.EffortSprints, &opp.ROIIndex, &urgentVal, &opp.VelocityGrowthPct,
		&prdMD, &prdNotion, &shippedStr, &createdStr, &updatedStr,
	)
	if err != nil {
		return nil, err
	}

	opp.UrgentRenewalAlert = (urgentVal == 1)
	if compStr.Valid && compStr.String != "" {
		_ = json.Unmarshal([]byte(compStr.String), &opp.CompetitorMentions)
	}
	if prdMD.Valid {
		opp.GeneratedPRDMarkdown = prdMD.String
	}
	if prdNotion.Valid {
		opp.GeneratedPRDNotionJSON = prdNotion.String
	}
	if shippedStr.Valid {
		t, _ := time.Parse(time.RFC3339, shippedStr.String)
		opp.ShippedAt = &t
	}
	opp.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	opp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)

	return &opp, nil
}

// ListOpportunities returns all opportunities ordered by Weighted Revenue Score (WRS) descending.
func (s *Store) ListOpportunities(ctx context.Context) ([]models.Opportunity, error) {
	query := `
		SELECT id, title, description, category, total_arr_usd, weighted_score, account_count, feedback_count, status, competitor_mentions, lost_pipeline_arr, effort_sprints, roi_index, urgent_renewal_alert, velocity_growth_pct, generated_prd_markdown, generated_prd_notion_json, shipped_at, created_at, updated_at
		FROM opportunities
		ORDER BY weighted_score DESC, total_arr_usd DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Opportunity
	for rows.Next() {
		var opp models.Opportunity
		var prdMD, prdNotion, shippedStr, compStr sql.NullString
		var urgentVal int
		var createdStr, updatedStr string

		if err := rows.Scan(
			&opp.ID, &opp.Title, &opp.Description, &opp.Category, &opp.TotalARRUSD, &opp.WeightedScore,
			&opp.AccountCount, &opp.FeedbackCount, &opp.Status, &compStr, &opp.LostPipelineARR,
			&opp.EffortSprints, &opp.ROIIndex, &urgentVal, &opp.VelocityGrowthPct,
			&prdMD, &prdNotion, &shippedStr, &createdStr, &updatedStr,
		); err != nil {
			return nil, err
		}

		opp.UrgentRenewalAlert = (urgentVal == 1)
		if compStr.Valid && compStr.String != "" {
			_ = json.Unmarshal([]byte(compStr.String), &opp.CompetitorMentions)
		}
		if prdMD.Valid {
			opp.GeneratedPRDMarkdown = prdMD.String
		}
		if prdNotion.Valid {
			opp.GeneratedPRDNotionJSON = prdNotion.String
		}
		if shippedStr.Valid {
			t, _ := time.Parse(time.RFC3339, shippedStr.String)
			opp.ShippedAt = &t
		}
		opp.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		opp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
		result = append(result, opp)
	}
	return result, rows.Err()
}

// UpdateOpportunityEffort updates the estimated engineering sprints and recalculates ROI index.
func (s *Store) UpdateOpportunityEffort(ctx context.Context, id string, sprints int) error {
	if sprints <= 0 {
		sprints = 1
	}
	query := `
		UPDATE opportunities
		SET effort_sprints = ?,
		    roi_index = total_arr_usd / ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, sprints, sprints, id)
	return err
}

// DeleteOpportunity removes an opportunity and its feedback links.
func (s *Store) DeleteOpportunity(ctx context.Context, id string) error {
	query := `DELETE FROM opportunities WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// UpdateOpportunityStatus updates the lifecycle status.
func (s *Store) UpdateOpportunityStatus(ctx context.Context, id string, status string) error {
	var shippedVal interface{}
	if status == models.StatusShipped {
		shippedVal = time.Now().UTC().Format(time.RFC3339)
	}
	query := `UPDATE opportunities SET status = ?, shipped_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, status, shippedVal, id)
	return err
}

// UpdateOpportunityPRD stores the synthesized PRD.
func (s *Store) UpdateOpportunityPRD(ctx context.Context, id string, md, notionJSON string) error {
	query := `UPDATE opportunities SET generated_prd_markdown = ?, generated_prd_notion_json = ?, status = CASE WHEN status = 'DISCOVERED' THEN 'IN_PRD' ELSE status END, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, md, notionJSON, id)
	return err
}

// LinkFeedbackOpportunity creates a relationship mapping.
func (s *Store) LinkFeedbackOpportunity(ctx context.Context, feedbackID, opportunityID string, relevance float64) error {
	query := `
		INSERT INTO feedback_opportunity_map (feedback_id, opportunity_id, relevance_score)
		VALUES (?, ?, ?)
		ON CONFLICT(feedback_id, opportunity_id) DO UPDATE SET relevance_score = excluded.relevance_score
	`
	_, err := s.db.ExecContext(ctx, query, feedbackID, opportunityID, relevance)
	return err
}

// ClearOpportunityFeedbackMaps deletes all mappings for a given opportunity.
func (s *Store) ClearOpportunityFeedbackMaps(ctx context.Context, opportunityID string) error {
	query := `DELETE FROM feedback_opportunity_map WHERE opportunity_id = ?`
	_, err := s.db.ExecContext(ctx, query, opportunityID)
	return err
}

// ClearAllOpportunities wipes opportunities and mappings for re-clustering.
func (s *Store) ClearAllOpportunities(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM feedback_opportunity_map`); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM opportunities`);
	return err
}

// ListEnrichedFeedbackForOpportunity returns all feedback entries linked to an opportunity with joined account metrics.
func (s *Store) ListEnrichedFeedbackForOpportunity(ctx context.Context, opportunityID string) ([]models.EnrichedFeedbackItem, error) {
	query := `
		SELECT 
			f.id, f.source, f.account_id, f.customer_email, f.customer_name, f.raw_text,
			f.extracted_pain_point, f.extracted_outcome, f.sentiment_score, f.urgency_score, f.competitor_mentions, f.embedding, f.created_at,
			COALESCE(a.name, ''), COALESCE(a.domain, ''), COALESCE(a.arr_usd, 0.0),
			COALESCE(a.plan_tier, 'STARTER'), COALESCE(a.renewal_date, ''), COALESCE(a.churn_risk_score, 0.2)
		FROM feedback_opportunity_map m
		JOIN feedback_items f ON m.feedback_id = f.id
		LEFT JOIN accounts a ON f.account_id = a.id
		WHERE m.opportunity_id = ?
		ORDER BY COALESCE(a.arr_usd, 0.0) DESC, f.urgency_score DESC
	`
	rows, err := s.db.QueryContext(ctx, query, opportunityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.EnrichedFeedbackItem
	for rows.Next() {
		var item models.EnrichedFeedbackItem
		var custName, painPoint, outcome, compStr sql.NullString
		var embBytes []byte
		var createdStr, renewalStr string

		err := rows.Scan(
			&item.ID, &item.Source, &item.AccountID, &item.CustomerEmail, &custName, &item.RawText,
			&item.ExtractedPainPoint, &item.ExtractedOutcome, &item.SentimentScore, &item.UrgencyScore,
			&compStr, &embBytes, &createdStr,
			&item.AccountName, &item.AccountDomain, &item.AccountARRUSD,
			&item.PlanTier, &renewalStr, &item.ChurnRiskScore,
		)
		if err != nil {
			return nil, err
		}

		if custName.Valid {
			item.CustomerName = custName.String
		}
		if painPoint.Valid {
			item.ExtractedPainPoint = painPoint.String
		}
		if outcome.Valid {
			item.ExtractedOutcome = outcome.String
		}
		if compStr.Valid && compStr.String != "" {
			_ = json.Unmarshal([]byte(compStr.String), &item.CompetitorMentions)
		}
		if len(embBytes) > 0 {
			item.Embedding = DecodeFloat32Slice(embBytes)
		}
		item.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		item.RenewalDate = renewalStr

		if renewalStr != "" {
			t, err := time.Parse("2006-01-02", renewalStr)
			if err == nil {
				diff := time.Until(t)
				days := int(diff.Hours() / 24)
				if days < 0 {
					days = 0
				}
				item.DaysToRenewal = days
			} else {
				item.DaysToRenewal = 365
			}
		} else {
			item.DaysToRenewal = 365
		}

		result = append(result, item)
	}
	return result, rows.Err()
}

// InsertNotification stores a closed-loop notification.
func (s *Store) InsertNotification(ctx context.Context, n *models.ClosedLoopNotification) error {
	query := `
		INSERT INTO closed_loop_notifications (id, opportunity_id, account_id, customer_email, customer_name, subject, body, status, sent_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().UTC()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}

	var sentStr interface{}
	if n.SentAt != nil {
		sentStr = n.SentAt.Format(time.RFC3339)
	}

	var accID interface{}
	if n.AccountID != "" {
		accID = n.AccountID
	}

	_, err := s.db.ExecContext(ctx, query,
		n.ID, n.OpportunityID, accID, n.CustomerEmail, n.CustomerName, n.Subject, n.Body, n.Status, sentStr, n.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// ListNotificationsByOpportunity returns generated notifications for an opportunity.
func (s *Store) ListNotificationsByOpportunity(ctx context.Context, opportunityID string) ([]models.ClosedLoopNotification, error) {
	query := `SELECT id, opportunity_id, account_id, customer_email, customer_name, subject, body, status, sent_at, created_at FROM closed_loop_notifications WHERE opportunity_id = ? ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, opportunityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ClosedLoopNotification
	for rows.Next() {
		var n models.ClosedLoopNotification
		var accID, custName, sentStr sql.NullString
		var createdStr string

		if err := rows.Scan(&n.ID, &n.OpportunityID, &accID, &n.CustomerEmail, &custName, &n.Subject, &n.Body, &n.Status, &sentStr, &createdStr); err != nil {
			return nil, err
		}
		if accID.Valid {
			n.AccountID = accID.String
		}
		if custName.Valid {
			n.CustomerName = custName.String
		}
		if sentStr.Valid {
			t, _ := time.Parse(time.RFC3339, sentStr.String)
			n.SentAt = &t
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		result = append(result, n)
	}
	return result, rows.Err()
}

// UpdateNotificationStatus updates notification dispatch status.
func (s *Store) UpdateNotificationStatus(ctx context.Context, id string, status string, sentAt *time.Time) error {
	var sentStr interface{}
	if sentAt != nil {
		sentStr = sentAt.Format(time.RFC3339)
	}
	query := `UPDATE closed_loop_notifications SET status = ?, sent_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, status, sentStr, id)
	return err
}

// GetDashboardStats aggregates top-level portfolio revenue, competitor loss, and feedback metrics.
func (s *Store) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalFeedback int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM feedback_items`).Scan(&totalFeedback)
	if err != nil {
		return nil, err
	}
	stats["total_feedback"] = totalFeedback

	var totalAccounts int
	var totalPortfolioARR float64
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(arr_usd), 0.0) FROM accounts`).Scan(&totalAccounts, &totalPortfolioARR)
	if err != nil {
		return nil, err
	}
	stats["total_accounts"] = totalAccounts
	stats["portfolio_arr_usd"] = totalPortfolioARR

	var totalOpportunities int
	var demandedARR float64
	var lostPipelineARR float64
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(total_arr_usd), 0.0), COALESCE(SUM(lost_pipeline_arr), 0.0) FROM opportunities`).Scan(&totalOpportunities, &demandedARR, &lostPipelineARR)
	if err != nil {
		return nil, err
	}
	stats["total_opportunities"] = totalOpportunities
	stats["total_demanded_arr_usd"] = demandedARR
	stats["total_lost_pipeline_arr_usd"] = lostPipelineARR

	return stats, nil
}
