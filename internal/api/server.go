package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/feedbackloom/internal/closedloop"
	"github.com/alexandrmotologa/feedbackloom/internal/cluster"
	"github.com/alexandrmotologa/feedbackloom/internal/config"
	"github.com/alexandrmotologa/feedbackloom/internal/crm"
	"github.com/alexandrmotologa/feedbackloom/internal/ingest"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/prd"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
	"github.com/google/uuid"
)

// Server coordinates the HTTP REST API and embedded web studio.
type Server struct {
	cfg        *config.Config
	store      *store.Store
	resolver   *crm.Resolver
	pipeline   *cluster.Pipeline
	prdGen     *prd.PRDGenerator
	mux        *http.ServeMux
	fileServer http.Handler
}

// NewServer sets up routing and handlers.
func NewServer(cfg *config.Config, s *store.Store, webFS http.FileSystem) *Server {
	resolver := crm.NewResolver(s)
	vec := cluster.NewLocalSemanticVectorizer(128)
	ext := cluster.NewFeedbackExtractor()
	pipeline := cluster.NewPipeline(s, vec, ext, resolver)
	prdGen := prd.NewPRDGenerator()

	srv := &Server{
		cfg:      cfg,
		store:    s,
		resolver: resolver,
		pipeline: pipeline,
		prdGen:   prdGen,
		mux:      http.NewServeMux(),
	}

	if webFS != nil {
		srv.fileServer = http.FileServer(webFS)
	}

	srv.registerRoutes()
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS Headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	// Health & Stats
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/stats", s.handleStats)

	// CRM & Accounts
	s.mux.HandleFunc("/api/v1/accounts", s.handleAccounts)
	s.mux.HandleFunc("/api/v1/crm/sync", s.handleCRMSync)

	// Ingestion & Feedback
	s.mux.HandleFunc("/api/v1/feedback", s.handleFeedback)
	s.mux.HandleFunc("/api/v1/feedback/seed", s.handleSeedFeedback)
	s.mux.HandleFunc("/api/v1/ingest/csv", s.handleIngestCSV)
	s.mux.HandleFunc("/api/v1/ingest/json", s.handleIngestJSON)
	s.mux.HandleFunc("/api/v1/ingest/zendesk", s.handleIngestZendesk)
	s.mux.HandleFunc("/api/v1/ingest/intercom", s.handleIngestIntercom)
	s.mux.HandleFunc("/api/v1/ingest/gong", s.handleIngestGong)
	s.mux.HandleFunc("/api/v1/ingest/slack", s.handleIngestSlack)

	// Clustering & Opportunities
	s.mux.HandleFunc("/api/v1/cluster/run", s.handleClusterRun)
	s.mux.HandleFunc("/api/v1/opportunities", s.handleOpportunities)
	s.mux.HandleFunc("/api/v1/opportunities/", s.handleOpportunityDetail)

	// Frontend SPA fallback
	if s.fileServer != nil {
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			s.fileServer.ServeHTTP(w, r)
		})
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "feedbackloom",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.GetDashboardStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.store.ListAccounts(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"accounts": accounts,
		"count":    len(accounts),
	})
}

func (s *Server) handleCRMSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "stripe"
	}

	var count int
	var err error

	if provider == "hubspot" {
		count, err = crm.SyncHubSpotCompanies(r.Context(), s.store, crm.HubSpotSyncOptions{
			AccessToken: s.cfg.HubSpotKey,
		})
	} else {
		count, err = crm.SyncStripeCustomers(r.Context(), s.store, crm.StripeSyncOptions{
			APIKey: s.cfg.StripeKey,
		})
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"synced":   count,
		"provider": provider,
		"message":  fmt.Sprintf("Successfully synced %d customer accounts from %s", count, provider),
	})
}

func (s *Server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodGet {
		items, err := s.store.ListFeedbackItems(ctx, 100, 0)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"items": items,
			"count": len(items),
		})
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Source        string `json:"source"`
			CustomerEmail string `json:"customer_email"`
			CustomerName  string `json:"customer_name"`
			RawText       string `json:"raw_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.RawText == "" {
			respondError(w, http.StatusBadRequest, "raw_text is required")
			return
		}

		accountID := ""
		if req.CustomerEmail != "" {
			if acc, _ := s.resolver.ResolveAccount(ctx, req.CustomerEmail); acc != nil {
				accountID = acc.ID
			}
		}

		source := req.Source
		if source == "" {
			source = models.SourceAPI
		}

		item := &models.FeedbackItem{
			ID:            fmt.Sprintf("fb_%s", uuid.New().String()[:8]),
			Source:        strings.ToUpper(source),
			AccountID:     accountID,
			CustomerEmail: req.CustomerEmail,
			CustomerName:  req.CustomerName,
			RawText:       req.RawText,
			CreatedAt:     time.Now().UTC(),
		}

		if err := s.store.InsertFeedbackItem(ctx, item); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, item)
		return
	}

	respondError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleSeedFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()

	// Ensure accounts exist first
	_, _ = crm.SyncStripeCustomers(ctx, s.store, crm.StripeSyncOptions{APIKey: "mock"})

	count, err := ingest.SeedRealisticFeedback(ctx, s.store, s.resolver)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"seeded":  count,
		"message": fmt.Sprintf("Seeded %d multi-channel enterprise feedback items", count),
	})
}

func (s *Server) handleIngestCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	source := r.URL.Query().Get("source")
	if source == "" {
		source = models.SourceCSV
	}

	count, err := ingest.ImportFeedbackFromReader(r.Context(), s.store, s.resolver, r.Body, source)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"imported": count,
		"source":   source,
	})
}

func (s *Server) handleIngestJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	source := r.URL.Query().Get("source")
	if source == "" {
		source = models.SourceAPI
	}

	count, err := ingest.ImportFeedbackFromJSONReader(r.Context(), s.store, s.resolver, r.Body, source)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"imported": count,
		"source":   source,
	})
}

func (s *Server) handleIngestZendesk(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	item, err := ingest.IngestZendeskPayload(r.Context(), s.store, s.resolver, body)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleIngestIntercom(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	item, err := ingest.IngestIntercomPayload(r.Context(), s.store, s.resolver, body)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleIngestGong(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	items, err := ingest.IngestGongCall(r.Context(), s.store, s.resolver, body)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (s *Server) handleIngestSlack(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	// Handle Slack URL verification challenge if present
	var challengeReq struct {
		Challenge string `json:"challenge"`
		Type      string `json:"type"`
	}
	if err := json.Unmarshal(body, &challengeReq); err == nil && challengeReq.Type == "url_verification" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"challenge": challengeReq.Challenge})
		return
	}

	item, err := ingest.IngestSlackPayload(r.Context(), s.store, s.resolver, body)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleClusterRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	opts := cluster.PipelineOptions{
		MinClusterSize: s.cfg.MinClusterSize,
		MinPoints:      s.cfg.MinPoints,
	}

	summary, err := s.pipeline.Run(r.Context(), opts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (s *Server) handleOpportunities(w http.ResponseWriter, r *http.Request) {
	opps, err := s.store.ListOpportunities(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"opportunities": opps,
		"count":         len(opps),
	})
}

func (s *Server) handleOpportunityDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/opportunities/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	oppID := parts[0]
	ctx := r.Context()

	// Direct GET /api/v1/opportunities/{id}
	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			opp, err := s.store.GetOpportunityByID(ctx, oppID)
			if err != nil {
				respondError(w, http.StatusNotFound, "opportunity not found")
				return
			}
			quotes, _ := s.store.ListEnrichedFeedbackForOpportunity(ctx, oppID)
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"opportunity": opp,
				"quotes":      quotes,
			})
			return
		}
		if r.Method == http.MethodDelete {
			_ = s.store.DeleteOpportunity(ctx, oppID)
			respondJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
			return
		}
	}

	// Sub-actions: /prd, /export-notion, /status, /notifications
	action := parts[1]
	switch action {
	case "prd":
		if r.Method == http.MethodPost {
			opp, err := s.store.GetOpportunityByID(ctx, oppID)
			if err != nil {
				respondError(w, http.StatusNotFound, "opportunity not found")
				return
			}
			quotes, err := s.store.ListEnrichedFeedbackForOpportunity(ctx, oppID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			md, notionJSON, err := s.prdGen.GeneratePRD(ctx, opp, quotes)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			_ = s.store.UpdateOpportunityPRD(ctx, oppID, md, notionJSON)
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"markdown":    md,
				"notion_json": notionJSON,
			})
			return
		}

	case "export-notion":
		if r.Method == http.MethodPost {
			opp, err := s.store.GetOpportunityByID(ctx, oppID)
			if err != nil || opp.GeneratedPRDNotionJSON == "" {
				respondError(w, http.StatusBadRequest, "PRD must be generated before exporting to Notion")
				return
			}

			token := s.cfg.NotionKey
			if token == "" {
				token = "mock"
			}
			pageID := s.cfg.NotionPageID
			if pageID == "" {
				pageID = "page_default"
			}

			url, err := prd.ExportToNotionAPI(ctx, opp.Title, opp.GeneratedPRDNotionJSON, token, pageID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			respondJSON(w, http.StatusOK, map[string]string{
				"url":     url,
				"message": "Successfully published PRD to Notion",
			})
			return
		}

	case "status":
		if r.Method == http.MethodPatch {
			var body struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				respondError(w, http.StatusBadRequest, "invalid status payload")
				return
			}
			if err := s.store.UpdateOpportunityStatus(ctx, oppID, body.Status); err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]string{"status": body.Status})
			return
		}

	case "notifications":
		clService := closedloop.NewService(s.store)
		if len(parts) == 2 && r.Method == http.MethodGet {
			notifs, _ := s.store.ListNotificationsByOpportunity(ctx, oppID)
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"notifications": notifs,
				"count":         len(notifs),
			})
			return
		}
		if len(parts) == 3 && parts[2] == "generate" && r.Method == http.MethodPost {
			notifs, err := clService.GenerateNotifications(ctx, oppID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"generated":     len(notifs),
				"notifications": notifs,
			})
			return
		}
		if len(parts) == 3 && parts[2] == "dispatch" && r.Method == http.MethodPost {
			sent, err := clService.DispatchAllPending(ctx, oppID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"sent":    sent,
				"message": fmt.Sprintf("Dispatched %d personalized customer notifications", sent),
			})
			return
		}
	}

	http.NotFound(w, r)
}

func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}
