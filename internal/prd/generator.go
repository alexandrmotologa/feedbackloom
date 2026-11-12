package prd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/ranking"
)

// PRDGenerator creates comprehensive Product Requirements Documents from opportunity clusters.
type PRDGenerator struct{}

// NewPRDGenerator initializes a new PRD generator.
func NewPRDGenerator() *PRDGenerator {
	return &PRDGenerator{}
}

// GeneratePRD creates Markdown and Notion JSON specs for an opportunity and its evidence quotes.
func (g *PRDGenerator) GeneratePRD(ctx context.Context, opp *models.Opportunity, quotes []models.EnrichedFeedbackItem) (string, string, error) {
	metrics := ranking.CalculateOpportunityMetrics(quotes)

	var sb strings.Builder

	// Header metadata
	sb.WriteString(fmt.Sprintf("# PRD: %s\n\n", opp.Title))
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("| :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| **Status** | %s |\n", opp.Status))
	sb.WriteString(fmt.Sprintf("| **Category** | %s |\n", opp.Category))
	sb.WriteString(fmt.Sprintf("| **Total ARR at Stake** | $%.2f USD |\n", opp.TotalARRUSD))
	sb.WriteString(fmt.Sprintf("| **Weighted Revenue Score** | %.2f |\n", opp.WeightedScore))
	sb.WriteString(fmt.Sprintf("| **Requesting Accounts** | %d unique accounts (%d total quotes) |\n", metrics.AccountCount, len(quotes)))
	sb.WriteString(fmt.Sprintf("| **Target Quarter** | Q%d %d |\n", (int(time.Now().Month())-1)/3+1, time.Now().Year()))
	sb.WriteString(fmt.Sprintf("| **Created** | %s |\n\n", time.Now().Format("2006-01-02")))

	// Section 1: Executive summary
	sb.WriteString("## 1. Executive Summary and Business Context\n\n")
	sb.WriteString(fmt.Sprintf(
		"This initiative addresses a critical product capability gap demanded by %d customer accounts representing $%.2f in Annual Recurring Revenue. ",
		metrics.AccountCount, opp.TotalARRUSD,
	))
	if metrics.EnterpriseARRUSD > 0 {
		sb.WriteString(fmt.Sprintf(
			"Enterprise tier accounts represent $%.2f of this total. ",
			metrics.EnterpriseARRUSD,
		))
	}
	if metrics.MinDaysToRenewal < 60 {
		sb.WriteString(fmt.Sprintf(
			"Customer accounts with pending renewals inside %d days require this capability to avoid contract churn. ",
			metrics.MinDaysToRenewal,
		))
	}
	sb.WriteString("\n\n")

	// Section 2: Problem statement
	sb.WriteString("## 2. Problem Statement\n\n")
	sb.WriteString(opp.Description)
	sb.WriteString("\n\n")
	sb.WriteString("Without this capability, customer security and operations teams must either implement manual workarounds or halt company-wide rollouts, creating immediate churn risk during contract renewal evaluations.\n\n")

	// Section 3: Verified customer evidence
	sb.WriteString("## 3. Verified Customer Evidence Quotes\n\n")
	sb.WriteString("The following direct customer quotes substantiate this product requirement:\n\n")

	for i, q := range quotes {
		if i >= 6 {
			break // Top 6 quotes
		}
		name := q.AccountName
		if name == "" {
			name = q.AccountDomain
		}
		if name == "" {
			name = "Anonymous Requester"
		}

		sb.WriteString(fmt.Sprintf("### Quote %d: %s ($%.0f ARR, %s Plan)\n", i+1, name, q.AccountARRUSD, q.PlanTier))
		sb.WriteString(fmt.Sprintf("> \"%s\"\n\n", strings.TrimSpace(q.RawText)))
		sb.WriteString(fmt.Sprintf(
			"- **Channel Source**: %s\n- **Reported By**: %s (%s)\n- **Days to Renewal**: %d days\n- **Urgency Score**: %.2f / 1.00\n\n",
			q.Source, q.CustomerName, q.CustomerEmail, q.DaysToRenewal, q.UrgencyScore,
		))
	}

	// Section 4: User personas
	sb.WriteString("## 4. Target Personas\n\n")
	sb.WriteString("1. **Enterprise Security and Compliance Lead**: Needs centralized authentication enforcement and verifiable audit trails to satisfy corporate InfoSec policies.\n")
	sb.WriteString("2. **Workspace Administrator**: Configures identity connections, provisions user seats, and maps organizational roles without opening support tickets.\n")
	sb.WriteString("3. **End User Employee**: Accesses the platform using familiar single-click enterprise credentials without managing separate passwords.\n\n")

	// Section 5: User stories and Gherkin acceptance criteria
	sb.WriteString("## 5. User Stories and Gherkin Acceptance Criteria\n\n")

	switch opp.Category {
	case models.CategorySecurity:
		if strings.Contains(strings.ToLower(opp.Title), "sso") {
			sb.WriteString(generateSSOUserStories())
		} else if strings.Contains(strings.ToLower(opp.Title), "audit") {
			sb.WriteString(generateAuditUserStories())
		} else {
			sb.WriteString(generateRBACUserStories())
		}
	case models.CategoryPerformance:
		sb.WriteString(generateExportUserStories())
	default:
		sb.WriteString(generateDefaultUserStories(opp.Title))
	}

	// Section 6: Non-functional requirements
	sb.WriteString("## 6. Technical and Non-Functional Requirements\n\n")
	sb.WriteString("- **Availability SLO**: 99.95% uptime for the authentication and webhook ingress pipelines.\n")
	sb.WriteString("- **Latency**: P95 response times must remain below 150ms under peak load.\n")
	sb.WriteString("- **Audit Trail**: Every administrative update must emit an immutable audit event with actor ID, timestamp, and IP address.\n")
	sb.WriteString("- **Data Security**: Secrets and credentials must be encrypted at rest with AES-256-GCM and in transit via TLS 1.3.\n\n")

	// Section 7: Out of scope
	sb.WriteString("## 7. Out of Scope Guardrails\n\n")
	sb.WriteString("- Legacy proprietary authentication protocols (LDAP over plain text, NTLM) will not be supported.\n")
	sb.WriteString("- Custom on-premises hardware security module (HSM) integrations are deferred to Phase 2.\n")
	sb.WriteString("- Real-time video or audio streaming audit feeds are excluded from the initial release.\n\n")

	// Section 8: Rollout plan and success metrics
	sb.WriteString("## 8. Rollout Plan and Success Metrics\n\n")
	sb.WriteString("- **Private Alpha (Week 1-2)**: Deploy with 2 high-ARR enterprise design partners (Cyberdyne Systems and Acme Global Tech).\n")
	sb.WriteString("- **Public Beta (Week 3-4)**: Enable self-service configuration for all Growth and Enterprise accounts.\n")
	sb.WriteString("- **General Availability (Week 5)**: Automated closed-loop email updates dispatched to all requesting customer accounts.\n\n")
	sb.WriteString("### Key Success Metrics\n")
	sb.WriteString("- **100% Renewal Retention**: Zero accounts citing this capability gap as a cancellation reason.\n")
	sb.WriteString("- **Adoption Rate**: > 75% of requesting accounts activate the capability within 30 days of release.\n")
	sb.WriteString("- **Support Ticket Volume**: < 2% of active users submit configuration assistance requests.\n")

	markdown := sb.String()
	notionJSON := ConvertMarkdownToNotionBlocks(opp.Title, markdown)

	return markdown, notionJSON, nil
}

func generateSSOUserStories() string {
	return `### Story 1: Enterprise Identity Provider Configuration
**As an** Enterprise IT Administrator,  
**I want to** configure our SAML 2.0 metadata URL and upload our X.509 certificate in the workspace settings,  
**So that** our corporate employees can authenticate through Okta or Azure AD.

` + "```gherkin" + `
Feature: SAML 2.0 Identity Provider Configuration

  Scenario: Successfully saving valid Okta metadata URL
    Given an administrator is logged into the workspace settings
    When they enter a valid Okta SAML metadata URL
    And they click "Test SAML Connection"
    Then the system validates the identity provider certificate
    And displays "Connection verified successfully"

  Scenario: Enforcing SSO for all company email domains
    Given SAML configuration is verified
    When the administrator checks "Enforce SSO for all @company.com accounts"
    Then direct password logins are disabled for users matching that domain
    And users are redirected to the corporate Okta sign-in page
` + "```\n\n"
}

func generateAuditUserStories() string {
	return `### Story 1: Streaming SIEM Audit Event Delivery
**As a** Corporate Compliance Officer,  
**I want to** subscribe our Datadog or Splunk HTTPS webhook endpoint to workspace security events,  
**So that** all administrator actions and permission changes are logged immutably.

` + "```gherkin" + `
Feature: SIEM Security Audit Webhook Streaming

  Scenario: Delivering real-time audit logs on role update
    Given a configured webhook endpoint with HMAC signature secret
    When an administrator changes an account permission from Member to Admin
    Then a signed JSON payload with event type "audit.permission.updated" is dispatched
    And includes actor ID, target user ID, timestamp, and client IP address
    And delivers within 2 seconds of the administrative action
` + "```\n\n"
}

func generateRBACUserStories() string {
	return `### Story 1: Custom Role Creation with Granular Scopes
**As an** Operations Lead,  
**I want to** define custom roles with restricted read-only permissions,  
**So that** external contractors can review analytics without editing billing or workspace configuration.

` + "```gherkin" + `
Feature: Granular Role-Based Access Control

  Scenario: Contractor attempts unauthorized billing update
    Given an external contractor assigned the "Auditor / Read-Only" custom role
    When they send a PATCH request to the billing payment method endpoint
    Then the system denies the request with HTTP 403 Forbidden
    And records a security access rejection event in the audit trail
` + "```\n\n"
}

func generateExportUserStories() string {
	return `### Story 1: Asynchronous Large Dataset Export API
**As a** Data Platform Engineer,  
**I want to** trigger an asynchronous batch data export for over 100,000 records,  
**So that** our analytics pipeline can extract warehouse updates without HTTP timeout errors.

` + "```gherkin" + `
Feature: Asynchronous Bulk Data Export

  Scenario: Initiating export job and receiving presigned download link
    Given an API client with export scopes
    When they post a job request to "/api/v1/exports" specifying a 1-year date filter
    Then the API returns HTTP 202 Accepted with a unique job ID
    And dispatches a completion webhook containing a signed S3 download URL upon completion
` + "```\n\n"
}

func generateDefaultUserStories(title string) string {
	return fmt.Sprintf(`### Story 1: Implementation of %s
**As an** Account Owner,  
**I want to** utilize this capability seamlessly within the product workflow,  
**So that** our team can complete our primary objectives efficiently.

`+"```gherkin"+`
Feature: %s Capability Verification

  Scenario: Standard user execution
    Given an authenticated customer account
    When the user activates the feature
    Then the action completes successfully with verifiable confirmation
`+"```\n\n", title, title)
}
