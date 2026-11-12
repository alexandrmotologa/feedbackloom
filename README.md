# feedbackloom

Autonomous Customer Voice and ARR-Weighted Feature Request Synthesizer.

feedbackloom ingests noisy customer feedback across support tickets, sales transcripts, and community channels, clusters them by semantic similarity using HDBSCAN, weights each opportunity by customer contract value (ARR) and renewal urgency, and drafts developer-ready Product Requirements Documents (PRDs) with verified customer evidence quotes.

## Key Capabilities

- **Multi-Channel Ingestion**: Accepts feedback from Zendesk, Intercom, Gong transcripts, Slack, and batch CSV or JSON exports.
- **CRM and Revenue Attribution**: Connects customer email domains to Stripe subscriptions and HubSpot deals, pulling Annual Recurring Revenue (ARR), contract renewal dates, and churn risk scores.
- **Pure Go HDBSCAN Clustering**: Executes density-based spatial clustering over 128-dimensional dense vectors with zero Python runtime dependencies or external API costs.
- **Weighted Revenue Score (WRS)**: Prioritizes product themes by the actual revenue at stake, factoring in proximity to renewal dates and account risk.
- **Autonomous PRD Synthesis**: Generates complete Product Requirements Documents in Markdown and Notion format, including user stories, Gherkin acceptance criteria, and verbatim customer quotes with ARR badges.
- **Closed-Loop Customer Updates**: Drafts personalized notifications quoting the customer's original words when a feature transitions to shipped status.
- **Single Self-Contained Binary**: Packages the Go API daemon, SQLite WAL database, and React 19 Product Studio into a single binary requiring less than 25MB of RAM.

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
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |         ARR-Weighted Opportunity Rank       |
                       |  - Sums Total ARR per unique account        |
                       |  - Applies renewal proximity multiplier     |
                       |  - Computes Weighted Revenue Score (WRS)    |
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
         |  - Notion API block trees     |         |  - Customer Quote Inspector   |
         |  - Jira / GitHub issue specs  |         |  - 1-click Closed-Loop emails |
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
- `GET /api/v1/stats`: Portfolio and demanded ARR metrics.
- `GET /api/v1/accounts`: Customer accounts sorted by ARR.
- `POST /api/v1/crm/sync`: Trigger Stripe or HubSpot sync.
- `GET /api/v1/feedback`: Paginated customer feedback items.
- `POST /api/v1/ingest/zendesk`: Zendesk support ticket webhook.
- `POST /api/v1/ingest/intercom`: Intercom conversation webhook.
- `POST /api/v1/ingest/gong`: Gong call transcript webhook.
- `POST /api/v1/ingest/slack`: Slack events API webhook.
- `POST /api/v1/cluster/run`: Trigger HDBSCAN semantic clustering.
- `GET /api/v1/opportunities`: Ranked opportunities list.
- `GET /api/v1/opportunities/:id`: Opportunity details and customer quotes.
- `POST /api/v1/opportunities/:id/prd`: Synthesize PRD document.
- `POST /api/v1/opportunities/:id/export-notion`: Publish PRD to Notion.
- `PATCH /api/v1/opportunities/:id/status`: Update lifecycle status.
- `POST /api/v1/opportunities/:id/notifications/dispatch`: Send closed-loop updates.

## Technical Specifications

- **Language**: Go 1.27
- **Database**: SQLite with Write-Ahead Logging (WAL) mode via `modernc.org/sqlite` (pure Go, zero CGO)
- **Vector Embeddings**: 128-dimensional dense vectorizer with L2 normalization
- **Clustering Algorithm**: HDBSCAN with Prim's Minimum Spanning Tree ($O(N^2)$)
- **Web Frontend**: React 19, TypeScript, Vite, Tailwind CSS (embedded directly into Go binary)
- **Binary Size**: ~12MB single static executable with zero external dependencies
- **Memory Footprint**: < 25MB RAM under active load

## License

Apache-2.0 License. See LICENSE for details.
