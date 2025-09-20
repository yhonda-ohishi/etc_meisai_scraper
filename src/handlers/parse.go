package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/adapters"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/parser"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// ParseHandler はCSV解析関連のハンドラー（統合サービス対応）
type ParseHandler struct {
	*BaseHandler
	Parser        *parser.ETCCSVParser
	CompatAdapter *adapters.ETCMeisaiCompatAdapter
}

// ParseResponse はCSV解析レスポンス
type ParseResponse struct {
	Success     bool                    `json:"success"`
	RecordCount int                     `json:"record_count"`
	Records     interface{}             `json:"records"`
	Errors      []string                `json:"errors,omitempty"`
	ImportResult *models.ETCImportResult `json:"import_result,omitempty"`
}

// NewParseHandler creates a new parse handler with service registry
func NewParseHandler(serviceRegistry *services.ServiceRegistry, logger *log.Logger) *ParseHandler {
	return &ParseHandler{
		BaseHandler:   NewBaseHandler(serviceRegistry, logger),
		Parser:        parser.NewETCCSVParser(),
		CompatAdapter: adapters.NewETCMeisaiCompatAdapter(),
	}
}

// ParseCSV はアップロードされたCSVファイルを解析
func (h *ParseHandler) ParseCSV(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// multipart/form-dataの解析
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "parse_error",
			"Failed to parse multipart form", err.Error())
		return
	}

	// ファイルの取得
	file, header, err := r.FormFile("file")
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "file_error",
			"Failed to get file from form", err.Error())
		return
	}
	defer file.Close()

	// アカウントタイプの取得
	accountType := r.FormValue("account_type")
	if accountType == "" {
		accountType = "corporate"
	}

	// 自動保存オプション
	autoSave := r.FormValue("auto_save") == "true"

	// 一時ファイルに保存
	tempDir := "./downloads"
	os.MkdirAll(tempDir, 0755)
	tempFile := filepath.Join(tempDir, fmt.Sprintf("upload_%d_%s", time.Now().Unix(), header.Filename))

	dst, err := os.Create(tempFile)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "file_save_error",
			"Failed to save uploaded file", err.Error())
		return
	}
	defer dst.Close()
	defer os.Remove(tempFile)

	// ファイルの内容をコピー
	if _, err := io.Copy(dst, file); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "file_copy_error",
			"Failed to copy file content", err.Error())
		return
	}

	// CSVファイルの解析
	rawRecords, err := h.Parser.ParseCSVFile(tempFile, accountType == "corporate")
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "csv_parse_error",
			"Failed to parse CSV file", err.Error())
		return
	}

	// rawRecords are already ETCMeisai models from the parser
	var etcRecords []*models.ETCMeisai

	for _, rawRecord := range rawRecords {
		etcRecords = append(etcRecords, &rawRecord)
	}

	response := ParseResponse{
		Success:     true,
		RecordCount: len(etcRecords),
		Records:     etcRecords,
		Errors:      nil,
	}

	// 自動保存が有効な場合は、データベースに保存
	if autoSave && len(etcRecords) > 0 {
		importService := h.ServiceRegistry.GetImportService()
		if importService != nil {
			// Create a batch and import records
			accountID := r.FormValue("account_id")
			if accountID == "" {
				accountID = "default"
			}
			batch, err := importService.ProcessCSVFile(ctx, tempFile, accountID, accountType)
			if err != nil {
				h.Logger.Printf("Auto-save failed: %v", err)
				// エラーがあっても解析結果は返す
			} else {
				response.ImportResult = &models.ETCImportResult{
					Success:      batch.Status == "completed",
					Message:      fmt.Sprintf("Imported %d/%d records", batch.SuccessCount, batch.TotalRows),
					RecordCount:  int(batch.TotalRows),
					ImportedRows: int(batch.SuccessCount),
					ImportedAt:   batch.CreatedAt,
				}
			}
		}
	}

	if response.Success {
		h.RespondSuccess(w, response, fmt.Sprintf("Successfully parsed %d records", len(etcRecords)))
	} else {
		h.RespondJSON(w, http.StatusPartialContent, response)
	}
}

// ParseAndImport はCSVファイルを解析して即座にインポート
func (h *ParseHandler) ParseAndImport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// multipart/form-dataの解析
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "parse_error",
			"Failed to parse multipart form", err.Error())
		return
	}

	// ファイルの取得
	file, header, err := r.FormFile("file")
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "file_error",
			"Failed to get file from form", err.Error())
		return
	}
	defer file.Close()

	// アカウントタイプの取得
	accountType := r.FormValue("account_type")
	if accountType == "" {
		accountType = "corporate"
	}
	accountID := r.FormValue("account_id")
	if accountID == "" {
		accountID = "default"
	}

	// 一時ファイルに保存
	tempDir := "./downloads"
	os.MkdirAll(tempDir, 0755)
	tempFile := filepath.Join(tempDir, fmt.Sprintf("import_%d_%s", time.Now().Unix(), header.Filename))

	dst, err := os.Create(tempFile)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "file_save_error",
			"Failed to save uploaded file", err.Error())
		return
	}
	defer dst.Close()
	defer os.Remove(tempFile)

	// ファイルの内容をコピー
	if _, err := io.Copy(dst, file); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "file_copy_error",
			"Failed to copy file content", err.Error())
		return
	}

	// Get import service
	importService := h.ServiceRegistry.GetImportService()
	if importService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Import service not available", nil)
		return
	}

	// Validate file
	if err := importService.ValidateImportFile(tempFile); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid_file",
			"Invalid import file", err.Error())
		return
	}

	// Process CSV file using import service
	batch, err := importService.ProcessCSVFile(ctx, tempFile, accountID, accountType)
	if err != nil {
		h.RespondGRPCError(w, err, r.Header.Get("X-Request-ID"))
		return
	}

	response := map[string]interface{}{
		"success":        batch.Status == "completed",
		"batch_id":       batch.ID,
		"total_rows":     batch.TotalRows,
		"processed_rows": batch.ProcessedRows,
		"success_count":  batch.SuccessCount,
		"error_count":    batch.ErrorCount,
		"status":         batch.Status,
	}

	if batch.Status == "completed" {
		h.RespondSuccess(w, response, fmt.Sprintf("Successfully imported %d records", batch.SuccessCount))
	} else {
		h.RespondJSON(w, http.StatusPartialContent, response)
	}
}