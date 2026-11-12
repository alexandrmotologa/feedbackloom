<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="feedbackloom Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">feedbackloom</h1>

<p align="center">
  <b>Autonomous Customer Voice and ARR-Weighted Feature Request Synthesizer</b><br>
  <i>From chaotic multi-channel feedback to revenue-ranked PRD specifications with verified customer quotes.</i>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go_1.24-00ADD8?style=flat-square&logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/Frontend-React_19_+_TypeScript-3178C6?style=flat-square&logo=react" alt="React 19" />
  <img src="https://img.shields.io/badge/Database-SQLite_WAL-003B57?style=flat-square&logo=sqlite" alt="SQLite" />
  <img src="https://img.shields.io/badge/Clustering-HDBSCAN_(Pure_Go)-10B981?style=flat-square" alt="HDBSCAN" />
  <img src="https://img.shields.io/badge/Styling-Tailwind_CSS_v4-38BDF8?style=flat-square&logo=tailwindcss" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/License-Apache_2.0-blue?style=flat-square" alt="License" />
</p>

<p align="center">
  <img src="docs/images/feedbackloom_demo.gif?raw=true" alt="feedbackloom Interactive Demo" width="900" style="border-radius: 12px; border: 1px solid #1e293b;" />
</p>

## Overview

feedbackloom ingests customer feedback across support tickets, sales transcripts, and community channels, clusters items by semantic similarity using pure Go HDBSCAN, weights each opportunity by customer contract value (ARR) and renewal proximity, and drafts developer-ready Product Requirements Documents (PRDs) with verified customer evidence quotes.

## Key Capabilities

- **Multi-Channel Ingestion**: Accepts feedback from Zendesk, Intercom, Gong transcripts, Slack, and batch CSV or JSON exports.
- **CRM and Revenue Attribution**: Connects customer email domains to Stripe subscriptions and HubSpot deals, pulling Annual Recurring Revenue (ARR), contract renewal dates, and churn risk scores.
- **Pure Go HDBSCAN Clustering**: Executes density-based spatial clustering over 128-dimensional dense vectors with zero Python runtime dependencies or external API costs. Granularity can be adjusted dynamically (Macro, Balanced, Micro).
- **Competitor Loss Attribution**: Identifies competitor mentions (such as Okta, Auth0, Splunk, Datadog) in feedback text and tallies Lost Pipeline ARR to expose immediate churn risks to named platforms.
- **ROI Index and Engineering Effort**: Links opportunity ARR to two-week engineering sprint allocations, providing an immediate ARR-per-sprint capital efficiency rating.
- **Interactive Ingestion Sandbox**: Real-time playground to test incoming customer quotes against the NLP pipeline, inspecting domain matching, sentiment/urgency, and nearest cluster cosine distance.
- **Autonomous PRD Synthesis**: Generates complete Product Requirements Documents in Markdown and Notion format, including user stories, Gherkin acceptance criteria, and verbatim customer quotes with ARR badges.
- **Bidirectional GitHub and Jira Exporters**: Formats PRDs into structured GitHub Issue payloads (with acceptance checklists and labels) and Jira Epic payloads for instant engineering handoff.
- **Renewal Danger Alerting**: Highlights opportunities demanded by high-value accounts facing renewals within 45 days.
- **Closed-Loop Customer Updates**: Drafts personalized notifications quoting the customer's original words when a feature transitions to shipped status.
- **Single Self-Contained Binary**: Packages the Go API daemon, SQLite WAL database, and React 19 Product Studio into a single binary requiring less than 25MB of RAM.

## Authentic Software Execution

### 1. Revenue Opportunities Waterfall & Competitor Loss Matrix
Live dashboard showing real-time portfolio coverage ($1.59M demanded ARR / 91.6% coverage), competitor churn attribution (Okta, Azure AD, Splunk, Datadog), and critical renewal alerts.

<p align="center">
  <img src="docs/images/feedbackloom_dashboard.png?raw=true" alt="feedbackloom Opportunities Dashboard" width="900" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### 2. Interactive Ingestion Sandbox & Real-Time Cosine Matching
Live payload tester demonstrating domain resolution against Stripe accounts, entity extraction, competitor detection, and cosine similarity matching (96.1% match against existing cluster).

<p align="center">
  <img src="docs/images/feedbackloom_sandbox.png?raw=true" alt="feedbackloom Ingestion Sandbox" width="900" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### 3. Autonomous PRD Studio with Customer Evidence Quotes
Complete Product Requirements Document with executive business context, problem statements, customer evidence quotes linked to ARR, and one-click export to GitHub Issues and Jira Epics.

<p align="center">
  <img src="docs/images/feedbackloom_prd_studio.png?raw=true" alt="feedbackloom PRD Studio" width="900" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### 4. Multi-Channel Data Connectors & Inbound Webhook Endpoints
Native connectors for Stripe Billing, HubSpot CRM, and incoming webhook triggers for Zendesk, Intercom, Gong, and Slack.

<p align="center">
  <img src="docs/images/feedbackloom_connectors.png?raw=true" alt="feedbackloom Data Connectors" width="900" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### 5. Multi-Device Adaptability & Responsive Design
Full mobile and tablet viewport adaptability with collapsible cards, touch-friendly navigation, and responsive financial metrics.

<p align="center">
  <img src="docs/images/feedbackloom_mobile.png?raw=true" alt="feedbackloom Mobile Responsive Layout" width="360" style="border-radius: 16px; border: 1px solid #1e293b;" />
</p>

## System Architecture

```
                       +---------------------------------------------+
                       |           Multi-Channel Ingestion           |
                       |  (Zendesk, Intercom, Gong, Slack, CSV/JSON) |
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
                       |  - Matches email domain to customer account |
                       |  - Pulls ARR, plan tier, and renewal date   |
                       |  - Tracks account churn risk probability    |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |       Semantic Clustering & Extraction      |
                       |  - Distills core pain point and outcome     |
                       |  - 128-dim dense subword vectorizer         |
                       |  - Pure Go HDBSCAN density clustering       |
                       |  - Detects competitor churn signals         |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |         ARR-Weighted Opportunity Rank       |
                       |  - Sums Total ARR per unique account        |
                       |  - Tallies Lost Pipeline ARR to competitors |
                       |  - Computes Weighted Revenue Score (WRS)    |
                       |  - Calculates ROI Index ($ARR / Sprint)     |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |         Autonomous PRD Synthesizer          |
                       |  - Executive summary and problem statement  |
                       |  - Verified customer evidence quotes        |
                       |  - User stories and Gherkin criteria        |
                       |  - Non-functional specs and guardrails      |
                       +---------------------------------------------+
                                              |
                         +--------------------+--------------------+
                         |                                         |
                         v                                         v
         +-------------------------------+         +-------------------------------+
         |     Exported PRD Document     |         |    React 19 Product Studio    |
         |  - Standalone Markdown (.md)  |         |  - Revenue Waterfall ranking  |
         |  - Notion API block trees     |         |  - Competitor Loss Matrix     |
         |  - GitHub Issue checklists    |         |  - Ingestion Sandbox tester   |
         |  - Jira Epic specifications   |         |  - 1-click Closed-Loop emails |
         +-------------------------------+         +-------------------------------+
```

## Mathematical Revenue Weighting

Standard product management tools count raw feedback frequency, which tends to over-index on free-tier requests while obscuring enterprise blocking requirements.

For a cluster $K$ comprising feedback items $\{f_1, f_2, \dots, f_n\}$ originating from customer accounts $A(f_i)$:

$$\text{Total ARR}(K) = \sum_{a \in \text{UniqueAccounts}(K)} \text{ARR}(a)$$

To factor in contract renewal urgency and churn probability, feedbackloom calculates the **Weighted Revenue Score ($WRS$)**:

$$WRS(K) = \sum_{a \in \text{UniqueAccounts}(K)} \left[ \text{ARR}(a) \times \left( 1 + \frac{1}{\max(1, \text{DaysToRenewal}(a))} \right) \times \text{ChurnRisk}(a) \right]$$

- Deduplication by unique account ensures an account submitting multiple requests contributes its ARR only once.
- Contracts nearing renewal receive an elevated urgency multiplier ($1 / \text{DaysToRenewal}$).
- Churn risk score scales the retention priority.
- ROI Index balances revenue against sprint complexity: $\text{ROI}(K) = \frac{\text{Total ARR}(K)}{\max(1, \text{EffortSprints}(K))}$.

## Quick Start

### 1. Build from Source

Requirements: Go 1.22+ and Node.js 20+.

```bash
# Clone the repository
git clone https://github.com/alexandrmotologa/feedbackloom.git
cd feedbackloom

# Build the embedded React 19 frontend
cd web && npm install && npm run build && cd ..

# Compile the single Go binary
go build -ldflags="-s -w" -o feedbackloom ./cmd/feedbackloom
```

### 2. Start the Server

```bash
# Launch on port 8080 with SQLite storage
./feedbackloom serve --port 8080 --db ./feedbackloom.db
```

Open `http://localhost:8080` in your browser to access the embedded Product Studio.

### 3. Seed Benchmark Data and Run Clustering

```bash
# Sync customer accounts from Stripe (or realistic demo accounts)
./feedbackloom sync-crm --provider stripe

# Ingest realistic multi-channel customer quotes
./feedbackloom ingest seed

# Check system stats
./feedbackloom stats

# Execute HDBSCAN clustering and rank opportunities
./feedbackloom cluster

# Synthesize a complete PRD for the highest-ranked opportunity
./feedbackloom generate-prd --output ./PRD_Enterprise_SSO.md
```

## CLI Reference

| Command | Flags | Description |
| :--- | :--- | :--- |
| `serve` | `--port`, `--db` | Starts the HTTP REST API server and embedded web studio |
| `sync-crm` | `--provider`, `--stripe-key`, `--hubspot-key`, `--csv` | Synchronizes customer ARR accounts |
| `ingest` | `csv`, `json`, `seed` | Imports customer feedback from files or seeds demo data |
| `cluster` | `--min-cluster-size`, `--min-points` | Executes HDBSCAN clustering and ranks opportunities |
| `generate-prd` | `--opportunity-id`, `--output` | Generates a Markdown PRD with Gherkin criteria |
| `export-notion` | `--opportunity-id`, `--notion-key`, `--page-id` | Publishes a PRD directly to Notion |
| `stats` | `--db` | Prints database metrics and portfolio ARR coverage |

## REST API Endpoints

- `GET /health`: Service health check.
- `GET /api/v1/stats`: Portfolio and demanded ARR metrics, plus competitor lost pipeline figures.
- `GET /api/v1/accounts`: Customer accounts sorted by ARR.
- `POST /api/v1/crm/sync`: Trigger Stripe or HubSpot sync.
- `GET /api/v1/feedback`: Paginated customer feedback items.
- `POST /api/v1/ingest/zendesk`: Zendesk support ticket webhook.
- `POST /api/v1/ingest/intercom`: Intercom conversation webhook.
- `POST /api/v1/ingest/gong`: Gong call transcript webhook.
- `POST /api/v1/ingest/slack`: Slack events API webhook.
- `POST /api/v1/ingest/test-live`: Interactive sandbox payload tester with domain resolution and cosine cluster match.
- `POST /api/v1/cluster/run`: Trigger HDBSCAN semantic clustering with optional granularity parameters.
- `GET /api/v1/opportunities`: Ranked opportunities list with competitor exposure, renewal alerts, and velocity.
- `GET /api/v1/opportunities/:id`: Opportunity details and customer quotes.
- `PATCH /api/v1/opportunities/:id/effort`: Set sprint effort estimate and recalculate ROI index.
- `POST /api/v1/opportunities/:id/prd`: Synthesize PRD document.
- `POST /api/v1/opportunities/:id/export-notion`: Publish PRD to Notion.
- `POST /api/v1/opportunities/:id/export-github`: Export PRD as a structured GitHub Issue payload.
- `POST /api/v1/opportunities/:id/export-jira`: Export PRD as a formatted Jira Epic payload.
- `PATCH /api/v1/opportunities/:id/status`: Update lifecycle status.
- `POST /api/v1/opportunities/:id/notifications/dispatch`: Send closed-loop updates.

## Technical Specifications

- **Language**: Go 1.24
- **Database**: SQLite with Write-Ahead Logging (WAL) mode via `modernc.org/sqlite` (pure Go, zero CGO)
- **Vector Embeddings**: 128-dimensional dense vectorizer with L2 normalization
- **Clustering Algorithm**: HDBSCAN with Prim's Minimum Spanning Tree ($O(N^2)$)
- **Web Frontend**: React 19, TypeScript, Vite, Tailwind CSS v4 (embedded directly into Go binary)
- **Binary Size**: ~12MB single static executable with zero external dependencies
- **Memory Footprint**: < 25MB RAM under active load

## License

Apache-2.0 License. See LICENSE for details.
