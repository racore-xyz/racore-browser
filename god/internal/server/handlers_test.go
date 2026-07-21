package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/racore/god/pkg/api"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()
	dir, err := os.MkdirTemp("", "god-server-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfg := api.Config{
		Port:             47991,
		DataDir:          dir,
		NodeName:         "Test",
		MeshEnabled:      false,
		MeshHeartbeatSec: 300,
	}
	return New(cfg)
}

func TestHealthHandler(t *testing.T) {
	s := setupTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	s.healthHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.OK {
		t.Fatal("expected ok")
	}
	if resp.Version != "0.2.0" {
		t.Fatalf("expected 0.2.0, got %s", resp.Version)
	}
}

func TestProvidersHandler(t *testing.T) {
	s := setupTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/providers", nil)
	s.providersHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var providers []api.Provider
	if err := json.NewDecoder(w.Body).Decode(&providers); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(providers) != 9 {
		t.Fatalf("expected 9 providers, got %d", len(providers))
	}
}

func TestMeshStatusHandler(t *testing.T) {
	s := setupTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/mesh/status", nil)
	s.meshStatusHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMeshPeersHandler(t *testing.T) {
	s := setupTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/mesh/peers", nil)
	s.meshPeersHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDomainsHandlerList(t *testing.T) {
	s := setupTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/authority/domains", nil)
	s.domainsHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDomainsHandlerCreate(t *testing.T) {
	s := setupTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/authority/domains", strings.NewReader(`{"domain":"test-handler.com"}`))
	r.Header.Set("Content-Type", "application/json")
	s.domainsHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["domain"] != "test-handler.com" {
		t.Fatalf("expected test-handler.com, got %v", result["domain"])
	}
}

func TestDomainAvailableHandler(t *testing.T) {
	s := setupTestServer(t)

	// First create a domain
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/v1/authority/domains", strings.NewReader(`{"domain":"existing.com"}`))
	r1.Header.Set("Content-Type", "application/json")
	s.domainsHandler(w1, r1)

	// Check available
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/v1/authority/domains/existing.com/available", nil)
	s.domainByIDHandler(w2, r2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var avail map[string]any
	json.NewDecoder(w2.Body).Decode(&avail)
	if a, _ := avail["available"].(bool); a {
		t.Fatal("existing.com should not be available")
	}

	// Check free domain
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("GET", "/v1/authority/domains/free.com/available", nil)
	s.domainByIDHandler(w3, r3)

	var avail2 map[string]any
	json.NewDecoder(w3.Body).Decode(&avail2)
	if a, _ := avail2["available"].(bool); !a {
		t.Fatal("free.com should be available")
	}
}

func TestCorsMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	tests := []struct {
		origin string
		allow  bool
	}{
		{"http://localhost:3000", true},
		{"http://127.0.0.1:3000", true},
		{"https://racore.xyz", true},
		{"http://evil.com", false},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Origin", tt.origin)
		handler.ServeHTTP(w, r)

		header := w.Header().Get("Access-Control-Allow-Origin")
		if tt.allow && header == "" {
			t.Errorf("expected CORS header for origin %s", tt.origin)
		}
		if !tt.allow && header != "" {
			t.Errorf("unexpected CORS header for origin %s", tt.origin)
		}
	}
}

func TestCorsPreflight(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/", nil)
	r.Header.Set("Origin", "http://localhost:3000")
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", w.Code)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, 201, map[string]string{"status": "created"})

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Fatal("expected JSON content type")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, 404, "not found")

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["detail"] != "not found" {
		t.Fatalf("expected 'not found', got '%s'", result["detail"])
	}
}


