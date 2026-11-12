package models

import "time"

// Plan tiers for customer accounts.
const (
	PlanTierEnterprise = "ENTERPRISE"
	PlanTierGrowth     = "GROWTH"
	PlanTierStarter    = "STARTER"
)

// Feedback sources.
const (
	SourceZendesk  = "ZENDESK"
	SourceIntercom = "INTERCOM"
	SourceGong     = "GONG"
	SourceSlack    = "SLACK"
	SourceCSV      = "CSV"
	SourceAPI      = "API"
)

// Opportunity statuses.
const (
	StatusDiscovered = "DISCOVERED"
	StatusInPRD      = "IN_PRD"
	StatusPlanned    = "PLANNED"
	StatusInProgress = "IN_PROGRESS"
	StatusShipped    = "SHIPPED"
)

// Categories for opportunities.
const (
	CategorySecurity       = "SECURITY"
	CategoryIntegrations   = "INTEGRATIONS"
	CategoryPerformance    = "PERFORMANCE"
	CategoryUsability      = "USABILITY"
	CategoryInfrastructure = "INFRASTRUCTURE"
	CategoryAnalytics      = "ANALYTICS"
)

// Account represents a customer account enriched from CRM or billing providers.
type Account struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Domain         string    `json:"domain"`
	ARRUSD         float64   `json:"arr_usd"`
	PlanTier       string    `json:"plan_tier"`
	RenewalDate    time.Time `json:"renewal_date"`
	ChurnRiskScore float64   `json:"churn_risk_score"` // 0.0 (safe) to 1.0 (imminent churn)
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DaysToRenewal returns the number of days until account contract renewal.
func (a *Account) DaysToRenewal() int {
	if a.RenewalDate.IsZero() {
		return 365
	}
	diff := time.Until(a.RenewalDate)
	days := int(diff.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// FeedbackItem represents raw and extracted customer feedback from any supported channel.
type FeedbackItem struct {
	ID                  string    `json:"id"`
	Source              string    `json:"source"`
	AccountID           string    `json:"account_id,omitempty"`
	CustomerEmail       string    `json:"customer_email"`
	CustomerName        string    `json:"customer_name,omitempty"`
	RawText             string    `json:"raw_text"`
	ExtractedPainPoint  string    `json:"extracted_pain_point,omitempty"`
	ExtractedOutcome    string    `json:"extracted_outcome,omitempty"`
	SentimentScore      float64   `json:"sentiment_score"` // -1.0 to 1.0
	UrgencyScore        float64   `json:"urgency_score"`   // 0.0 to 1.0
	CompetitorMentions  []string  `json:"competitor_mentions,omitempty"`
	Embedding           []float32 `json:"embedding,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// Opportunity represents an ARR-ranked product theme synthesized from clustered feedback.
type Opportunity struct {
	ID                     string     `json:"id"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	Category               string     `json:"category"`
	TotalARRUSD            float64    `json:"total_arr_usd"`
	WeightedScore          float64    `json:"weighted_score"`
	AccountCount           int        `json:"account_count"`
	FeedbackCount          int        `json:"feedback_count"`
	Status                 string     `json:"status"`
	CompetitorMentions     []string   `json:"competitor_mentions,omitempty"`
	LostPipelineARR        float64    `json:"lost_pipeline_arr_usd"`
	EffortSprints          int        `json:"effort_sprints"`
	ROIIndex               float64    `json:"roi_index"`
	UrgentRenewalAlert     bool       `json:"urgent_renewal_alert"`
	VelocityGrowthPct      float64    `json:"velocity_growth_pct"`
	GeneratedPRDMarkdown   string     `json:"generated_prd_markdown,omitempty"`
	GeneratedPRDNotionJSON string     `json:"generated_prd_notion_json,omitempty"`
	ShippedAt              *time.Time `json:"shipped_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// FeedbackOpportunityMap links individual feedback items to synthesized opportunities.
type FeedbackOpportunityMap struct {
	FeedbackID     string  `json:"feedback_id"`
	OpportunityID  string  `json:"opportunity_id"`
	RelevanceScore float64 `json:"relevance_score"`
}

// EnrichedFeedbackItem joins customer feedback with verified account revenue data.
type EnrichedFeedbackItem struct {
	FeedbackItem
	AccountName    string  `json:"account_name,omitempty"`
	AccountDomain  string  `json:"account_domain,omitempty"`
	AccountARRUSD  float64 `json:"account_arr_usd,omitempty"`
	PlanTier       string  `json:"plan_tier,omitempty"`
	RenewalDate    string  `json:"renewal_date,omitempty"`
	DaysToRenewal  int     `json:"days_to_renewal"`
	ChurnRiskScore float64 `json:"churn_risk_score"`
}

// ClosedLoopNotification represents an automated customer update regarding a feature they requested.
type ClosedLoopNotification struct {
	ID             string     `json:"id"`
	OpportunityID  string     `json:"opportunity_id"`
	AccountID      string     `json:"account_id"`
	CustomerEmail  string     `json:"customer_email"`
	CustomerName   string     `json:"customer_name"`
	Subject        string     `json:"subject"`
	Body           string     `json:"body"`
	Status         string     `json:"status"` // DRAFT, QUEUED, SENT, FAILED
	SentAt         *time.Time `json:"sent_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// OpportunitySummary provides an aggregate overview for dashboards and rankings.
type OpportunitySummary struct {
	Opportunity
	EnterpriseARRUSD float64  `json:"enterprise_arr_usd"`
	GrowthARRUSD     float64  `json:"growth_arr_usd"`
	StarterARRUSD    float64  `json:"starter_arr_usd"`
	AvgChurnRisk     float64  `json:"avg_churn_risk"`
	MinDaysToRenewal int      `json:"min_days_to_renewal"`
	TopAccounts      []string `json:"top_accounts"`
}

// LiveIngestRequest represents an interactive sandbox test payload.
type LiveIngestRequest struct {
	Source        string `json:"source"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
	RawText       string `json:"raw_text"`
	Commit        bool   `json:"commit"` // If true, persists into database; else runs in simulation mode
}

// LiveIngestResponse represents the real-time simulation output from the sandbox.
type LiveIngestResponse struct {
	Account                 *Account `json:"account,omitempty"`
	ExtractedPainPoint      string   `json:"extracted_pain_point"`
	ExtractedOutcome        string   `json:"extracted_outcome"`
	Category                string   `json:"category"`
	UrgencyScore            float64  `json:"urgency_score"`
	SentimentScore          float64  `json:"sentiment_score"`
	CompetitorMentions      []string `json:"competitor_mentions"`
	MatchedOpportunityID    string   `json:"matched_opportunity_id,omitempty"`
	MatchedOpportunityTitle string   `json:"matched_opportunity_title,omitempty"`
	SimilarityScore         float64  `json:"similarity_score"`
	CreatedItemID           string   `json:"created_item_id,omitempty"`
}

// GitHubIssuePayload represents an exported issue structure.
type GitHubIssuePayload struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

// JiraEpicPayload represents an exported Jira Epic structure.
type JiraEpicPayload struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	IssueType   string `json:"issue_type"`
	Priority    string `json:"priority"`
}
