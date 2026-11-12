package cluster

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/ingest"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

func TestVectorizer(t *testing.T) {
	vec := NewLocalSemanticVectorizer(128)
	ctx := context.Background()

	emb1, err := vec.Embed(ctx, "We require Okta SAML 2.0 single sign-on authentication")
	if err != nil {
		t.Fatalf("embed error: %v", err)
	}
	if len(emb1) != 128 {
		t.Fatalf("expected 128 dimensions, got %d", len(emb1))
	}

	// Verify L2 normalization
	var sumSq float64
	for _, val := range emb1 {
		sumSq += float64(val * val)
	}
	if math.Abs(sumSq-1.0) > 0.001 {
		t.Errorf("expected L2 norm 1.0, got %f", sumSq)
	}

	emb2, _ := vec.Embed(ctx, "Support for SAML SSO and Okta integration is mandatory")
	emb3, _ := vec.Embed(ctx, "Bulk export of orders and invoice CSV rows via REST API")

	distSimilar := CosineDistance(emb1, emb2)
	distDifferent := CosineDistance(emb1, emb3)

	if distSimilar >= distDifferent {
		t.Errorf("expected similar texts to have smaller distance: similar=%f, different=%f", distSimilar, distDifferent)
	}
}

func TestFeedbackExtractor(t *testing.T) {
	ext := NewFeedbackExtractor()
	ctx := context.Background()

	res1 := ext.Extract(ctx, "If Okta SAML SSO is not supported next month, our CISO will block renewal and cancel.")
	if res1.Category != models.CategorySecurity {
		t.Errorf("expected SECURITY category, got %s", res1.Category)
	}
	if res1.UrgencyScore < 0.9 {
		t.Errorf("expected high urgency for cancel/renewal threat, got %f", res1.UrgencyScore)
	}

	res2 := ext.Extract(ctx, "Can we export 50k rows without timeout errors in the CSV API?")
	if res2.Category != models.CategoryPerformance {
		t.Errorf("expected PERFORMANCE category, got %s", res2.Category)
	}
}

func TestHDBSCANClustering(t *testing.T) {
	cfg := DefaultHDBSCANConfig()
	cfg.MinPoints = 2
	cfg.MinClusterSize = 2
	hdb := NewHDBSCAN(cfg)

	// Create 2 distinct spatial clusters:
	// Cluster A around (0.1, 0.1)
	// Cluster B around (0.9, 0.9)
	// Outlier at (0.5, 0.0)
	data := [][]float32{
		{0.10, 0.10},
		{0.12, 0.11},
		{0.11, 0.09},
		{0.90, 0.90},
		{0.92, 0.89},
		{0.89, 0.91},
	}

	res := hdb.Fit(data)
	if res.NumClusters < 2 {
		t.Fatalf("expected at least 2 clusters, got %d", res.NumClusters)
	}
	if res.Labels[0] != res.Labels[1] || res.Labels[1] != res.Labels[2] {
		t.Errorf("first group should share cluster label: %+v", res.Labels[:3])
	}
	if res.Labels[3] != res.Labels[4] || res.Labels[4] != res.Labels[5] {
		t.Errorf("second group should share cluster label: %+v", res.Labels[3:6])
	}
	if res.Labels[0] == res.Labels[3] {
		t.Errorf("expected different clusters for group 1 and 2, got %d", res.Labels[0])
	}
}

func TestClusteringPipelineEndToEnd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "feedbackloom_pipeline_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "pipeline.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store new error: %v", err)
	}
	defer s.Close()

	ctx := context.Background()

	// 1. Seed CRM accounts
	_, err = crm.SyncStripeCustomers(ctx, s, crm.StripeSyncOptions{APIKey: "mock"})
	if err != nil {
		t.Fatalf("seed stripe error: %v", err)
	}

	resolver := crm.NewResolver(s)

	// 2. Seed feedback
	_, err = ingest.SeedRealisticFeedback(ctx, s, resolver)
	if err != nil {
		t.Fatalf("seed feedback error: %v", err)
	}

	// 3. Run Pipeline
	vec := NewLocalSemanticVectorizer(128)
	ext := NewFeedbackExtractor()
	pipeline := NewPipeline(s, vec, ext, resolver)

	opts := PipelineOptions{
		MinClusterSize: 2,
		MinPoints:      2,
	}

	summary, err := pipeline.Run(ctx, opts)
	if err != nil {
		t.Fatalf("pipeline run error: %v", err)
	}

	if summary.ClustersCreated < 2 {
		t.Fatalf("expected at least 2 clusters created, got %d", summary.ClustersCreated)
	}
	if summary.TotalDemandedARR <= 0 {
		t.Errorf("expected positive total demanded ARR, got %f", summary.TotalDemandedARR)
	}

	opps, err := s.ListOpportunities(ctx)
	if err != nil {
		t.Fatalf("list opportunities error: %v", err)
	}
	if len(opps) != summary.ClustersCreated {
		t.Errorf("expected %d opportunities, got %d", summary.ClustersCreated, len(opps))
	}

	// Verify top opportunity is weighted properly
	topOpp := opps[0]
	if topOpp.WeightedScore <= 0 || topOpp.TotalARRUSD <= 0 {
		t.Errorf("invalid metrics on top opportunity: %+v", topOpp)
	}

	// Verify linked feedback items
	linked, err := s.ListEnrichedFeedbackForOpportunity(ctx, topOpp.ID)
	if err != nil {
		t.Fatalf("list enriched error: %v", err)
	}
	if len(linked) < 2 {
		t.Errorf("expected at least 2 linked items for top opportunity, got %d", len(linked))
	}
}
