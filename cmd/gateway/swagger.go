package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SwaggerUIHTML returns the HTML content for Swagger UI
func getSwaggerUIHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="description" content="SwaggerUI for ETC Meisai API" />
    <title>ETC Meisai API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.css" />
    <style>
        body {
            margin: 0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #2563eb;
        }
        .swagger-ui .topbar .download-url-wrapper {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js" crossorigin></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                url: '/swagger/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.presets.standalone,
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true,
                tryItOutEnabled: true,
                requestInterceptor: (request) => {
                    // Add any custom headers or modifications here
                    console.log('API Request:', request);
                    return request;
                },
                responseInterceptor: (response) => {
                    console.log('API Response:', response);
                    return response;
                }
            });
        };
    </script>
</body>
</html>`
}

// setupSwaggerRoutes configures the Swagger UI and documentation endpoints
func setupSwaggerRoutes(mux *http.ServeMux) {
	// Swagger UI endpoint
	mux.HandleFunc("/swagger-ui/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getSwaggerUIHTML()))
	})

	// Swagger JSON endpoint
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract the file path from the URL
		path := strings.TrimPrefix(r.URL.Path, "/swagger/")
		if path == "" {
			http.Error(w, "File not specified", http.StatusBadRequest)
			return
		}

		// Serve swagger files from the swagger directory
		serveSwaggerFile(w, r, path)
	})

	// API docs redirect endpoint
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger-ui/", http.StatusMovedPermanently)
	})

	// API docs with trailing slash
	mux.HandleFunc("/docs/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger-ui/", http.StatusMovedPermanently)
	})

	// OpenAPI specification endpoint
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/swagger.json", http.StatusMovedPermanently)
	})
}

// serveSwaggerFile serves swagger files from the swagger directory
func serveSwaggerFile(w http.ResponseWriter, r *http.Request, filename string) {
	// Security: only allow certain file types
	allowedExtensions := map[string]string{
		".json": "application/json",
		".yaml": "application/yaml",
		".yml":  "application/yaml",
	}

	ext := filepath.Ext(filename)
	contentType, allowed := allowedExtensions[ext]
	if !allowed {
		http.Error(w, "File type not allowed", http.StatusForbidden)
		return
	}

	// Construct the file path (relative to project root)
	swaggerDir := getSwaggerDirectory()
	filePath := filepath.Join(swaggerDir, filename)

	// Security: prevent path traversal
	if !strings.HasPrefix(filePath, swaggerDir) {
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// If swagger.json doesn't exist, generate a default one
		if filename == "swagger.json" {
			serveDefaultSwaggerSpec(w)
			return
		}
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Read and serve the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// getSwaggerDirectory returns the path to the swagger directory
func getSwaggerDirectory() string {
	// Try to find swagger directory relative to current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return "./swagger"
	}

	swaggerDir := filepath.Join(workDir, "swagger")
	if _, err := os.Stat(swaggerDir); os.IsNotExist(err) {
		// If not found in current directory, try relative to executable
		if execDir := getExecutableDir(); execDir != "" {
			swaggerDir = filepath.Join(execDir, "swagger")
		}
	}

	return swaggerDir
}

// getExecutableDir returns the directory containing the executable
func getExecutableDir() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(execPath)
}

// serveDefaultSwaggerSpec serves a default OpenAPI specification
func serveDefaultSwaggerSpec(w http.ResponseWriter) {
	defaultSpec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "ETC Meisai API",
			"description": "API for ETC toll statement management system",
			"version":     "0.0.19",
			"contact": map[string]interface{}{
				"name": "ETC Meisai API Support",
			},
		},
		"servers": []map[string]interface{}{
			{
				"url":         "/",
				"description": "Current server",
			},
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check endpoint",
					"description": "Returns the health status of the service",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status": map[string]interface{}{
												"type":    "string",
												"example": "healthy",
											},
											"service": map[string]interface{}{
												"type":    "string",
												"example": "etc-meisai-gateway",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"HealthResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Health status",
						},
						"service": map[string]interface{}{
							"type":        "string",
							"description": "Service name",
						},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(defaultSpec); err != nil {
		http.Error(w, "Failed to generate swagger specification", http.StatusInternalServerError)
		return
	}
}