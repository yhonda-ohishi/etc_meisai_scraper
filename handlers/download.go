package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yhonda-ohishi/etc_meisai/services"
)

// DownloadHandler はダウンロード関連のハンドラー
type DownloadHandler struct {
	BaseHandler
	DownloadService *services.DownloadService
}

// DownloadRequest はダウンロードリクエスト
type DownloadRequest struct {
	Accounts []string `json:"accounts"`
	FromDate string   `json:"from_date"`
	ToDate   string   `json:"to_date"`
	Mode     string   `json:"mode"`
}

// JobStatus はジョブステータス
type JobStatus struct {
	JobID        string     `json:"job_id"`
	Status       string     `json:"status"`
	Progress     int        `json:"progress"`
	TotalRecords int        `json:"total_records"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// NewDownloadHandler creates a new download handler
func NewDownloadHandler(base BaseHandler, downloadService *services.DownloadService) *DownloadHandler {
	return &DownloadHandler{
		BaseHandler:     base,
		DownloadService: downloadService,
	}
}

// DownloadSync は同期ダウンロードを実行
func (h *DownloadHandler) DownloadSync(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// パラメータのデフォルト値設定
	if req.FromDate == "" || req.ToDate == "" {
		now := time.Now()
		if req.ToDate == "" {
			req.ToDate = now.Format("2006-01-02")
		}
		if req.FromDate == "" {
			lastMonth := now.AddDate(0, -1, 0)
			req.FromDate = lastMonth.Format("2006-01-02")
		}
	}

	if len(req.Accounts) == 0 {
		h.RespondError(w, http.StatusBadRequest, "MISSING_ACCOUNTS", "At least one account is required", nil)
		return
	}

	// TODO: 実際のダウンロード処理を実装
	response := map[string]interface{}{
		"success":      true,
		"record_count": 0,
		"csv_path":     "",
		"records":      []interface{}{},
	}

	h.RespondSuccess(w, response, "Download completed successfully")
}

// DownloadAsync は非同期ダウンロードを開始
func (h *DownloadHandler) DownloadAsync(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// パラメータのデフォルト値設定
	if req.FromDate == "" || req.ToDate == "" {
		now := time.Now()
		if req.ToDate == "" {
			req.ToDate = now.Format("2006-01-02")
		}
		if req.FromDate == "" {
			lastMonth := now.AddDate(0, -1, 0)
			req.FromDate = lastMonth.Format("2006-01-02")
		}
	}

	if len(req.Accounts) == 0 {
		// デフォルトで全アカウントを使用
		req.Accounts = h.DownloadService.GetAllAccountIDs()
		if len(req.Accounts) == 0 {
			h.RespondError(w, http.StatusBadRequest, "NO_ACCOUNTS", "No accounts configured", nil)
			return
		}
	}

	// ジョブIDを生成
	jobID := uuid.New().String()

	// 非同期でダウンロード開始
	go h.DownloadService.ProcessAsync(jobID, req.Accounts, req.FromDate, req.ToDate)

	response := map[string]interface{}{
		"job_id":  jobID,
		"status":  "pending",
		"message": "Download job started",
	}

	h.RespondJSON(w, http.StatusAccepted, response)
}

// GetDownloadStatus はダウンロードステータスを取得
func (h *DownloadHandler) GetDownloadStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		// URLパスから取得を試みる
		// 例: /api/download/status/{jobId}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			jobID = parts[len(parts)-1]
		}
	}

	if jobID == "" {
		h.RespondError(w, http.StatusBadRequest, "MISSING_JOB_ID", "Job ID is required", nil)
		return
	}

	// TODO: 実際のステータス取得処理を実装
	status := JobStatus{
		JobID:        jobID,
		Status:       "processing",
		Progress:     50,
		TotalRecords: 100,
	}

	h.RespondSuccess(w, status, fmt.Sprintf("Status for job %s", jobID))
}