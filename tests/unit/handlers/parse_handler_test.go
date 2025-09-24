package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
)

// createTestParseHandler creates a test handler with mock service
func createTestParseHandler() *handlers.ParseHandler {
	mockRegistry := createMockServiceRegistry()
	logger := createTestLogger()

	// Create handler using BaseHandler approach, similar to other tests
	handler := &handlers.ParseHandler{
		BaseHandler:   handlers.NewBaseHandler(mockRegistry, logger),
		Parser:        nil, // Will cause parsing to fail, testing error paths
		CompatAdapter: nil, // Not used in the test paths
	}
	return handler
}

// createMultipartRequest creates a multipart form request with file
func createMultipartRequest(filename, content string, params map[string]string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	io.WriteString(part, content)

	// Add other form fields
	for key, value := range params {
		writer.WriteField(key, value)
	}

	writer.Close()

	req := httptest.NewRequest("POST", "/api/parse", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

// createInvalidMultipartRequest creates an invalid multipart request
func createInvalidMultipartRequest() *http.Request {
	req := httptest.NewRequest("POST", "/api/parse", strings.NewReader("invalid multipart data"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=invalid")
	return req
}

// TestParseHandlerCreation tests handler creation
func TestParseHandlerCreation(t *testing.T) {
	handler := createTestParseHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.BaseHandler)
}

// TestParseCSV tests the ParseCSV endpoint error paths
func TestParseCSV(t *testing.T) {
	csvContent := `利用日,利用時刻,カード番号,入口IC,出口IC,利用金額
2023-01-01,10:30,1234567890,東京,大阪,1000
2023-01-02,15:45,1234567890,名古屋,福岡,1500`

	t.Run("invalid multipart form", func(t *testing.T) {
		handler := createTestParseHandler()

		req := createInvalidMultipartRequest()
		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing file in form", func(t *testing.T) {
		handler := createTestParseHandler()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/parse", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid form with nil parser causes panic", func(t *testing.T) {
		handler := createTestParseHandler()

		req, err := createMultipartRequest("test.csv", csvContent, map[string]string{
			"account_type": "corporate",
			"auto_save":    "false",
		})
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// With nil parser, should panic when trying to parse
		assert.Panics(t, func() {
			handler.ParseCSV(w, req)
		})
	})

	t.Run("auto-save with nil parser causes panic", func(t *testing.T) {
		handler := createTestParseHandler()

		// Mock service registry to return nil import service
		mockRegistry, ok := handler.ServiceRegistry.(*MockServiceRegistry)
		assert.True(t, ok)
		mockRegistry.On("GetImportService").Return(nil)

		req, err := createMultipartRequest("test.csv", csvContent, map[string]string{
			"account_type": "personal",
			"auto_save":    "true",
			"account_id":   "test123",
		})
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Should panic when trying to use nil parser
		assert.Panics(t, func() {
			handler.ParseCSV(w, req)
		})
	})

	t.Run("default parameters with nil parser causes panic", func(t *testing.T) {
		handler := createTestParseHandler()

		req, err := createMultipartRequest("test.csv", csvContent, map[string]string{})
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Should panic when trying to use nil parser
		assert.Panics(t, func() {
			handler.ParseCSV(w, req)
		})
	})
}

// TestParseAndImport tests the ParseAndImport endpoint error paths
func TestParseAndImport(t *testing.T) {
	csvContent := `利用日,利用時刻,カード番号,入口IC,出口IC,利用金額
2023-01-01,10:30,1234567890,東京,大阪,1000`

	t.Run("invalid multipart form", func(t *testing.T) {
		handler := createTestParseHandler()

		req := createInvalidMultipartRequest()
		w := httptest.NewRecorder()

		handler.ParseAndImport(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing file in form", func(t *testing.T) {
		handler := createTestParseHandler()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/parse-import", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ParseAndImport(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestParseHandler()

		mockRegistry, ok := handler.ServiceRegistry.(*MockServiceRegistry)
		assert.True(t, ok)
		mockRegistry.On("GetImportService").Return(nil)

		req, err := createMultipartRequest("test.csv", csvContent, map[string]string{
			"account_type": "corporate",
			"account_id":   "test123",
		})
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		handler.ParseAndImport(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("default parameters", func(t *testing.T) {
		handler := createTestParseHandler()

		mockRegistry, ok := handler.ServiceRegistry.(*MockServiceRegistry)
		assert.True(t, ok)
		mockRegistry.On("GetImportService").Return(nil)

		req, err := createMultipartRequest("test.csv", csvContent, map[string]string{})
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		handler.ParseAndImport(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestParseHandlerFileHandling tests file handling scenarios
func TestParseHandlerFileHandling(t *testing.T) {
	t.Run("empty file with nil parser causes panic", func(t *testing.T) {
		handler := createTestParseHandler()

		req, err := createMultipartRequest("empty.csv", "", map[string]string{
			"account_type": "corporate",
		})
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Should panic when trying to use nil parser
		assert.Panics(t, func() {
			handler.ParseCSV(w, req)
		})
	})

	t.Run("special characters in filename with nil parser causes panic", func(t *testing.T) {
		handler := createTestParseHandler()

		csvContent := "利用日,利用時刻,カード番号,入口IC,出口IC,利用金額\n2023-01-01,10:30,1234567890,東京,大阪,1000"

		req, err := createMultipartRequest("テスト ファイル (1).csv", csvContent, map[string]string{
			"account_type": "corporate",
		})
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Should panic when trying to use nil parser
		assert.Panics(t, func() {
			handler.ParseCSV(w, req)
		})
	})
}

// TestParseHandlerErrorHandling tests various error scenarios
func TestParseHandlerErrorHandling(t *testing.T) {
	t.Run("corrupted multipart data", func(t *testing.T) {
		handler := createTestParseHandler()

		// Create corrupted multipart request
		body := bytes.NewReader([]byte("--boundary\r\nContent-Disposition: form-data;\r\n\r\ncorrupted"))
		req := httptest.NewRequest("POST", "/api/parse", body)
		req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing content type", func(t *testing.T) {
		handler := createTestParseHandler()

		req := httptest.NewRequest("POST", "/api/parse", strings.NewReader("some data"))
		w := httptest.NewRecorder()

		handler.ParseCSV(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}