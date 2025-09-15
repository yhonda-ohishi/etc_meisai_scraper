package etc_meisai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yhonda-ohishi/etc_meisai/config"
)

// DownloadJob represents a download job status
type DownloadJob struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	Result    interface{} `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	downloadJobs = make(map[string]*DownloadJob)
	jobsMutex    sync.RWMutex
)

// StartDownloadJobHandler godoc
// @Summary ETC明細ダウンロード（非同期）
// @Description ダウンロードジョブを開始し、ジョブIDを返す
// @Tags Download
// @Accept json
// @Produce json
// @Param request body DownloadRequest true "ダウンロードリクエスト"
// @Success 202 {object} map[string]string
// @Router /api/etc/download-async [post]
func StartDownloadJobHandler(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid from_date format",
		})
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid to_date format",
		})
		return
	}

	// Create job
	jobID := uuid.New().String()
	job := &DownloadJob{
		ID:        jobID,
		Status:    "pending",
		Progress:  0,
		Message:   "ジョブを開始しています...",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	jobsMutex.Lock()
	downloadJobs[jobID] = job
	jobsMutex.Unlock()

	// Start async processing
	go processDownloadJob(job, req, fromDate, toDate)

	// Return job ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status_url": fmt.Sprintf("/api/etc/download-status/%s", jobID),
	})
}

// GetDownloadJobStatusHandler godoc
// @Summary ダウンロードジョブステータス取得
// @Description ジョブIDを指定してダウンロード進捗を取得
// @Tags Download
// @Produce json
// @Param job_id path string true "ジョブID"
// @Success 200 {object} DownloadJob
// @Router /api/etc/download-status/{job_id} [get]
func GetDownloadJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract job_id from URL path
	jobID := r.URL.Path[len("/api/etc/download-status/"):]

	jobsMutex.RLock()
	job, exists := downloadJobs[jobID]
	jobsMutex.RUnlock()

	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Job not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// DownloadETCDataSSEHandler godoc
// @Summary ETC明細ダウンロード（Server-Sent Events）
// @Description リアルタイムで進捗を送信しながらダウンロード
// @Tags Download
// @Accept json
// @Produce text/event-stream
// @Param request body DownloadRequest true "ダウンロードリクエスト"
// @Router /api/etc/download-sse [post]
func DownloadETCDataSSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Invalid request\"}\n\n")
		flusher.Flush()
		return
	}

	// Send progress updates
	sendProgress := func(progress int, message string) {
		data := map[string]interface{}{
			"progress": progress,
			"message":  message,
		}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		flusher.Flush()
	}

	// Start download process with progress updates
	sendProgress(10, "初期化中...")
	time.Sleep(500 * time.Millisecond)

	sendProgress(20, "アカウント情報を確認中...")
	accounts := req.Accounts
	if len(accounts) == 0 {
		accounts, _ = config.LoadCorporateAccountsFromEnv()
	}

	sendProgress(30, fmt.Sprintf("%d件のアカウントを処理します", len(accounts)))

	progressPerAccount := 60 / len(accounts)
	currentProgress := 30

	for i, account := range accounts {
		currentProgress += progressPerAccount / 3
		sendProgress(currentProgress, fmt.Sprintf("アカウント %s にログイン中...", account.UserID))
		time.Sleep(1 * time.Second)

		currentProgress += progressPerAccount / 3
		sendProgress(currentProgress, fmt.Sprintf("アカウント %s のデータをダウンロード中...", account.UserID))
		time.Sleep(2 * time.Second)

		currentProgress += progressPerAccount / 3
		sendProgress(currentProgress, fmt.Sprintf("アカウント %s の処理完了 (%d/%d)", account.UserID, i+1, len(accounts)))
	}

	sendProgress(95, "データを整理中...")
	time.Sleep(500 * time.Millisecond)

	sendProgress(100, "完了しました！")

	// Send completion event
	fmt.Fprintf(w, "event: complete\ndata: {\"message\": \"All downloads completed\"}\n\n")
	flusher.Flush()
}

func processDownloadJob(job *DownloadJob, req DownloadRequest, fromDate, toDate time.Time) {
	// Update job status
	updateJob := func(status string, progress int, message string) {
		jobsMutex.Lock()
		job.Status = status
		job.Progress = progress
		job.Message = message
		job.UpdatedAt = time.Now()
		jobsMutex.Unlock()
	}

	updateJob("processing", 10, "処理を開始しています...")
	time.Sleep(1 * time.Second)

	// Get accounts
	accounts := req.Accounts
	if len(accounts) == 0 {
		var err error
		accounts, err = config.LoadCorporateAccountsFromEnv()
		if err != nil || len(accounts) == 0 {
			updateJob("failed", 0, "アカウントが見つかりません")
			job.Error = "No accounts available"
			return
		}
	}

	updateJob("processing", 20, fmt.Sprintf("%d件のアカウントを処理中...", len(accounts)))

	// Simulate download process
	progressPerAccount := 70 / len(accounts)
	currentProgress := 20

	results := []map[string]interface{}{}

	for i, account := range accounts {
		currentProgress += progressPerAccount
		updateJob("processing", currentProgress, fmt.Sprintf("アカウント %s を処理中... (%d/%d)", account.UserID, i+1, len(accounts)))

		// Simulate download time
		time.Sleep(2 * time.Second)

		// Add mock result
		results = append(results, map[string]interface{}{
			"user_id": account.UserID,
			"success": true,
			"records": 10 + i*5,
		})
	}

	updateJob("processing", 95, "結果をまとめています...")
	time.Sleep(500 * time.Millisecond)

	// Complete
	job.Result = results
	updateJob("completed", 100, "ダウンロード完了")
}

// RegisterProgressHandlers registers progress tracking API handlers
func RegisterProgressHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/etc/download-async", StartDownloadJobHandler)
	mux.HandleFunc("/api/etc/download-status/", GetDownloadJobStatusHandler)
	mux.HandleFunc("/api/etc/download-sse", DownloadETCDataSSEHandler)
}