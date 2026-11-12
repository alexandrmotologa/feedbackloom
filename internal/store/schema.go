package store

const schemaSQL = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT UNIQUE NOT NULL,
    arr_usd REAL NOT NULL DEFAULT 0.0,
    plan_tier TEXT NOT NULL DEFAULT 'GROWTH',
    renewal_date TEXT,
    churn_risk_score REAL DEFAULT 0.2,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_accounts_domain ON accounts(domain);

CREATE TABLE IF NOT EXISTS feedback_items (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    account_id TEXT REFERENCES accounts(id) ON DELETE SET NULL,
    customer_email TEXT,
    customer_name TEXT,
    raw_text TEXT NOT NULL,
    extracted_pain_point TEXT,
    extracted_outcome TEXT,
    sentiment_score REAL DEFAULT 0.0,
    urgency_score REAL DEFAULT 0.0,
    embedding BLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_feedback_account ON feedback_items(account_id);
CREATE INDEX IF NOT EXISTS idx_feedback_source ON feedback_items(source);

CREATE TABLE IF NOT EXISTS opportunities (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'USABILITY',
    total_arr_usd REAL NOT NULL DEFAULT 0.0,
    weighted_score REAL NOT NULL DEFAULT 0.0,
    account_count INTEGER NOT NULL DEFAULT 0,
    feedback_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'DISCOVERED',
    generated_prd_markdown TEXT,
    generated_prd_notion_json TEXT,
    shipped_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_opportunities_status ON opportunities(status);
CREATE INDEX IF NOT EXISTS idx_opportunities_arr ON opportunities(total_arr_usd DESC);

CREATE TABLE IF NOT EXISTS feedback_opportunity_map (
    feedback_id TEXT NOT NULL REFERENCES feedback_items(id) ON DELETE CASCADE,
    opportunity_id TEXT NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    relevance_score REAL NOT NULL DEFAULT 1.0,
    PRIMARY KEY (feedback_id, opportunity_id)
);

CREATE INDEX IF NOT EXISTS idx_fom_opportunity ON feedback_opportunity_map(opportunity_id);

CREATE TABLE IF NOT EXISTS closed_loop_notifications (
    id TEXT PRIMARY KEY,
    opportunity_id TEXT NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    account_id TEXT REFERENCES accounts(id) ON DELETE SET NULL,
    customer_email TEXT NOT NULL,
    customer_name TEXT,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT',
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_opp ON closed_loop_notifications(opportunity_id);
`
