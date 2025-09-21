//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	HTTPGatewayBaseURL = "http://localhost:8080"
	SwaggerUIURL       = HTTPGatewayBaseURL + "/swagger-ui/"
	SwaggerJSONURL     = HTTPGatewayBaseURL + "/swagger.json"
	APIBaseURL         = HTTPGatewayBaseURL + "/api/v1/etc-meisai"
)

// TestSwaggerUIAvailability tests Swagger UI and HTTP gateway availability
// This integration test verifies:
// 1. Check Swagger UI endpoint availability
// 2. Verify swagger.json is served
// 3. Test HTTP gateway endpoints
// 4. Verify CORS headers
// 5. Check API documentation completeness
func TestSwaggerUIAvailability(t *testing.T) {
	// Setup HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("SwaggerUIEndpoint", func(t *testing.T) {
		resp, err := httpClient.Get(SwaggerUIURL)
		if err != nil {
			t.Fatalf("Failed to access Swagger UI: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 for Swagger UI, got %d", resp.StatusCode)
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read Swagger UI response: %v", err)
		}

		bodyStr := string(body)

		// Verify it's actually Swagger UI
		if !strings.Contains(bodyStr, "swagger") && !strings.Contains(bodyStr, "Swagger") {
			t.Error("Response doesn't appear to be Swagger UI")
		}

		// Check for common Swagger UI elements
		expectedElements := []string{
			"swagger-ui",
			"api-docs",
		}

		for _, element := range expectedElements {
			if !strings.Contains(strings.ToLower(bodyStr), strings.ToLower(element)) {
				t.Errorf("Expected to find '%s' in Swagger UI response", element)
			}
		}

		t.Log("Successfully accessed Swagger UI")
	})

	t.Run("SwaggerJSONEndpoint", func(t *testing.T) {
		resp, err := httpClient.Get(SwaggerJSONURL)
		if err != nil {
			t.Fatalf("Failed to access swagger.json: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 for swagger.json, got %d", resp.StatusCode)
		}

		// Verify content type
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected JSON content type, got %s", contentType)
		}

		// Parse JSON response
		var swaggerDoc map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&swaggerDoc)
		if err != nil {
			t.Fatalf("Failed to parse swagger.json: %v", err)
		}

		// Verify required OpenAPI fields
		requiredFields := []string{"openapi", "info", "paths"}
		for _, field := range requiredFields {
			if _, exists := swaggerDoc[field]; !exists {
				t.Errorf("Missing required field '%s' in swagger.json", field)
			}
		}

		// Verify OpenAPI version
		if openapi, ok := swaggerDoc["openapi"].(string); ok {
			if !strings.HasPrefix(openapi, "3.") {
				t.Errorf("Expected OpenAPI 3.x, got %s", openapi)
			}
		}

		// Verify info section
		if info, ok := swaggerDoc["info"].(map[string]interface{}); ok {
			if title, exists := info["title"]; !exists || title == "" {
				t.Error("Missing or empty title in info section")
			}
			if version, exists := info["version"]; !exists || version == "" {
				t.Error("Missing or empty version in info section")
			}
		}

		// Verify paths section has expected endpoints
		if paths, ok := swaggerDoc["paths"].(map[string]interface{}); ok {
			expectedPaths := []string{
				"/api/v1/etc-meisai/records",
				"/api/v1/etc-meisai/records/{id}",
				"/api/v1/etc-meisai/import",
				"/api/v1/etc-meisai/mappings",
			}

			for _, expectedPath := range expectedPaths {
				if _, exists := paths[expectedPath]; !exists {
					t.Errorf("Missing expected path '%s' in swagger.json", expectedPath)
				}
			}
		}

		t.Log("Successfully validated swagger.json structure")
	})

	t.Run("CORSHeaders", func(t *testing.T) {
		// Test CORS preflight request
		req, err := http.NewRequest("OPTIONS", APIBaseURL+"/records", nil)
		if err != nil {
			t.Fatalf("Failed to create OPTIONS request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send OPTIONS request: %v", err)
		}
		defer resp.Body.Close()

		// Verify CORS headers are present
		corsHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "",
			"Access-Control-Allow-Headers": "",
		}

		for header, expectedValue := range corsHeaders {
			actualValue := resp.Header.Get(header)
			if actualValue == "" {
				t.Errorf("Missing CORS header '%s'", header)
			} else if expectedValue != "" && actualValue != expectedValue {
				t.Errorf("CORS header '%s': expected '%s', got '%s'", header, expectedValue, actualValue)
			}
		}

		t.Log("Successfully verified CORS headers")
	})
}

// TestHTTPGatewayEndpoints tests HTTP gateway API endpoints
func TestHTTPGatewayEndpoints(t *testing.T) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Test data for HTTP requests
	testRecord := map[string]interface{}{
		"record": map[string]interface{}{
			"hash":            "http-test-123",
			"date":            "2025-09-21",
			"time":            "10:30:00",
			"entrance_ic":     "東京IC",
			"exit_ic":         "横浜IC",
			"toll_amount":     1200,
			"car_number":      "品川 300 あ 1234",
			"etc_card_number": "1234567890123456",
		},
	}

	var createdRecordId int64

	t.Run("CreateRecordViaHTTP", func(t *testing.T) {
		jsonData, err := json.Marshal(testRecord)
		if err != nil {
			t.Fatalf("Failed to marshal test record: %v", err)
		}

		resp, err := httpClient.Post(
			APIBaseURL+"/records",
			"application/json",
			strings.NewReader(string(jsonData)),
		)
		if err != nil {
			t.Fatalf("Failed to create record via HTTP: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200/201, got %d. Response: %s", resp.StatusCode, string(body))
		}

		// Parse response
		var createResponse map[string]interface{}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read create response: %v", err)
		}

		err = json.Unmarshal(body, &createResponse)
		if err != nil {
			t.Fatalf("Failed to parse create response: %v", err)
		}

		// Extract record ID
		if record, ok := createResponse["record"].(map[string]interface{}); ok {
			if id, ok := record["id"].(float64); ok {
				createdRecordId = int64(id)
			}
		}

		if createdRecordId <= 0 {
			t.Fatalf("Expected positive record ID, got %d", createdRecordId)
		}

		t.Logf("Successfully created record via HTTP with ID: %d", createdRecordId)
	})

	t.Run("GetRecordViaHTTP", func(t *testing.T) {
		url := fmt.Sprintf("%s/records/%d", APIBaseURL, createdRecordId)
		resp, err := httpClient.Get(url)
		if err != nil {
			t.Fatalf("Failed to get record via HTTP: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
		}

		var getResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&getResponse)
		if err != nil {
			t.Fatalf("Failed to parse get response: %v", err)
		}

		// Verify record data
		if record, ok := getResponse["record"].(map[string]interface{}); ok {
			if hash, ok := record["hash"].(string); !ok || hash != "http-test-123" {
				t.Errorf("Expected hash 'http-test-123', got %v", hash)
			}
		} else {
			t.Error("Expected record in response")
		}

		t.Log("Successfully retrieved record via HTTP")
	})

	t.Run("ListRecordsViaHTTP", func(t *testing.T) {
		url := fmt.Sprintf("%s/records?page=1&page_size=10", APIBaseURL)
		resp, err := httpClient.Get(url)
		if err != nil {
			t.Fatalf("Failed to list records via HTTP: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
		}

		var listResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&listResponse)
		if err != nil {
			t.Fatalf("Failed to parse list response: %v", err)
		}

		// Verify pagination fields
		expectedFields := []string{"records", "total_count", "page", "page_size"}
		for _, field := range expectedFields {
			if _, exists := listResponse[field]; !exists {
				t.Errorf("Missing field '%s' in list response", field)
			}
		}

		// Verify our record is in the list
		if records, ok := listResponse["records"].([]interface{}); ok {
			found := false
			for _, recordInterface := range records {
				if record, ok := recordInterface.(map[string]interface{}); ok {
					if id, ok := record["id"].(float64); ok && int64(id) == createdRecordId {
						found = true
						break
					}
				}
			}
			if !found {
				t.Error("Created record not found in list")
			}
		}

		t.Log("Successfully listed records via HTTP")
	})

	t.Run("UpdateRecordViaHTTP", func(t *testing.T) {
		updateData := map[string]interface{}{
			"record": map[string]interface{}{
				"id":              createdRecordId,
				"hash":            "http-test-123",
				"date":            "2025-09-22", // Changed date
				"time":            "15:45:00",   // Changed time
				"entrance_ic":     "横浜IC",       // Changed entrance
				"exit_ic":         "静岡IC",       // Changed exit
				"toll_amount":     2500,         // Changed amount
				"car_number":      "品川 300 あ 1234",
				"etc_card_number": "1234567890123456",
			},
		}

		jsonData, err := json.Marshal(updateData)
		if err != nil {
			t.Fatalf("Failed to marshal update data: %v", err)
		}

		url := fmt.Sprintf("%s/records/%d", APIBaseURL, createdRecordId)
		req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonData)))
		if err != nil {
			t.Fatalf("Failed to create PUT request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to update record via HTTP: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
		}

		var updateResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&updateResponse)
		if err != nil {
			t.Fatalf("Failed to parse update response: %v", err)
		}

		// Verify the update
		if record, ok := updateResponse["record"].(map[string]interface{}); ok {
			if date, ok := record["date"].(string); !ok || date != "2025-09-22" {
				t.Errorf("Expected date '2025-09-22', got %v", date)
			}
			if tollAmount, ok := record["toll_amount"].(float64); !ok || int(tollAmount) != 2500 {
				t.Errorf("Expected toll amount 2500, got %v", tollAmount)
			}
		}

		t.Log("Successfully updated record via HTTP")
	})

	t.Run("DeleteRecordViaHTTP", func(t *testing.T) {
		url := fmt.Sprintf("%s/records/%d", APIBaseURL, createdRecordId)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			t.Fatalf("Failed to create DELETE request: %v", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to delete record via HTTP: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200/204, got %d. Response: %s", resp.StatusCode, string(body))
		}

		// Verify deletion by trying to get the record
		getURL := fmt.Sprintf("%s/records/%d", APIBaseURL, createdRecordId)
		getResp, err := httpClient.Get(getURL)
		if err != nil {
			t.Fatalf("Failed to verify deletion: %v", err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for deleted record, got %d", getResp.StatusCode)
		}

		t.Log("Successfully deleted record via HTTP")
	})
}

// TestAPIDocumentationCompleteness verifies API documentation completeness
func TestAPIDocumentationCompleteness(t *testing.T) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("VerifyAPIDocumentation", func(t *testing.T) {
		resp, err := httpClient.Get(SwaggerJSONURL)
		if err != nil {
			t.Fatalf("Failed to access swagger.json: %v", err)
		}
		defer resp.Body.Close()

		var swaggerDoc map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&swaggerDoc)
		if err != nil {
			t.Fatalf("Failed to parse swagger.json: %v", err)
		}

		// Verify comprehensive API coverage
		paths, ok := swaggerDoc["paths"].(map[string]interface{})
		if !ok {
			t.Fatal("No paths found in swagger.json")
		}

		// Check for all expected API endpoints
		expectedEndpoints := map[string][]string{
			"/api/v1/etc-meisai/records": {"get", "post"},
			"/api/v1/etc-meisai/records/{id}": {"get", "put", "delete"},
			"/api/v1/etc-meisai/import": {"post"},
			"/api/v1/etc-meisai/import-sessions": {"get"},
			"/api/v1/etc-meisai/import-sessions/{session_id}": {"get"},
			"/api/v1/etc-meisai/mappings": {"get", "post"},
			"/api/v1/etc-meisai/mappings/{id}": {"get", "put", "delete"},
		}

		for endpoint, expectedMethods := range expectedEndpoints {
			pathSpec, exists := paths[endpoint]
			if !exists {
				t.Errorf("Missing endpoint '%s' in API documentation", endpoint)
				continue
			}

			pathMap, ok := pathSpec.(map[string]interface{})
			if !ok {
				t.Errorf("Invalid path specification for '%s'", endpoint)
				continue
			}

			for _, method := range expectedMethods {
				if _, methodExists := pathMap[method]; !methodExists {
					t.Errorf("Missing method '%s' for endpoint '%s'", method, endpoint)
				}
			}
		}

		// Verify components/schemas section for data models
		if components, ok := swaggerDoc["components"].(map[string]interface{}); ok {
			if schemas, ok := components["schemas"].(map[string]interface{}); ok {
				expectedSchemas := []string{
					"ETCMeisaiRecord",
					"ETCMapping",
					"ImportSession",
					"CreateRecordRequest",
					"ListRecordsResponse",
				}

				for _, schema := range expectedSchemas {
					if _, exists := schemas[schema]; !exists {
						t.Errorf("Missing schema '%s' in API documentation", schema)
					}
				}
			} else {
				t.Error("Missing schemas section in components")
			}
		} else {
			t.Error("Missing components section in swagger.json")
		}

		t.Log("Successfully verified API documentation completeness")
	})

	t.Run("VerifyErrorResponses", func(t *testing.T) {
		resp, err := httpClient.Get(SwaggerJSONURL)
		if err != nil {
			t.Fatalf("Failed to access swagger.json: %v", err)
		}
		defer resp.Body.Close()

		var swaggerDoc map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&swaggerDoc)
		if err != nil {
			t.Fatalf("Failed to parse swagger.json: %v", err)
		}

		// Check that endpoints have error response documentation
		paths, ok := swaggerDoc["paths"].(map[string]interface{})
		if !ok {
			t.Fatal("No paths found in swagger.json")
		}

		checkEndpoint := "/api/v1/etc-meisai/records/{id}"
		if pathSpec, exists := paths[checkEndpoint]; exists {
			if pathMap, ok := pathSpec.(map[string]interface{}); ok {
				if getMethod, exists := pathMap["get"]; exists {
					if getMap, ok := getMethod.(map[string]interface{}); ok {
						if responses, exists := getMap["responses"]; exists {
							if respMap, ok := responses.(map[string]interface{}); ok {
								// Check for common error status codes
								expectedErrorCodes := []string{"400", "404", "500"}
								for _, code := range expectedErrorCodes {
									if _, exists := respMap[code]; !exists {
										t.Errorf("Missing error response %s for GET %s", code, checkEndpoint)
									}
								}
							}
						}
					}
				}
			}
		}

		t.Log("Successfully verified error response documentation")
	})
}

// TestHTTPErrorHandling tests HTTP error scenarios
func TestHTTPErrorHandling(t *testing.T) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("InvalidJSONRequest", func(t *testing.T) {
		invalidJSON := `{"record": {"hash": "test", "date": "invalid-json"`

		resp, err := httpClient.Post(
			APIBaseURL+"/records",
			"application/json",
			strings.NewReader(invalidJSON),
		)
		if err != nil {
			t.Fatalf("Failed to send invalid JSON request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
		}
	})

	t.Run("MissingContentType", func(t *testing.T) {
		validJSON := `{"record": {"hash": "test"}}`

		resp, err := httpClient.Post(
			APIBaseURL+"/records",
			"text/plain", // Wrong content type
			strings.NewReader(validJSON),
		)
		if err != nil {
			t.Fatalf("Failed to send request with wrong content type: %v", err)
		}
		defer resp.Body.Close()

		// Should either reject or handle gracefully
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnsupportedMediaType {
			t.Logf("Server handled wrong content type with status %d", resp.StatusCode)
		}
	})

	t.Run("NonExistentEndpoint", func(t *testing.T) {
		resp, err := httpClient.Get(APIBaseURL + "/non-existent-endpoint")
		if err != nil {
			t.Fatalf("Failed to access non-existent endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for non-existent endpoint, got %d", resp.StatusCode)
		}
	})
}