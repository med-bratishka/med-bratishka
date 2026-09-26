package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"medbratishka/models"

	"github.com/gorilla/mux"
)

func TestRouterNotFoundReturnsJSON(t *testing.T) {
	router := mux.NewRouter()
	ApplyRouterErrorHandlers(router, nil)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var resp models.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "ROUTE_NOT_FOUND" || resp.Type != "not_found" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRouterMethodNotAllowedReturnsJSON(t *testing.T) {
	router := mux.NewRouter()
	ApplyRouterErrorHandlers(router, nil)
	router.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {}).Methods(http.MethodPost)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/resource", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}

	var resp models.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "METHOD_NOT_ALLOWED" || resp.Type != "method_not_allowed" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRecoveryMiddlewareReturnsJSON(t *testing.T) {
	router := mux.NewRouter()
	router.Use(RecoveryMiddleware(nil))
	router.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var resp models.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "INTERNAL_ERROR" || resp.Type != "internal_error" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
