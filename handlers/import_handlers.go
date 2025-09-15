package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/services"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	importService *services.ImportService
}

// NewImportHandler creates a new import handler
func NewImportHandler(importService *services.ImportService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// WebImportRequest represents a request to import from web
type WebImportRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
	CardNo   string `json:"card_no,omitempty"`
}

// ImportFromWeb handles POST /api/etc/import/web
func (h *ImportHandler) ImportFromWeb(w http.ResponseWriter, r *http.Request) {
	var req WebImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Password == "" {
		h.respondError(w, http.StatusBadRequest, "UserID and Password are required")
		return
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid from_date format")
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid to_date format")
		return
	}

	// Execute import
	result, err := h.importService.ImportFromWeb(req.UserID, req.Password, fromDate, toDate, req.CardNo)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// ImportCSVFile handles POST /api/etc/import/csv
func (h *ImportHandler) ImportCSVFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Failed to get file")
		return
	}
	defer file.Close()

	// Create temp file
	tempDir := "./temp"
	os.MkdirAll(tempDir, 0755)

	tempFile := filepath.Join(tempDir, header.Filename)
	dst, err := os.Create(tempFile)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create temp file")
		return
	}
	defer dst.Close()
	defer os.Remove(tempFile) // Clean up temp file

	// Copy uploaded file
	if _, err := io.Copy(dst, file); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Import from CSV
	result, err := h.importService.ImportFromCSV(tempFile)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// Helper methods

func (h *ImportHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ImportHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, models.ErrorResponse{
		Code:    status,
		Message: message,
	})
}