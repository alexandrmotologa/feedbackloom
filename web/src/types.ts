export interface Account {
  id: string
  name: string
  domain: string
  arr_usd: number
  plan_tier: 'ENTERPRISE' | 'GROWTH' | 'STARTER'
  renewal_date: string
  churn_risk_score: number
  created_at: string
}

export interface FeedbackItem {
  id: string
  source: string
  account_id?: string
  customer_email: string
  customer_name?: string
  raw_text: string
  extracted_pain_point?: string
  extracted_outcome?: string
  sentiment_score: number
  urgency_score: number
  created_at: string
}

export interface EnrichedFeedbackItem extends FeedbackItem {
  account_name?: string
  account_domain?: string
  account_arr_usd?: number
  plan_tier?: string
  renewal_date?: string
  days_to_renewal: number
  churn_risk_score: number
}

export interface Opportunity {
  id: string
  title: string
  description: string
  category: string
  total_arr_usd: number
  weighted_score: number
  account_count: number
  feedback_count: number
  status: 'DISCOVERED' | 'IN_PRD' | 'PLANNED' | 'IN_PROGRESS' | 'SHIPPED'
  generated_prd_markdown?: string
  generated_prd_notion_json?: string
  shipped_at?: string
  created_at: string
}

export interface ClosedLoopNotification {
  id: string
  opportunity_id: string
  account_id: string
  customer_email: string
  customer_name: string
  subject: string
  body: string
  status: 'DRAFT' | 'QUEUED' | 'SENT' | 'FAILED'
  sent_at?: string
  created_at: string
}

export interface DashboardStats {
  total_accounts: number
  portfolio_arr_usd: number
  total_feedback: number
  total_opportunities: number
  total_demanded_arr_usd: number
}
