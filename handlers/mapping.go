package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// MappingHandler はマッピング関連のハンドラー
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
func NewMappingHandler(base BaseHandler) *MappingHandler {
	return &MappingHandler{BaseHandler: base}
}

// GetMappings はマッピング一覧を取得
func (h *MappingHandler) GetMappings(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータの取得
	etcNum := r.URL.Query().Get("etc_num")
	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")

	// TODO: データベースから実際のマッピングを取得
	mappings := []ETCDtakoMapping{}

	// デモデータ
	if etcNum != "" {
		mappings = append(mappings, ETCDtakoMapping{
			ID:          1,
			ETCMeisaiID: 100,
			DtakoRowID:  200,
			MatchType:   "exact",
			Confidence:  1.0,
			IsManual:    false,
			CreatedBy:   "system",
			CreatedAt:   "2025-09-19T10:00:00Z",
		})
	}

	response := map[string]interface{}{
		"mappings": mappings,
		"count":    len(mappings),
		"filters": map[string]string{
			"etc_num":   etcNum,
			"from_date": fromDate,
			"to_date":   toDate,
		},
	}

	h.RespondSuccess(w, response, "Mappings retrieved successfully")
}

// CreateMapping は新しいマッピングを作成
func (h *MappingHandler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	var req CreateMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if req.ETCMeisaiID == 0 || req.DtakoRowID == 0 {
		h.RespondError(w, http.StatusBadRequest, "MISSING_FIELDS", "etc_meisai_id and dtako_row_id are required", nil)
		return
	}

	if req.MatchType == "" {
		req.MatchType = "manual"
	}

	// TODO: データベースにマッピングを保存
	mapping := ETCDtakoMapping{
		ID:          1,
		ETCMeisaiID: req.ETCMeisaiID,
		DtakoRowID:  req.DtakoRowID,
		MatchType:   req.MatchType,
		Confidence:  1.0,
		IsManual:    true,
		CreatedBy:   "user",
		CreatedAt:   "2025-09-19T10:00:00Z",
	}

	h.RespondJSON(w, http.StatusCreated, mapping)
}

// UpdateMapping は既存のマッピングを更新
func (h *MappingHandler) UpdateMapping(w http.ResponseWriter, r *http.Request) {
	// URLパスからIDを取得
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 0 {
		h.RespondError(w, http.StatusBadRequest, "MISSING_ID", "Mapping ID is required", nil)
		return
	}

	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid mapping ID", nil)
		return
	}

	var req UpdateMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// TODO: データベースでマッピングを更新
	mapping := ETCDtakoMapping{
		ID:          id,
		ETCMeisaiID: 100,
		DtakoRowID:  200,
		MatchType:   "manual",
		Confidence:  1.0,
		IsManual:    true,
		CreatedBy:   "user",
		CreatedAt:   "2025-09-19T10:00:00Z",
	}

	if req.DtakoRowID != nil {
		mapping.DtakoRowID = *req.DtakoRowID
	}

	h.RespondSuccess(w, mapping, "Mapping updated successfully")
}

// AutoMatch は自動マッチングを実行
func (h *MappingHandler) AutoMatch(w http.ResponseWriter, r *http.Request) {
	var req AutoMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if req.ETCNum == "" {
		h.RespondError(w, http.StatusBadRequest, "MISSING_ETC_NUM", "etc_num is required", nil)
		return
	}

	if req.Threshold == 0 {
		req.Threshold = 0.8 // デフォルト値
	}

	// TODO: 実際の自動マッチング処理を実装
	response := map[string]interface{}{
		"matched_count":   0,
		"unmatched_count": 0,
		"matches":         []ETCDtakoMapping{},
		"threshold":       req.Threshold,
	}

	h.RespondSuccess(w, response, "Auto-matching completed")
}