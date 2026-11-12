package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

// SeedRealisticFeedback populates the database with realistic multi-channel enterprise feedback.
func SeedRealisticFeedback(ctx context.Context, s *store.Store, resolver *crm.Resolver) (int, error) {
	now := time.Now()

	samples := []struct {
		ID            string
		Source        string
		Email         string
		Name          string
		RawText       string
		CreatedOffset time.Duration
	}{
		// Cluster 1: Enterprise SAML 2.0 & Okta SSO
		{
			ID:            "fb_sso_cyberdyne",
			Source:        models.SourceZendesk,
			Email:         "cto@cyberdyne.com",
			Name:          "Miles Dyson",
			RawText:       "We cannot proceed with our enterprise rollout without Okta SAML 2.0 single sign-on. Our corporate InfoSec mandate requires SSO enforcement by end of quarter, or we will be forced to cancel our renewal.",
			CreatedOffset: -48 * time.Hour,
		},
		{
			ID:            "fb_sso_acme",
			Source:        models.SourceGong,
			Email:         "security-director@acme.com",
			Name:          "Rachel Green",
			RawText:       "In our security review call today, SAML SSO with Azure AD and Okta was highlighted as a mandatory requirement. Your current username/password login violates our corporate security guidelines.",
			CreatedOffset: -36 * time.Hour,
		},
		{
			ID:            "fb_sso_massivedynamic",
			Source:        models.SourceIntercom,
			Email:         "nina.sharp@massivedynamic.org",
			Name:          "Nina Sharp",
			RawText:       "Is Okta SSO on your near-term roadmap? Our compliance team requires SCIM provisioning and SAML authentication before we can upgrade all 400 seats.",
			CreatedOffset: -24 * time.Hour,
		},
		{
			ID:            "fb_sso_hooli",
			Source:        models.SourceSlack,
			Email:         "gavin.belson@hooli.xyz",
			Name:          "Gavin Belson",
			RawText:       "We have a company-wide policy that all third-party SaaS vendors must integrate directly with our Okta identity provider. When will SAML 2.0 be supported?",
			CreatedOffset: -12 * time.Hour,
		},

		// Cluster 2: SIEM Audit Logs & Real-Time Security Webhooks
		{
			ID:            "fb_audit_wayne",
			Source:        models.SourceZendesk,
			Email:         "lucius.fox@waynesec.com",
			Name:          "Lucius Fox",
			RawText:       "Our SOC 2 compliance auditor requires all administrative actions, permission changes, and login attempts to be streamed directly to our Datadog and Splunk SIEM via automated audit webhooks.",
			CreatedOffset: -72 * time.Hour,
		},
		{
			ID:            "fb_audit_stark",
			Source:        models.SourceGong,
			Email:         "jarvis-admin@stark.tech",
			Name:          "Pepper Potts",
			RawText:       "During the vendor security assessment, our CISO flagged that you do not provide immutable audit log exports. We need streaming audit events for SOC 2 Type II compliance.",
			CreatedOffset: -50 * time.Hour,
		},
		{
			ID:            "fb_audit_globex",
			Source:        models.SourceIntercom,
			Email:         "hank.scorpio@globex.corp",
			Name:          "Hank Scorpio",
			RawText:       "We urgently need an automated audit log feed to satisfy our annual security audit. Can you stream login timestamps, IP addresses, and policy changes to our security lake?",
			CreatedOffset: -18 * time.Hour,
		},

		// Cluster 3: Granular Role-Based Access Control (RBAC) & Custom Scopes
		{
			ID:            "fb_rbac_soylent",
			Source:        models.SourceIntercom,
			Email:         "compliance@soylent.io",
			Name:          "William Thorn",
			RawText:       "The current admin vs member roles are too broad. We need granular role-based access control where contractors can view reports without having export or billing edit permissions.",
			CreatedOffset: -60 * time.Hour,
		},
		{
			ID:            "fb_rbac_initech",
			Source:        models.SourceSlack,
			Email:         "peter.gibbons@initech.com",
			Name:          "Peter Gibbons",
			RawText:       "We have department managers who need to approve workflows without having full administrative rights. Custom permission scopes would solve this completely.",
			CreatedOffset: -30 * time.Hour,
		},
		{
			ID:            "fb_rbac_umbrella",
			Source:        models.SourceZendesk,
			Email:         "albert.wesker@umbrella.bio",
			Name:          "Albert Wesker",
			RawText:       "We need read-only auditor roles for our clinical compliance inspectors so they can review historical records without altering any configuration.",
			CreatedOffset: -15 * time.Hour,
		},

		// Cluster 4: High-Performance Bulk Data Export API
		{
			ID:            "fb_export_piedpiper",
			Source:        models.SourceSlack,
			Email:         "richard.hendricks@piedpiper.com",
			Name:          "Richard Hendricks",
			RawText:       "When we try to export more than 50,000 records via your REST API, the connection times out after 30 seconds. We need an asynchronous bulk export API with S3 presigned download URLs.",
			CreatedOffset: -80 * time.Hour,
		},
		{
			ID:            "fb_export_dunder",
			Source:        models.SourceZendesk,
			Email:         "michael.scott@dundermifflin.com",
			Name:          "Michael Scott",
			RawText:       "The CSV export fails whenever we select the full year of order data. Can we get scheduled weekly CSV exports sent automatically to email or an SFTP folder?",
			CreatedOffset: -40 * time.Hour,
		},
		{
			ID:            "fb_export_initech",
			Source:        models.SourceIntercom,
			Email:         "bill.lumbergh@initech.com",
			Name:          "Bill Lumbergh",
			RawText:       "Yeah, if you could just provide an automated API endpoint for bulk batch exports instead of pagination limits, that would be great. Our analytics warehouse needs daily syncs.",
			CreatedOffset: -10 * time.Hour,
		},
	}

	count := 0
	for _, sRow := range samples {
		accountID := ""
		if resolver != nil && sRow.Email != "" {
			if acc, _ := resolver.ResolveAccount(ctx, sRow.Email); acc != nil {
				accountID = acc.ID
			}
		}

		item := &models.FeedbackItem{
			ID:            sRow.ID,
			Source:        sRow.Source,
			AccountID:     accountID,
			CustomerEmail: sRow.Email,
			CustomerName:  sRow.Name,
			RawText:       sRow.RawText,
			CreatedAt:     now.Add(sRow.CreatedOffset),
		}

		if err := s.InsertFeedbackItem(ctx, item); err != nil {
			return count, fmt.Errorf("seed feedback %s: %w", sRow.ID, err)
		}
		count++
	}

	return count, nil
}
