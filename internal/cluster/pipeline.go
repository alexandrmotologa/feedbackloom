package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/ranking"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
	"github.com/google/uuid"
)

// PipelineOptions controls the clustering execution.
type PipelineOptions struct {
	MinClusterSize int
	MinPoints      int
	DistanceMetric string
}

// PipelineSummary summarizes the execution of a clustering run.
type PipelineSummary struct {
	TotalProcessed   int     `json:"total_processed"`
	ClustersCreated  int     `json:"clusters_created"`
	NoiseItemCount   int     `json:"noise_item_count"`
	TotalDemandedARR float64 `json:"total_demanded_arr_usd"`
	DurationMS       int64   `json:"duration_ms"`
}

// Pipeline coordinates extraction, embedding, density clustering, and opportunity ranking.
type Pipeline struct {
	store      *store.Store
	vectorizer Vectorizer
	extractor  *FeedbackExtractor
	resolver   *crm.Resolver
}

// NewPipeline initializes a new clustering pipeline.
func NewPipeline(s *store.Store, vec Vectorizer, ext *FeedbackExtractor, res *crm.Resolver) *Pipeline {
	if vec == nil {
		vec = NewLocalSemanticVectorizer(128)
	}
	if ext == nil {
		ext = NewFeedbackExtractor()
	}
	return &Pipeline{
		store:      s,
		vectorizer: vec,
		extractor:  ext,
		resolver:   res,
	}
}

// Run executes the complete clustering and opportunity synthesis process.
func (p *Pipeline) Run(ctx context.Context, opts PipelineOptions) (*PipelineSummary, error) {
	start := time.Now()

	// 1. Fetch all feedback items
	items, err := p.store.ListAllFeedbackItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("list feedback items: %w", err)
	}
	if len(items) == 0 {
		return &PipelineSummary{}, nil
	}

	// 2. Extract pain points & outcomes if missing
	for i := range items {
		it := &items[i]
		if it.ExtractedPainPoint == "" || it.UrgencyScore == 0.0 || len(it.CompetitorMentions) == 0 {
			res := p.extractor.Extract(ctx, it.RawText)
			it.ExtractedPainPoint = res.PainPoint
			it.ExtractedOutcome = res.DesiredOutcome
			it.SentimentScore = res.SentimentScore
			it.UrgencyScore = res.UrgencyScore
			it.CompetitorMentions = res.CompetitorMentions
			_ = p.store.UpdateFeedbackExtraction(ctx, it.ID, res.PainPoint, res.DesiredOutcome, res.SentimentScore, res.UrgencyScore, res.CompetitorMentions)
		}

		// Ensure account ID is mapped if missing
		if it.AccountID == "" && p.resolver != nil && it.CustomerEmail != "" {
			if acc, _ := p.resolver.ResolveAccount(ctx, it.CustomerEmail); acc != nil {
				it.AccountID = acc.ID
				_ = p.store.UpdateFeedbackAccount(ctx, it.ID, acc.ID)
			}
		}

		// 3. Generate embeddings if missing
		if len(it.Embedding) == 0 {
			textToEmbed := it.RawText + " " + it.ExtractedPainPoint + " " + it.ExtractedOutcome
			emb, err := p.vectorizer.Embed(ctx, textToEmbed)
			if err == nil {
				it.Embedding = emb
				_ = p.store.UpdateFeedbackEmbedding(ctx, it.ID, emb)
			}
		}
	}

	// 4. Prepare data matrix for HDBSCAN
	vectors := make([][]float32, len(items))
	for i, it := range items {
		vectors[i] = it.Embedding
	}

	cfg := DefaultHDBSCANConfig()
	if opts.MinClusterSize >= 2 {
		cfg.MinClusterSize = opts.MinClusterSize
	}
	if opts.MinPoints >= 2 {
		cfg.MinPoints = opts.MinPoints
	}
	if opts.DistanceMetric != "" {
		cfg.DistanceMetric = opts.DistanceMetric
	}

	hdb := NewHDBSCAN(cfg)
	clusterRes := hdb.Fit(vectors)

	// 5. Group feedback items by cluster label
	clusterGroups := make(map[int][]models.FeedbackItem)
	noiseCount := 0
	for i, label := range clusterRes.Labels {
		if label < 0 {
			noiseCount++
			continue
		}
		clusterGroups[label] = append(clusterGroups[label], items[i])
	}

	// 6. Clear old opportunities before updating with newly synthesized clusters
	_ = p.store.ClearAllOpportunities(ctx)

	var totalDemandedARR float64
	createdCount := 0

	for _, group := range clusterGroups {
		if len(group) == 0 {
			continue
		}

		// Enrich each feedback item in the group with account data
		var enrichedGroup []models.EnrichedFeedbackItem
		for _, fb := range group {
			enriched := models.EnrichedFeedbackItem{
				FeedbackItem: fb,
			}
			if fb.AccountID != "" {
				if acc, err := p.store.GetAccountByID(ctx, fb.AccountID); err == nil && acc != nil {
					enriched.AccountName = acc.Name
					enriched.AccountDomain = acc.Domain
					enriched.AccountARRUSD = acc.ARRUSD
					enriched.PlanTier = acc.PlanTier
					enriched.DaysToRenewal = acc.DaysToRenewal()
					enriched.ChurnRiskScore = acc.ChurnRiskScore
				}
			}
			enrichedGroup = append(enrichedGroup, enriched)
		}

		// Synthesize theme title, description, and category
		theme := SynthesizeOpportunityTheme(enrichedGroup)

		// Calculate ARR metrics and Weighted Revenue Score (WRS)
		metrics := ranking.CalculateOpportunityMetrics(enrichedGroup)

		oppID := fmt.Sprintf("opp_%s", uuid.New().String()[:8])
		opp := &models.Opportunity{
			ID:                 oppID,
			Title:              theme.Title,
			Description:        theme.Description,
			Category:           theme.Category,
			TotalARRUSD:        metrics.TotalARRUSD,
			WeightedScore:      metrics.WeightedScore,
			AccountCount:       metrics.AccountCount,
			FeedbackCount:      metrics.FeedbackCount,
			Status:             models.StatusDiscovered,
			CompetitorMentions: metrics.CompetitorMentions,
			LostPipelineARR:    metrics.LostPipelineARR,
			EffortSprints:      metrics.EffortSprints,
			ROIIndex:           metrics.ROIIndex,
			UrgentRenewalAlert: metrics.UrgentRenewalAlert,
			VelocityGrowthPct:  metrics.VelocityGrowthPct,
		}

		if err := p.store.UpsertOpportunity(ctx, opp); err != nil {
			return nil, fmt.Errorf("upsert opportunity: %w", err)
		}

		// Link all items in the group to the opportunity
		for _, fb := range group {
			_ = p.store.LinkFeedbackOpportunity(ctx, fb.ID, oppID, 0.95)
		}

		totalDemandedARR += metrics.TotalARRUSD
		createdCount++
	}

	elapsed := time.Since(start)

	return &PipelineSummary{
		TotalProcessed:   len(items),
		ClustersCreated:  createdCount,
		NoiseItemCount:   noiseCount,
		TotalDemandedARR: totalDemandedARR,
		DurationMS:       elapsed.Milliseconds(),
	}, nil
}
