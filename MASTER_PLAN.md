# `feedbackloom`: Architecture & Engineering Master Plan
> **Autonomous Customer Voice & ARR-Weighted Feature Request Synthesizer**  
> *Target: AI Product Engineering, Product Ops Automation, Revenue-Weighted Roadmaps & Automated PRD Synthesis*

---

## 1. Executive Summary & Market Problem

### 1.1 The Fragmented Feedback Trap in Product Management
Modern SaaS product management teams are overwhelmed by thousands of fragmented customer requests arriving daily across disparate channels:
- Support tickets in **Zendesk and Intercom**.
- Sales conversation transcripts from **Gong, Zoom, and Fireflies**.
- Community discussions on **Slack, Discord, and Reddit**.
- Bug reports and feature requests in **GitHub Issues and Jira**.

Current approaches to feedback analysis fail because of two fatal flaws:
1. **The "Loudest Voice / Frequency" Bias**: Standard tools (word clouds, keyword counting) measure volume rather than business value. A minor bug reported by 50 free-tier users looks 10x more important than an enterprise blocking issue demanded by 2 Fortune 500 accounts representing \$300,000 in ARR.
2. **The PRD Drafting Bottleneck**: Even when a critical product gap is identified, Product Managers spend 15 to 20 hours writing a comprehensive Product Requirements Document (PRD), detailing user stories, edge cases, and customer evidence quotes.

### 1.2 The `feedbackloom` Solution
`feedbackloom` is an autonomous, self-hosted Go and React 19 engine that turns noisy customer signals into **revenue-weighted product specifications**:
- **Multi-Channel Ingestion**: Automatically streams feedback from Zendesk, Intercom, Gong transcripts, Slack, and GitHub.
- **CRM & Revenue Enrichment**: Connects directly to Stripe, HubSpot, or Salesforce to tag every customer quote with **Account ARR, Contract Renewal Date, and Churn Risk**.
- **Semantic Opportunity Clustering**: Uses HDBSCAN and embedding distance to group raw complaints into actionable product themes (not just keywords).
- **ARR-Weighted Opportunity Matrix**: Ranks features by the actual dollars at risk or dollars demanding the capability.
- **Autonomous PRD Synthesizer**: Generates complete, developer-ready Product Requirements Documents (PRDs) in Markdown and Notion, complete with real customer evidence quotes, user stories, and acceptance criteria.
- **Closed-Loop Customer Updates**: When a feature is marked "Shipped", automatically drafts personalized emails to all affected accounts.

---

## 2. High-Level System Architecture

```
                       +---------------------------------------------+
                       |           Multi-Channel Ingestion           |
                       | (Zendesk, Intercom, Gong, Slack, GitHub)    |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |             feedbackloom Daemon             |
                       |       (Single Go Binary, <25MB RAM)         |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |          CRM & Revenue Enrichment           |
                       |  - Matches Email Domain to Account          |
                       |  - Fetches ARR, Plan Tier, Renewal Date     |
                       |  - Computes Account Churn Risk Score        |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |       Semantic Clustering & Extraction      |
                       |  - Extracts Core Pain Point & Desired Outcome|
                       |  - Dense Embeddings (Local ONNX / OpenAI)   |
                       |  - HDBSCAN Dynamic Theme Clustering         |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |         ARR-Weighted Opportunity Rank       |
                       |  - Sums Total ARR Demanding Feature         |
                       |  - Weighs Immediacy & Renewal Proximity     |
                       |  - Identifies Top 3 High-ROI Initiatives    |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |         Autonomous PRD Synthesizer          |
                       |  - Executive Summary & Problem Context      |
                       |  - Verified Customer Quotes with ARR tags   |
                       |  - User Stories & Gherkin Acceptance Criteria|
                       |  - Out-of-Scope Technical Guardrails        |
                       +---------------------------------------------+
                                              |
                         +--------------------+--------------------+
                         |                                         |
                         v                                         v
         +-------------------------------+         +-------------------------------+
         |     Exported PRD Document     |         |    React 19 Product Studio    |
         |  - Notion API Export          |         |  - ARR-Weighted Opportunity UI|
         |  - GitHub Issue / Markdown    |         |  - Customer Quote Inspector   |
         |  - Jira Epic Generation       |         |  - 1-Click Closed-Loop Emails |
         +-------------------------------+         +-------------------------------+
```

---

## 3. Mathematical Revenue-Weighted Opportunity Scoring

Given an opportunity cluster $K$ comprising individual feedback items $\{f_1, f_2, \dots, f_n\}$ originating from customer accounts $A(f_i)$:

$$\text{Total ARR}(K) = \sum_{a \in \text{UniqueAccounts}(K)} \text{ARR}(a)$$

To account for contract renewal urgency and churn probability, we define the **Weighted Revenue Score ($WRS$)**:
$$WRS(K) = \sum_{a \in \text{UniqueAccounts}(K)} \left[ \text{ARR}(a) \times \left( 1 + \frac{1}{\max(1, \text{DaysToRenewal}(a))} \right) \times \text{ChurnRisk}(a) \right]$$

This mathematical ranking ensures product engineers and PMs focus on features that directly defend revenue before contracts lapse.

---

## 4. SQLite Data Schema (`store/schema.sql`)

```sql
-- Customer Accounts from Stripe / HubSpot
CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT UNIQUE NOT NULL,
    arr_usd REAL NOT NULL DEFAULT 0.0,
    plan_tier TEXT NOT NULL,           -- 'ENTERPRISE', 'GROWTH', 'STARTER'
    renewal_date DATE,
    churn_risk_score REAL DEFAULT 0.2, -- 0.0 (safe) to 1.0 (imminent churn)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ingested Customer Feedback Items
CREATE TABLE IF NOT EXISTS feedback_items (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,              -- 'ZENDESK', 'INTERCOM', 'GONG', 'SLACK'
    account_id TEXT REFERENCES accounts(id),
    customer_email TEXT,
    raw_text TEXT NOT NULL,
    extracted_pain_point TEXT,
    embedding BLOB,                    -- Packed float32 dense vector
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Discovered Opportunity Themes (Clusters)
CREATE TABLE IF NOT EXISTS opportunities (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,            -- 'PERFORMANCE', 'INTEGRATIONS', 'SECURITY', 'USABILITY'
    total_arr_usd REAL NOT NULL,
    account_count INTEGER NOT NULL,
    status TEXT NOT NULL,              -- 'DISCOVERED', 'IN_PRD', 'SHIPPED'
    generated_prd_markdown TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Mapping Feedback to Opportunities
CREATE TABLE IF NOT EXISTS feedback_opportunity_map (
    feedback_id TEXT REFERENCES feedback_items(id) ON DELETE CASCADE,
    opportunity_id TEXT REFERENCES opportunities(id) ON DELETE CASCADE,
    relevance_score REAL NOT NULL,
    PRIMARY KEY (feedback_id, opportunity_id)
);
```

---

## 5. Repository & Directory Structure

```
feedbackloom/
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
├── cmd/
│   └── feedbackloom/
│       └── main.go                 # CLI entrypoint (ingest, cluster, prd, serve)
├── internal/
│   ├── config/                     # Configuration and API credentials
│   ├── ingest/                     # Inbound Feed Connectors
│   │   ├── zendesk.go              # Zendesk tickets API
│   │   ├── intercom.go             # Intercom conversations API
│   │   ├── gong.go                 # Gong call transcript parser
│   │   ├── slack.go                # Slack webhook/bot listener
│   │   └── csv_loader.go           # Bulk import for historical exports
│   ├── crm/                        # Revenue Data Connectors
│   │   ├── stripe.go               # Syncs active customer subscriptions & ARR
│   │   ├── hubspot.go              # Matches customer emails to company deals
│   │   └── resolver.go             # Email domain resolution logic
│   ├── cluster/                    # Semantic Clustering Engine
│   │   ├── embedder.go             # Vector generation
│   │   ├── hdbscan.go              # Density-based clustering
│   │   └── namer.go                # LLM agent synthesizing concise feature titles
│   ├── prd/                        # Autonomous PRD Synthesizer
│   │   ├── generator.go            # LLM prompt constructing full product requirements
│   │   ├── user_stories.go         # Gherkin-format acceptance criteria builder
│   │   └── notion_exporter.go      # Publishes PRDs directly into Notion databases
│   ├── store/                      # SQLite Persistence Layer
│   │   ├── db.go
│   │   └── opportunity_store.go
│   └── api/                        # HTTP REST & Webhook Receiver
│       ├── handler.go
│       └── routes.go
├── web/                            # Embedded React 19 Product Studio
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── App.tsx
│       ├── components/
│       │   ├── ARRWaterfall.tsx    # Visual ranking of features by dollars at stake
│       │   ├── PRDViewer.tsx       # Live markdown PRD previewer & editor
│       │   ├── QuoteStream.tsx     # Verifiable customer quotes with ARR badges
│       │   └── ClosedLoopModal.tsx # One-click email broadcast to requesters
│       └── pages/
│           ├── Opportunities.tsx
│           ├── PRDDetail.tsx
│           └── Integrations.tsx
├── examples/
│   ├── sample_feedback.csv
│   └── sample_stripe_customers.csv
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 6. CLI Command Surface

```bash
# Start the feedbackloom server and webhook receiver
feedbackloom serve --port 8080 --db ./feedback.db

# Ingest historical feedback from CSV
feedbackloom ingest csv --file ./support_tickets.csv --source zendesk

# Sync customer ARR data from Stripe or CSV
feedbackloom sync-crm --stripe-key "${STRIPE_SECRET_KEY}"

# Cluster all unassigned feedback items and rank opportunities
feedbackloom cluster --min-cluster-size 3

# Generate an autonomous PRD for the highest ARR opportunity
feedbackloom generate-prd --opportunity-id opp_87a2f1 --output ./PRD_SAML_SSO.md

# Publish PRD directly into Notion workspace
feedbackloom export-notion --opportunity-id opp_87a2f1 --notion-key "${NOTION_TOKEN}"
```

---

## 7. Implementation Roadmap (5 Phases)

### Phase 1: Ingestion Pipelines & CRM Revenue Matcher
- Implement multi-format feedback ingestion (CSV, JSON, Webhook).
- Build Stripe and HubSpot connectors mapping customer email domains to ARR.
- Set up SQLite schema with WAL mode and vector BLOB storage.

### Phase 2: Semantic Extraction & Clustering Engine
- Implement LLM extraction prompt distilling raw chat into clear pain points.
- Implement HDBSCAN / K-means clustering over dense embeddings.
- Synthesize crisp, action-oriented titles for discovered opportunity clusters.

### Phase 3: Mathematical Revenue Ranking & Opportunity Matrix
- Compute Total ARR and Renewal-Weighted Score ($WRS$) per feature.
- Filter out free/low-tier noise while highlighting high-risk enterprise retention blockers.
- Provide JSON API for ranked product initiatives.

### Phase 4: Autonomous PRD Synthesizer & Notion Exporter
- Build comprehensive PRD prompt generating problem statements, user stories, and acceptance criteria.
- Automatically cite top 5 customer quotes with company names and ARR badges.
- Implement Notion API integration to create rich formatted pages automatically.

### Phase 5: Embedded React 19 Studio & Closed-Loop Updates
- Build interactive product studio with ARR waterfall charts and quote inspector.
- Add 1-click customer email broadcast when an opportunity transitions to "Shipped".
- Package into single Go binary and write comprehensive documentation.
