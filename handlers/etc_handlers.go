package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/services"
)

// ETCHandler handles HTTP requests for ETC meisai
type ETCHandler struct {
	service *services.ETCService
}

// NewETCHandler creates a new ETC handler
func NewETCHandler(service *services.ETCService) *ETCHandler {
	return &ETCHandler{service: service}
}

// ImportData handles POST /api/etc/import
func (h *ETCHandler) ImportData(w http.ResponseWriter, r *http.Request) {
	var req models.ETCImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.ImportData(req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// GetMeisai handles GET /api/etc/meisai
func (h *ETCHandler) GetMeisai(w http.ResponseWriter, r *http.Request) {
	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")
	unkoNo := r.URL.Query().Get("unko_no")

	// If unko_no is provided, use it
	if unkoNo != "" {
		meisai, err := h.service.GetMeisaiByUnkoNo(unkoNo)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.respondJSON(w, http.StatusOK, meisai)
		return
	}

	// Otherwise, use date range
	if fromDate == "" || toDate == "" {
		h.respondError(w, http.StatusBadRequest, "from_date and to_date are required")
		return
	}

	meisai, err := h.service.GetMeisaiByDateRange(fromDate, toDate)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, meisai)
}

// GetMeisaiByID handles GET /api/etc/meisai/{id}
func (h *ETCHandler) GetMeisaiByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "ID is required")
		return
	}

	// Implement get by ID logic here
	h.respondError(w, http.StatusNotImplemented, "Get by ID not implemented yet")
}

// CreateMeisai handles POST /api/etc/meisai
func (h *ETCHandler) CreateMeisai(w http.ResponseWriter, r *http.Request) {
	var m models.ETCMeisai
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.CreateMeisai(&m); err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, m)
}

// GetSummary handles GET /api/etc/summary
func (h *ETCHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")

	if fromDate == "" || toDate == "" {
		h.respondError(w, http.StatusBadRequest, "from_date and to_date are required")
		return
	}

	summary, err := h.service.GetSummary(fromDate, toDate)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, summary)
}

// BulkImport handles POST /api/etc/bulk-import
func (h *ETCHandler) BulkImport(w http.ResponseWriter, r *http.Request) {
	var records []models.ETCMeisai
	if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.BulkImport(records)
	if err != nil {
		if result != nil {
			h.respondJSON(w, http.StatusPartialContent, result)
		} else {
			h.respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// Helper methods

func (h *ETCHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ETCHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, models.ErrorResponse{
		Code:    status,
		Message: message,
	})
}

// HealthCheck handles GET /health
func (h *ETCHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "etc_meisai",
		"time":    time.Now().Format(time.RFC3339),
	})
}