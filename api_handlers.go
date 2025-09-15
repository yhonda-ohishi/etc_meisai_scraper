package etc_meisai

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/config"
)

// @title ETC Meisai API
// @version 0.0.1
// @description ETC利用明細の自動取得と管理API
// @host localhost:8080
// @BasePath /api

// DownloadRequest represents a download request
type DownloadRequest struct {
	Accounts []config.SimpleAccount `json:"accounts,omitempty"`
	FromDate string                  `json:"from_date" binding:"required"`
	ToDate   string                  `json:"to_date" binding:"required"`
	Config   *ClientConfig           `json:"config,omitempty"`
}

// DownloadResponse represents a download response
type DownloadResponse struct {
	Results      []DownloadResult `json:"results"`
	TotalRecords int              `json:"total_records"`
	SuccessCount int              `json:"success_count"`
	FailedCount  int              `json:"failed_count"`
}

// SingleDownloadRequest represents a single account download request
type SingleDownloadRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Password string `json:"password" binding:"required"`
	FromDate string `json:"from_date" binding:"required"`
	ToDate   string `json:"to_date" binding:"required"`
}

// HealthCheckHandler godoc
// @Summary ヘルスチェック
// @Description APIサーバーの稼働状態を確認
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadETCDataHandler godoc
// @Summary ETC明細ダウンロード
// @Description 指定した期間のETC明細をダウンロード
// @Tags Download
// @Accept json
// @Produce json
// @Param request body DownloadRequest true "ダウンロードリクエスト"
// @Success 200 {object} DownloadResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/etc/download [post]
func DownloadETCDataHandler(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		http.Error(w, "Invalid from_date format", http.StatusBadRequest)
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		http.Error(w, "Invalid to_date format", http.StatusBadRequest)
		return
	}

	// Use environment accounts if not provided
	accounts := req.Accounts
	if len(accounts) == 0 {
		accounts, _ = config.LoadCorporateAccountsFromEnv()
	}

	// Create client
	client := NewETCClient(req.Config)

	// Download data
	results, err := client.DownloadETCData(accounts, fromDate, toDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	response := DownloadResponse{
		Results: results,
	}

	for _, result := range results {
		if result.Success {
			response.SuccessCount++
			response.TotalRecords += len(result.Records)
		} else {
			response.FailedCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadSingleAccountHandler godoc
// @Summary 単一アカウントETC明細ダウンロード
// @Description 単一アカウントの明細をダウンロード
// @Tags Download
// @Accept json
// @Produce json
// @Param request body SingleDownloadRequest true "ダウンロードリクエスト"
// @Success 200 {object} DownloadResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/etc/download-single [post]
func DownloadSingleAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req SingleDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		http.Error(w, "Invalid from_date format", http.StatusBadRequest)
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		http.Error(w, "Invalid to_date format", http.StatusBadRequest)
		return
	}

	// Create client and download
	client := NewETCClient(nil)
	result, err := client.DownloadETCDataSingle(req.UserID, req.Password, fromDate, toDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ParseCSVHandler godoc
// @Summary CSVファイルパース
// @Description アップロードされたCSVファイルをパース
// @Tags Parse
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSVファイル"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/etc/parse-csv [post]
func ParseCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save temporarily and parse
	// (Implementation would save file and call ParseETCCSV)

	response := map[string]interface{}{
		"message": "CSV parsing endpoint - implementation pending",
		"status":  "not_implemented",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterAPIHandlers registers all API handlers
func RegisterAPIHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", HealthCheckHandler)
	mux.HandleFunc("/api/etc/download", DownloadETCDataHandler)
	mux.HandleFunc("/api/etc/download-single", DownloadSingleAccountHandler)
	mux.HandleFunc("/api/etc/parse-csv", ParseCSVHandler)
}