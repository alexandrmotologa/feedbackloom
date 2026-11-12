# feedbackloom Architecture Specification

feedbackloom converts disparate customer feedback streams into revenue-ranked product specifications with verified customer quotes.

## 1. System Components

The engine is built as a single Go binary with an embedded React 19 web application and SQLite persistence:

1. **Multi-Channel Ingestion Pipeline (`internal/ingest`)**
   Accepts feedback from Zendesk, Intercom, Gong transcripts, Slack channels, and batch CSV or JSON files. Webhook endpoints validate payload structures and extract customer quotes.

2. **CRM and Revenue Matcher (`internal/crm`)**
   Resolves customer email addresses to corporate domains (`cto@cyberdyne.com` -> `cyberdyne.com`), filtering out consumer webmail providers (such as Gmail, Yahoo, and Outlook). Connects to Stripe billing and HubSpot deals to pull Annual Recurring Revenue (ARR), contract renewal dates, and churn risk metrics.

3. **Dense Vectorizer (`internal/cluster/vectorizer.go`)**
   Generates 128-dimensional dense vectors using a deterministic subword n-gram vectorizer with L2 normalization. Operates completely offline without requiring paid API keys or external network requests.

4. **HDBSCAN Clustering Engine (`internal/cluster/hdbscan.go`)**
   Groups dense feedback vectors by density:
   - Computes pairwise Euclidean distance matrices.
   - Calculates core distances for min points $k = 2$.
   - Builds a Minimum Spanning Tree (MST) on mutual reachability graphs using Prim's algorithm.
   - Condenses clusters satisfying minimum cluster size thresholds.
   - Labels outliers as noise ($label = -1$).

5. **Mathematical Revenue Ranking Matrix (`internal/ranking`)**
   Scores product themes using the Weighted Revenue Score (WRS), factoring in ARR, renewal proximity, and account churn risk.

6. **Autonomous PRD Synthesizer (`internal/prd`)**
   Generates comprehensive Product Requirements Documents with executive summaries, customer evidence citations, user stories, and Gherkin acceptance criteria (`Feature`, `Scenario`, `Given`, `When`, `Then`).

7. **Closed-Loop Communications (`internal/closedloop`)**
   Drafts personalized customer emails quoting original feedback when an opportunity transitions to "SHIPPED".

8. **Embedded Product Studio (`web/`)**
   A React 19 single-page application served directly from the Go binary via `embed.FS`.

## 2. Ingestion to PRD Data Flow

```
+-----------------------------------------------------------+
| Multi-Channel Ingestion: Zendesk, Intercom, Gong, Slack   |
+-----------------------------------------------------------+
                             |
                             v
+-----------------------------------------------------------+
| CRM Domain Matcher: Map customer email to Account ARR     |
+-----------------------------------------------------------+
                             |
                             v
+-----------------------------------------------------------+
| Semantic Vectorizer: 128-dim dense float32 vector         |
+-----------------------------------------------------------+
                             |
                             v
+-----------------------------------------------------------+
| HDBSCAN Density Clustering: Group into product themes     |
+-----------------------------------------------------------+
                             |
                             v
+-----------------------------------------------------------+
| Revenue Scoring: Sum ARR and compute Weighted WRS         |
+-----------------------------------------------------------+
                             |
                             v
+-----------------------------------------------------------+
| PRD Synthesizer: Generate Markdown and Notion block JSON  |
+-----------------------------------------------------------+
                             |
                             v
+-----------------------------------------------------------+
| Closed-Loop Updates: Dispatch email updates to requesters |
+-----------------------------------------------------------+
```

## 3. Storage Layer

The SQLite store uses Write-Ahead Logging (WAL) mode for concurrency:
- `accounts`: Customer names, domains, ARR in USD, plan tier, renewal date, and churn risk.
- `feedback_items`: Raw text, extracted pain points, customer email, urgency score, and binary vector blob.
- `opportunities`: Synthesized title, description, category, total ARR, WRS score, and generated PRD markdown.
- `feedback_opportunity_map`: Links individual feedback items to opportunities.
- `closed_loop_notifications`: Generated customer update emails and dispatch status.
