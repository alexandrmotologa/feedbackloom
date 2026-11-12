package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/feedbackloom/internal/config"
	"github.com/alexandrmotologa/feedbackloom/internal/models"
	"github.com/alexandrmotologa/feedbackloom/internal/store"
)

func setupAPITestServer(t *testing.T) (*Server, func()) {
	tmpDir, err := os.MkdirTemp("", "feedbackloom_api_test_*")
	if err != nil {
		t.Fatalf("temp dir error: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "api_test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store new error: %v", err)
	}

	cfg := &config.Config{
		Port:           "8080",
		DBPath:         dbPath,
		MinClusterSize: 2,
		MinPoints:      2,
	}

	srv := NewServer(cfg, s, nil)
	cleanup := func() {
		s.Close()
		os.RemoveAll(tmpDir)
	}

	return srv, cleanup
}

func TestAPIEndpoints(t *testing.T) {
	srv, cleanup := setupAPITestServer(t)
	defer cleanup()

	// 1. Health check
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /health, got %d", w.Code)
	}

	// 2. CRM sync
	req = httptest.NewRequest(http.MethodPost, "/api/v1/crm/sync?provider=stripe", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for crm sync, got %d", w.Code)
	}

	// 3. Seed feedback
	req = httptest.NewRequest(http.MethodPost, "/api/v1/feedback/seed", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for feedback seed, got %d", w.Code)
	}

	// 4. Cluster run
	req = httptest.NewRequest(http.MethodPost, "/api/v1/cluster/run", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for cluster run, got %d: %s", w.Code, w.Body.String())
	}

	var clusterSummary struct {
		ClustersCreated  int     `json:"clusters_created"`
		TotalDemandedARR float64 `json:"total_demanded_arr_usd"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &clusterSummary)
	if clusterSummary.ClustersCreated < 2 {
		t.Fatalf("expected >= 2 clusters, got %d", clusterSummary.ClustersCreated)
	}

	// 5. List opportunities
	req = httptest.NewRequest(http.MethodGet, "/api/v1/opportunities", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for list opportunities, got %d", w.Code)
	}

	var oppsResponse struct {
		Opportunities []models.Opportunity `json:"opportunities"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &oppsResponse)
	if len(oppsResponse.Opportunities) == 0 {
		t.Fatalf("expected opportunities, got 0")
	}

	oppID := oppsResponse.Opportunities[0].ID

	// 6. Generate PRD for top opportunity
	req = httptest.NewRequest(http.MethodPost, "/api/v1/opportunities/"+oppID+"/prd", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for PRD generation, got %d: %s", w.Code, w.Body.String())
	}

	var prdResponse struct {
		Markdown string `json:"markdown"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &prdResponse)
	if prdResponse.Markdown == "" {
		t.Errorf("expected non-empty PRD markdown")
	}

	// 7. Update status to SHIPPED
	statusBody := bytes.NewBufferString(`{"status":"SHIPPED"}`)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/opportunities/"+oppID+"/status", statusBody)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for status update, got %d", w.Code)
	}

	// 8. Generate and dispatch closed-loop notifications
	req = httptest.NewRequest(http.MethodPost, "/api/v1/opportunities/"+oppID+"/notifications/generate", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for notifications generate, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/opportunities/"+oppID+"/notifications/dispatch", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for notifications dispatch, got %d", w.Code)
	}
}
