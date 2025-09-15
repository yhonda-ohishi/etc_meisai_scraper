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
	FromDate string                  `json:"from_date,omitempty"`
	ToDate   string                  `json:"to_date,omitempty"`
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
// @Summary ETC明細ダウンロード（複数アカウント対応）
// @Description 指定した期間のETC明細を複数アカウントから一括ダウンロード
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
			"error": "Invalid from_date format. Expected: YYYY-MM-DD",
		})
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid to_date format. Expected: YYYY-MM-DD",
		})
		return
	}

	// Use environment accounts if not provided
	accounts := req.Accounts
	if len(accounts) == 0 {
		var err error
		accounts, err = config.LoadCorporateAccountsFromEnv()
		if err != nil || len(accounts) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "No accounts provided in request and no accounts found in environment variable ETC_CORP_ACCOUNTS",
				"hint":  "Either provide 'accounts' in request body or set ETC_CORP_ACCOUNTS environment variable",
			})
			return
		}
	}

	// Create client
	client := NewETCClient(req.Config)

	// Download data
	results, err := client.DownloadETCData(accounts, fromDate, toDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Download failed: " + err.Error(),
		})
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
// @Description 単一アカウントの明細をダウンロード（環境変数を使用せず、リクエストボディで直接アカウント情報を指定）
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "user_id is required",
		})
		return
	}

	if req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "password is required",
		})
		return
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid from_date format. Expected: YYYY-MM-DD",
		})
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid to_date format. Expected: YYYY-MM-DD",
		})
		return
	}

	// Create client and download
	client := NewETCClient(nil)
	result, err := client.DownloadETCDataSingle(req.UserID, req.Password, fromDate, toDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Download failed: " + err.Error(),
			"hint":  "Check your credentials and ensure the account type (ohishiexp/ohishiexp1) matches the user_id",
		})
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

// GetAvailableAccountsHandler godoc
// @Summary 利用可能なアカウント一覧取得
// @Description 環境変数に設定されているアカウント名の一覧を取得（パスワードは表示しない）
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/etc/accounts [get]
func GetAvailableAccountsHandler(w http.ResponseWriter, r *http.Request) {
	accounts, err := config.LoadCorporateAccountsFromEnv()

	response := map[string]interface{}{
		"configured": err == nil && len(accounts) > 0,
		"accounts":   []string{},
		"count":      0,
	}

	if err == nil && len(accounts) > 0 {
		accountNames := make([]string, len(accounts))
		for i, account := range accounts {
			accountNames[i] = account.UserID
		}
		response["accounts"] = accountNames
		response["count"] = len(accounts)
		response["message"] = "環境変数 ETC_CORP_ACCOUNTS から読み込み"
	} else {
		response["message"] = "環境変数 ETC_CORP_ACCOUNTS が設定されていません"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadAsyncHandler handles async download requests (wrapper for StartDownloadJobHandler)
func DownloadAsyncHandler(w http.ResponseWriter, r *http.Request) {
	StartDownloadJobHandler(w, r)
}

// GetDownloadStatusHandler handles status check requests (wrapper for GetDownloadJobStatusHandler)
func GetDownloadStatusHandler(w http.ResponseWriter, r *http.Request) {
	GetDownloadJobStatusHandler(w, r)
}

// RegisterAPIHandlers registers all API handlers
func RegisterAPIHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", HealthCheckHandler)
	mux.HandleFunc("/api/etc/accounts", GetAvailableAccountsHandler)
	mux.HandleFunc("/api/etc/download", DownloadETCDataHandler)
	mux.HandleFunc("/api/etc/download-single", DownloadSingleAccountHandler)
	mux.HandleFunc("/api/etc/parse-csv", ParseCSVHandler)
}