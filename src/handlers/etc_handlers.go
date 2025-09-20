package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// ETCHandler handles HTTP requests for ETC meisai with integrated services
type ETCHandler struct {
	*BaseHandler
}

// NewETCHandler creates a new ETC handler with service registry
func NewETCHandler(serviceRegistry *services.ServiceRegistry, logger *log.Logger) *ETCHandler {
	return &ETCHandler{
		BaseHandler: NewBaseHandler(serviceRegistry, logger),
	}
}

// ImportData handles POST /api/etc/import
func (h *ETCHandler) ImportData(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req models.ETCImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request",
			"Invalid request body", err.Error())
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	result, err := etcService.ImportData(req)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondSuccess(w, result, "Import completed successfully")
}

// GetMeisai handles GET /api/etc/meisai
func (h *ETCHandler) GetMeisai(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")

	// Validate required parameters
	if fromDate == "" || toDate == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_parameters",
			"from_date and to_date are required", nil)
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	meisai, err := etcService.GetMeisaiByDateRange(fromDate, toDate)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondSuccess(w, meisai, "Meisai data retrieved successfully")
}

// GetMeisaiByID handles GET /api/etc/meisai/{id}
func (h *ETCHandler) GetMeisaiByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_id",
			"ID parameter is required", nil)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_id",
			"ID must be a valid integer", err.Error())
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	meisai, err := etcService.GetByID(ctx, id)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondSuccess(w, meisai, "Meisai record retrieved successfully")
}

// CreateMeisai handles POST /api/etc/meisai
func (h *ETCHandler) CreateMeisai(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var m models.ETCMeisai
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request",
			"Invalid request body", err.Error())
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	created, err := etcService.Create(ctx, &m)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondJSON(w, http.StatusCreated, created)
}

// ListETCMeisai handles GET /api/etc - List all ETC records with filtering
func (h *ETCHandler) ListETCMeisai(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Parse query parameters
	params := &models.ETCListParams{
		Limit:  100,
		Offset: 0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	if etcNum := r.URL.Query().Get("etc_number"); etcNum != "" {
		params.ETCNumber = etcNum
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	records, total, err := etcService.List(ctx, params)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	response := map[string]interface{}{
		"records": records,
		"total":   total,
		"limit":   params.Limit,
		"offset":  params.Offset,
	}

	h.RespondSuccess(w, response, "Records retrieved successfully")
}

// GetETCMeisai handles GET /api/etc/{id} - Get single ETC record by ID
func (h *ETCHandler) GetETCMeisai(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_id",
			"ID parameter is required", nil)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_id",
			"ID must be a valid integer", err.Error())
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	record, err := etcService.GetByID(ctx, id)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondSuccess(w, record, "Record retrieved successfully")
}

// CreateETCMeisai handles POST /api/etc - Create new ETC record
func (h *ETCHandler) CreateETCMeisai(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var record models.ETCMeisai
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request",
			"Invalid request body", err.Error())
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	created, err := etcService.Create(ctx, &record)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondJSON(w, http.StatusCreated, created)
}

// UpdateETCMeisai handles PUT /api/etc/{id} - Update ETC record
func (h *ETCHandler) UpdateETCMeisai(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_id",
			"ID parameter is required", nil)
		return
	}

	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_id",
			"ID must be a valid integer", err.Error())
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request",
			"Invalid request body", err.Error())
		return
	}

	// ETCService doesn't have Update method yet, return not implemented
	h.RespondError(w, http.StatusNotImplemented, "not_implemented",
		"Update operation not yet implemented", nil)
}

// DeleteETCMeisai handles DELETE /api/etc/{id} - Delete ETC record
func (h *ETCHandler) DeleteETCMeisai(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_id",
			"ID parameter is required", nil)
		return
	}

	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_id",
			"ID must be a valid integer", err.Error())
		return
	}

	// ETCService doesn't have Delete method yet, return not implemented
	h.RespondError(w, http.StatusNotImplemented, "not_implemented",
		"Delete operation not yet implemented", nil)
}

// BulkCreateETCMeisai handles POST /api/etc/bulk - Create multiple records
func (h *ETCHandler) BulkCreateETCMeisai(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	var records []*models.ETCMeisai
	if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request",
			"Invalid request body", err.Error())
		return
	}

	if len(records) == 0 {
		h.RespondError(w, http.StatusBadRequest, "empty_request",
			"No records provided", nil)
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	result, err := etcService.ImportCSV(ctx, records)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondSuccess(w, result, "Bulk import completed")
}

// GetETCSummary handles GET /api/etc/summary - Get summary statistics
func (h *ETCHandler) GetETCSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")

	if fromDate == "" || toDate == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_parameters",
			"from_date and to_date are required", nil)
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	summary, err := etcService.GetSummary(ctx, fromDate, toDate)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondSuccess(w, summary, "Summary retrieved successfully")
}

// GetSummary handles GET /api/etc/summary
func (h *ETCHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")

	if fromDate == "" || toDate == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_parameters",
			"from_date and to_date are required", nil)
		return
	}

	// Parse dates
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_date",
			"Invalid from_date format", err.Error())
		return
	}

	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_date",
			"Invalid to_date format", err.Error())
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	// Get records and create summary
	records, err := etcService.GetByDateRange(ctx, from, to)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	// Create summary from records
	summary := h.createSummary(records)
	h.RespondSuccess(w, summary, "Summary retrieved successfully")
}

// BulkImport handles POST /api/etc/bulk-import
func (h *ETCHandler) BulkImport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second) // Longer timeout for bulk operations
	defer cancel()

	var records []*models.ETCMeisai
	if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request",
			"Invalid request body", err.Error())
		return
	}

	etcService := h.ServiceRegistry.GetETCService()
	if etcService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"ETC service not available", nil)
		return
	}

	result, err := etcService.ImportCSV(ctx, records)
	if err != nil {
		if result != nil && result.Success {
			h.RespondJSON(w, http.StatusPartialContent, result)
		} else {
			h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		}
		return
	}

	h.RespondSuccess(w, result, "Bulk import completed successfully")
}

// Helper methods

// createSummary creates a summary from ETC records
func (h *ETCHandler) createSummary(records []*models.ETCMeisai) map[string]interface{} {
	if records == nil {
		return map[string]interface{}{
			"total_count":  0,
			"total_amount": 0,
			"date_range":   map[string]string{},
		}
	}

	totalAmount := int32(0)
	var minDate, maxDate time.Time

	for i, record := range records {
		if record != nil {
			totalAmount += record.Amount

			if i == 0 {
				minDate = record.UseDate
				maxDate = record.UseDate
			} else {
				if record.UseDate.Before(minDate) {
					minDate = record.UseDate
				}
				if record.UseDate.After(maxDate) {
					maxDate = record.UseDate
				}
			}
		}
	}

	summary := map[string]interface{}{
		"total_count":  len(records),
		"total_amount": totalAmount,
	}

	if !minDate.IsZero() && !maxDate.IsZero() {
		summary["date_range"] = map[string]string{
			"from": minDate.Format("2006-01-02"),
			"to":   maxDate.Format("2006-01-02"),
		}
	}

	return summary
}