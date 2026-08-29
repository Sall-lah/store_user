package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocHandler_ServeSwaggerUI(t *testing.T) {
	h := NewDocHandler("docs/openapi.yaml", "docs/openapi.json")

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	h.ServeSwaggerUI(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected text/html Content-Type, got: %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") || !strings.Contains(body, "./openapi.yaml") {
		t.Errorf("Swagger UI HTML does not contain required scripts or relative openapi URL: %s", body)
	}
}

func TestDocHandler_ServeOpenAPIYAML(t *testing.T) {
	h := NewDocHandler("docs/openapi.yaml", "docs/openapi.json")

	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	h.ServeOpenAPIYAML(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "yaml") {
		t.Errorf("expected yaml Content-Type, got: %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "openapi:") {
		t.Errorf("expected openapi YAML header in response body")
	}
}

func TestDocHandler_ServeOpenAPIJSON(t *testing.T) {
	h := NewDocHandler("docs/openapi.yaml", "docs/openapi.json")

	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	rec := httptest.NewRecorder()
	h.ServeOpenAPIJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected application/json Content-Type, got: %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"openapi"`) {
		t.Errorf("expected openapi JSON key in response body")
	}
}

func TestDocHandler_MissingFileReturns404(t *testing.T) {
	h := NewDocHandler("non_existent_file.yaml", "non_existent_file.json")

	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	h.ServeOpenAPIYAML(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got: %d", rec.Code)
	}

	reqJSON := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	recJSON := httptest.NewRecorder()
	h.ServeOpenAPIJSON(recJSON, reqJSON)

	if recJSON.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got: %d", recJSON.Code)
	}
}
