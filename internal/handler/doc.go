package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// DocHandler serves interactive Swagger UI and raw OpenAPI 3.1 schema definitions.
// Why: Provides built-in developer documentation, testing sandboxes, and API contract discovery directly from the microservice.
type DocHandler struct {
	yamlPath string
	jsonPath string
}

// NewDocHandler constructs a new DocHandler with target OpenAPI document file paths.
// Why: Injects filesystem locations for YAML and JSON OpenAPI specification files.
func NewDocHandler(yamlPath, jsonPath string) *DocHandler {
	if yamlPath == "" {
		yamlPath = "docs/openapi.yaml"
	}
	if jsonPath == "" {
		jsonPath = "docs/openapi.json"
	}
	return &DocHandler{
		yamlPath: yamlPath,
		jsonPath: jsonPath,
	}
}

// ServeOpenAPIYAML handles GET /docs/openapi.yaml.
// Why: Returns the OpenAPI specification in raw YAML format for CLI tooling and schema parsers.
func (h *DocHandler) ServeOpenAPIYAML(w http.ResponseWriter, r *http.Request) {
	data, err := h.readFile(h.yamlPath)
	if err != nil {
		http.Error(w, "OpenAPI YAML document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ServeOpenAPIJSON handles GET /docs/openapi.json.
// Why: Returns the OpenAPI specification in raw JSON format for client generation and automated testing.
func (h *DocHandler) ServeOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	data, err := h.readFile(h.jsonPath)
	if err != nil {
		http.Error(w, "OpenAPI JSON document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ServeSwaggerUI handles GET /docs and GET /swagger.
// Why: Renders an interactive Swagger UI web console configured to parse /docs/openapi.yaml.
func (h *DocHandler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Store User API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; padding: 0; background: #fafafa; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/docs/openapi.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

	_, _ = fmt.Fprint(w, html)
}

func (h *DocHandler) readFile(path string) ([]byte, error) {
	// Try relative or direct path
	if data, err := os.ReadFile(path); err == nil {
		return data, nil
	}

	// Try checking parent directories if running from test package directory
	candidates := []string{
		path,
		filepath.Join("..", path),
		filepath.Join("..", "..", path),
	}

	for _, c := range candidates {
		if data, err := os.ReadFile(c); err == nil {
			return data, nil
		}
	}

	return nil, os.ErrNotExist
}
