package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// MappingHandler handles mapping-related HTTP requests
type MappingHandler struct {
	BaseHandler
}

// ETCDtakoMapping はETC明細とDtakoのマッピング
type ETCDtakoMapping struct {
	ID          int64   `json:"id"`
	ETCMeisaiID int64   `json:"etc_meisai_id"`
	DtakoRowID  int64   `json:"dtako_row_id"`
	MatchType   string  `json:"match_type"`
	Confidence  float64 `json:"confidence"`
	IsManual    bool    `json:"is_manual"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
}

// CreateMappingRequest はマッピング作成リクエスト
type CreateMappingRequest struct {
	ETCMeisaiID int64  `json:"etc_meisai_id"`
	DtakoRowID  int64  `json:"dtako_row_id"`
	MatchType   string `json:"match_type"`
}

// UpdateMappingRequest はマッピング更新リクエスト
type UpdateMappingRequest struct {
	DtakoRowID *int64 `json:"dtako_row_id,omitempty"`
	IsActive   *bool  `json:"is_active,omitempty"`
}

// AutoMatchRequest は自動マッチングリクエスト
type AutoMatchRequest struct {
	ETCNum    string  `json:"etc_num"`
	FromDate  string  `json:"from_date,omitempty"`
	ToDate    string  `json:"to_date,omitempty"`
	Threshold float64 `json:"threshold"`
}

// NewMappingHandler creates a new mapping handler
func NewMappingHandler(serviceRegistry *services.ServiceRegistry, logger *log.Logger) *MappingHandler {
	return &MappingHandler{
		BaseHandler: BaseHandler{
			ServiceRegistry: serviceRegistry,
			Logger:         logger,
		},
	}
}

// DeleteMapping deletes a mapping by ID
func (h *MappingHandler) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get ID from URL path using chi router
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_id", "Mapping ID is required", nil)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid mapping ID", nil)
		return
	}

	// Get mapping service
	mappingService := h.ServiceRegistry.GetMappingService()
	if mappingService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Mapping service not available", nil)
		return
	}

	// Delete mapping
	if err := mappingService.DeleteMapping(ctx, id); err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Mapping deleted successfully",
		"id":      id,
	}

	h.RespondSuccess(w, response, "Mapping deleted successfully")
}

// GetMappings handles GET /api/mapping
func (h *MappingHandler) GetMappings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Parse query parameters
	params := &models.MappingListParams{
		Limit:  100,
		Offset: 0,
	}

	// Parse limit and offset
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

	// Parse filters
	if etcMeisaiIDStr := r.URL.Query().Get("etc_meisai_id"); etcMeisaiIDStr != "" {
		if id, err := strconv.ParseInt(etcMeisaiIDStr, 10, 64); err == nil {
			params.ETCMeisaiID = &id
		}
	}
	if dtakoRowID := r.URL.Query().Get("dtako_row_id"); dtakoRowID != "" {
		params.DTakoRowID = dtakoRowID
	}
	if mappingType := r.URL.Query().Get("mapping_type"); mappingType != "" {
		params.MappingType = mappingType
	}

	// Get mapping service
	mappingService := h.ServiceRegistry.GetMappingService()
	if mappingService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Mapping service not available", nil)
		return
	}

	// Get mappings
	mappings, total, err := mappingService.ListMappings(ctx, params)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	response := map[string]interface{}{
		"mappings": mappings,
		"total":    total,
		"limit":    params.Limit,
		"offset":   params.Offset,
	}

	h.RespondSuccess(w, response, "Mappings retrieved successfully")
}

// CreateMapping handles POST /api/mapping
func (h *MappingHandler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		ETCMeisaiID int64   `json:"etc_meisai_id"`
		DTakoRowID  string  `json:"dtako_row_id"`
		MappingType string  `json:"mapping_type"`
		Confidence  float32 `json:"confidence"`
		Notes       string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request",
			"Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if req.ETCMeisaiID == 0 || req.DTakoRowID == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_fields",
			"etc_meisai_id and dtako_row_id are required", nil)
		return
	}

	// Set defaults
	if req.MappingType == "" {
		req.MappingType = "manual"
	}
	if req.Confidence == 0 {
		req.Confidence = 1.0
	}

	// Get mapping service
	mappingService := h.ServiceRegistry.GetMappingService()
	if mappingService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Mapping service not available", nil)
		return
	}

	// Create mapping
	mapping := &models.ETCMeisaiMapping{
		ETCMeisaiID: req.ETCMeisaiID,
		DTakoRowID:  req.DTakoRowID,
		MappingType: req.MappingType,
		Confidence:  req.Confidence,
		Notes:       req.Notes,
		CreatedBy:   "api_user", // TODO: Get from auth context
	}

	err := mappingService.CreateMapping(ctx, mapping)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondJSON(w, http.StatusCreated, mapping)
}

// UpdateMapping は既存のマッピングを更新
func (h *MappingHandler) UpdateMapping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get ID from URL path using chi router
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.RespondError(w, http.StatusBadRequest, "missing_id", "Mapping ID is required", nil)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid mapping ID", nil)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", err.Error())
		return
	}

	// Get mapping service
	mappingService := h.ServiceRegistry.GetMappingService()
	if mappingService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Mapping service not available", nil)
		return
	}

	// Update mapping
	if err := mappingService.UpdateMapping(ctx, id, req); err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	// Get updated mapping
	mapping, err := mappingService.GetMappingByID(ctx, id)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	h.RespondSuccess(w, mapping, "Mapping updated successfully")
}

// AutoMatch は自動マッチングを実行
func (h *MappingHandler) AutoMatch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	var req AutoMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", err.Error())
		return
	}

	if req.Threshold == 0 {
		req.Threshold = 0.8 // デフォルト値
	}

	// Parse date range
	var startDate, endDate time.Time
	if req.FromDate != "" {
		parsed, err := time.Parse("2006-01-02", req.FromDate)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "invalid_date", "Invalid from_date format", err.Error())
			return
		}
		startDate = parsed
	} else {
		// Default to last 30 days
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if req.ToDate != "" {
		parsed, err := time.Parse("2006-01-02", req.ToDate)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "invalid_date", "Invalid to_date format", err.Error())
			return
		}
		endDate = parsed
	} else {
		endDate = time.Now()
	}

	// Get mapping service
	mappingService := h.ServiceRegistry.GetMappingService()
	if mappingService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Mapping service not available", nil)
		return
	}

	// Run auto-matching
	results, err := mappingService.AutoMatch(ctx, startDate, endDate, float32(req.Threshold))
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	// Count matches and errors
	var matchedCount, unmatchedCount, errorCount int
	var matches []*models.ETCMeisaiMapping

	for _, result := range results {
		if result.Error != "" {
			errorCount++
		} else if result.BestMatch != nil {
			matchedCount++
			// Create mapping for best match if confidence is high enough
			if result.BestMatch.Confidence >= float32(req.Threshold) {
				mapping := &models.ETCMeisaiMapping{
					ETCMeisaiID: result.ETCMeisaiID,
					DTakoRowID:  result.BestMatch.DTakoRowID,
					MappingType: "auto",
					Confidence:  result.BestMatch.Confidence,
					Notes:       strings.Join(result.BestMatch.MatchReasons, ", "),
					CreatedBy:   "auto_match",
				}
				matches = append(matches, mapping)
			}
		} else {
			unmatchedCount++
		}
	}

	response := map[string]interface{}{
		"matched_count":   matchedCount,
		"unmatched_count": unmatchedCount,
		"error_count":     errorCount,
		"matches":         matches,
		"threshold":       req.Threshold,
		"date_range": map[string]string{
			"from": startDate.Format("2006-01-02"),
			"to":   endDate.Format("2006-01-02"),
		},
	}

	h.RespondSuccess(w, response, "Auto-matching completed")
}