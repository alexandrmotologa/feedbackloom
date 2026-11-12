package crm

import (
	"context"
	"net/mail"
	"strings"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

// GenericEmailDomains defines free or consumer webmail domains that should not be treated as enterprise company domains.
var GenericEmailDomains = map[string]bool{
	"gmail.com":        true,
	"yahoo.com":        true,
	"hotmail.com":      true,
	"outlook.com":      true,
	"live.com":         true,
	"msn.com":          true,
	"icloud.com":       true,
	"me.com":           true,
	"mac.com":          true,
	"aol.com":          true,
	"proton.me":        true,
	"protonmail.com":   true,
	"zoho.com":         true,
	"yandex.com":       true,
	"mail.com":         true,
	"gmx.com":          true,
	"fastmail.com":     true,
	"googlemail.com":   true,
}

// ExtractDomain extracts and normalizes the domain from an email address.
func ExtractDomain(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}

	// Parse in case formatted as "Name <email@domain.com>"
	if parsed, err := mail.ParseAddress(email); err == nil {
		email = parsed.Address
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	domain = strings.TrimPrefix(domain, "www.")

	// Check if this domain is a generic webmail provider.
	if GenericEmailDomains[domain] {
		return ""
	}

	// Handle corporate subdomains (e.g. "corp.cyberdyne.com" -> "cyberdyne.com")
	domainParts := strings.Split(domain, ".")
	if len(domainParts) > 2 {
		// Keep last 2 parts unless it is a country code TLD like .co.uk
		lastPart := domainParts[len(domainParts)-1]
		secondLast := domainParts[len(domainParts)-2]
		if (lastPart == "uk" || lastPart == "au" || lastPart == "jp" || lastPart == "br" || lastPart == "nz") && len(secondLast) <= 3 {
			if len(domainParts) >= 3 {
				domain = strings.Join(domainParts[len(domainParts)-3:], ".")
			}
		} else {
			domain = strings.Join(domainParts[len(domainParts)-2:], ".")
		}
	}

	return domain
}

// Resolver resolves email addresses to registered customer accounts in the store.
type Resolver struct {
	store *store.Store
}

// NewResolver initializes a CRM email domain resolver.
func NewResolver(s *store.Store) *Resolver {
	return &Resolver{store: s}
}

// ResolveAccount matches an email address to a customer account.
func (r *Resolver) ResolveAccount(ctx context.Context, email string) (*models.Account, error) {
	domain := ExtractDomain(email)
	if domain == "" {
		return nil, nil
	}

	acc, err := r.store.GetAccountByDomain(ctx, domain)
	if err != nil {
		return nil, nil // Not found or error
	}
	return acc, nil
}
