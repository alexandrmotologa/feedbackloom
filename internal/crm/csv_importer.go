package crm

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
	"github.com/google/uuid"
)

// ImportAccountsFromCSV parses customer account data from a CSV file and upserts into the store.
func ImportAccountsFromCSV(ctx context.Context, s *store.Store, filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open csv file: %w", err)
	}
	defer file.Close()

	return ImportAccountsFromReader(ctx, s, file)
}

// ImportAccountsFromReader parses customer account data from an io.Reader and upserts into the store.
func ImportAccountsFromReader(ctx context.Context, s *store.Store, r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}

	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Required: domain and either name or arr
	domainIdx, ok := headerMap["domain"]
	if !ok {
		return 0, fmt.Errorf("missing 'domain' column in csv")
	}

	nameIdx, hasName := headerMap["name"]
	arrIdx, hasARR := headerMap["arr_usd"]
	if !hasARR {
		arrIdx, hasARR = headerMap["arr"]
	}
	tierIdx, hasTier := headerMap["plan_tier"]
	if !hasTier {
		tierIdx, hasTier = headerMap["tier"]
	}
	renewalIdx, hasRenewal := headerMap["renewal_date"]
	if !hasRenewal {
		renewalIdx, hasRenewal = headerMap["renewal"]
	}
	churnIdx, hasChurn := headerMap["churn_risk_score"]
	if !hasChurn {
		churnIdx, hasChurn = headerMap["churn_risk"]
	}
	idIdx, hasID := headerMap["id"]

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("read csv record: %w", err)
		}

		domain := strings.TrimSpace(record[domainIdx])
		if domain == "" {
			continue
		}

		name := domain
		if hasName && nameIdx < len(record) && strings.TrimSpace(record[nameIdx]) != "" {
			name = strings.TrimSpace(record[nameIdx])
		}

		var arr float64 = 0.0
		if hasARR && arrIdx < len(record) {
			cleanVal := strings.ReplaceAll(record[arrIdx], "$", "")
			cleanVal = strings.ReplaceAll(cleanVal, ",", "")
			if v, err := strconv.ParseFloat(strings.TrimSpace(cleanVal), 64); err == nil {
				arr = v
			}
		}

		tier := models.PlanTierGrowth
		if hasTier && tierIdx < len(record) && strings.TrimSpace(record[tierIdx]) != "" {
			tier = strings.ToUpper(strings.TrimSpace(record[tierIdx]))
		} else {
			if arr >= 50000.0 {
				tier = models.PlanTierEnterprise
			} else if arr < 10000.0 {
				tier = models.PlanTierStarter
			}
		}

		renewalDate := time.Now().AddDate(0, 6, 0)
		if hasRenewal && renewalIdx < len(record) && strings.TrimSpace(record[renewalIdx]) != "" {
			val := strings.TrimSpace(record[renewalIdx])
			if t, err := time.Parse("2006-01-02", val); err == nil {
				renewalDate = t
			} else if t, err := time.Parse(time.RFC3339, val); err == nil {
				renewalDate = t
			}
		}

		churnRisk := 0.2
		if hasChurn && churnIdx < len(record) && strings.TrimSpace(record[churnIdx]) != "" {
			if v, err := strconv.ParseFloat(strings.TrimSpace(record[churnIdx]), 64); err == nil {
				churnRisk = v
			}
		}

		id := fmt.Sprintf("acc_%s", uuid.New().String()[:8])
		if hasID && idIdx < len(record) && strings.TrimSpace(record[idIdx]) != "" {
			id = strings.TrimSpace(record[idIdx])
		}

		acc := &models.Account{
			ID:             id,
			Name:           name,
			Domain:         domain,
			ARRUSD:         arr,
			PlanTier:       tier,
			RenewalDate:    renewalDate,
			ChurnRiskScore: churnRisk,
		}

		if err := s.UpsertAccount(ctx, acc); err != nil {
			return count, fmt.Errorf("upsert account %s: %w", domain, err)
		}
		count++
	}

	return count, nil
}
