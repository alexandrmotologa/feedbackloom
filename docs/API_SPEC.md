# feedbackloom REST API Reference

The feedbackloom HTTP daemon exposes REST endpoints on port 8080 (or the value set in the `PORT` environment variable).

## Health and System Stats

### `GET /health`
Returns service status.

**Response `200 OK`**:
```json
{
  "service": "feedbackloom",
  "status": "healthy",
  "timestamp": "2026-11-12T12:00:00Z"
}
```

### `GET /api/v1/stats`
Returns aggregated portfolio figures, including total portfolio ARR, demanded ARR, and competitor lost pipeline ARR.

**Response `200 OK`**:
```json
{
  "total_accounts": 12,
  "portfolio_arr_usd": 1736500.0,
  "total_feedback": 13,
  "total_opportunities": 2,
  "total_demanded_arr_usd": 1591500.0,
  "total_lost_pipeline_arr_usd": 680000.0
}
```

---

## Accounts and CRM

### `GET /api/v1/accounts`
Lists all customer accounts sorted by ARR descending.

### `POST /api/v1/crm/sync?provider=stripe|hubspot`
Synchronizes accounts and subscriptions from Stripe or HubSpot.

---

## Feedback Ingestion and Live Sandbox

### `GET /api/v1/feedback`
Returns paginated customer feedback items.

### `POST /api/v1/feedback`
Submits a single feedback item.

**Request body**:
```json
{
  "source": "ZENDESK",
  "customer_email": "cto@cyberdyne.com",
  "customer_name": "Miles Dyson",
  "raw_text": "We cannot renew without Okta SAML 2.0 single sign-on."
}
```

### `POST /api/v1/feedback/seed`
Seeds benchmark multi-channel enterprise feedback.

### `POST /api/v1/ingest/csv`
Uploads feedback CSV data directly in the request body.

### `POST /api/v1/ingest/zendesk`
Webhook endpoint for Zendesk ticket payloads.

### `POST /api/v1/ingest/intercom`
Webhook endpoint for Intercom conversation events.

### `POST /api/v1/ingest/gong`
Webhook endpoint for Gong call transcripts.

### `POST /api/v1/ingest/slack`
Webhook endpoint for Slack events API and incoming messages.

### `POST /api/v1/ingest/test-live`
Interactive sandbox endpoint for real-time payload testing. Executes domain resolution, ARR extraction, sentiment/urgency scoring, competitor identification, subword vectorization, and cosine similarity matching against existing opportunity clusters.

**Request body**:
```json
{
  "source": "zendesk",
  "customer_email": "sarah.connor@cyberdyne.com",
  "customer_name": "Sarah Connor",
  "raw_text": "Our enterprise compliance mandate requires Okta SAML 2.0 single sign-on with SCIM automated user provisioning. Without this before Q3 renewal, our CISO will block contract extension and migrate to Auth0.",
  "commit": false
}
```

**Response `200 OK`**:
```json
{
  "account": {
    "id": "acc_cyberdyne",
    "name": "Cyberdyne Systems",
    "domain": "cyberdyne.com",
    "arr_usd": 240000.0,
    "plan_tier": "ENTERPRISE",
    "renewal_date": "2026-12-15",
    "churn_risk_score": 0.82
  },
  "extracted_pain_point": "enterprise compliance mandate requires Okta SAML 2.0 single sign-on with SCIM automated user provisioning",
  "extracted_outcome": "block contract extension and migrate to Auth0",
  "category": "SECURITY",
  "urgency_score": 0.95,
  "sentiment_score": -0.7,
  "competitor_mentions": ["Okta", "Auth0"],
  "matched_opportunity_id": "opp-sec-01",
  "matched_opportunity_title": "Enterprise SSO: Okta and Azure AD Integration",
  "similarity_score": 0.884,
  "created_item_id": ""
}
```

---

## Clustering and Opportunities

### `POST /api/v1/cluster/run?min_cluster_size=2&min_points=2`
Executes subword n-gram vector embeddings and pure Go HDBSCAN clustering. Accepts optional query parameters to adjust clustering granularity:
- `min_cluster_size` (default: 2, range: 2 to 10)
- `min_points` (default: 2, range: 1 to 10)

**Response `200 OK`**:
```json
{
  "total_processed": 13,
  "clusters_created": 2,
  "noise_item_count": 2,
  "total_demanded_arr_usd": 1591500.0,
  "duration_ms": 27
}
```

### `GET /api/v1/opportunities`
Lists synthesized opportunities ordered by Weighted Revenue Score (WRS). Includes competitor exposure metrics, ROI index, renewal urgency alerts, and feedback velocity figures.

### `GET /api/v1/opportunities/:id`
Returns opportunity details along with linked customer quotes and ARR metrics.

### `PATCH /api/v1/opportunities/:id/effort`
Updates engineering effort estimate in sprints and recalculates ROI Index ($ARR / Sprints).

**Request body**:
```json
{
  "effort_sprints": 4
}
```

### `POST /api/v1/opportunities/:id/prd`
Generates a complete Product Requirements Document in Markdown and Notion format.

### `POST /api/v1/opportunities/:id/export-notion`
Publishes the generated PRD directly to a Notion workspace.

### `POST /api/v1/opportunities/:id/export-github`
Generates a structured GitHub Issue payload containing executive summary, customer evidence quotes, acceptance criteria checklists, and scoped issue labels.

### `POST /api/v1/opportunities/:id/export-jira`
Generates a Jira Epic payload containing formatted description, business rationale, technical constraints, and issue type metadata.

### `PATCH /api/v1/opportunities/:id/status`
Updates opportunity lifecycle status (`DISCOVERED`, `IN_PRD`, `PLANNED`, `IN_PROGRESS`, `SHIPPED`).

---

## Closed-Loop Communications

### `GET /api/v1/opportunities/:id/notifications`
Lists customer outreach drafts for an opportunity.

### `POST /api/v1/opportunities/:id/notifications/generate`
Generates personalized update drafts quoting each customer's original feedback.

### `POST /api/v1/opportunities/:id/notifications/dispatch`
Sends all pending customer notifications.
