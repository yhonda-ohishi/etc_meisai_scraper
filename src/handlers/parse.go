package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/parser"
)

// ParseHandler はCSV解析関連のハンドラー
type ParseHandler struct {
	BaseHandler
	Parser *parser.ETCCSVParser
}

// ParseResponse はCSV解析レスポンス
type ParseResponse struct {
	Success     bool        `json:"success"`
	RecordCount int         `json:"record_count"`
	Records     interface{} `json:"records"`
	Errors      []string    `json:"errors,omitempty"`
}

// NewParseHandler creates a new parse handler
func NewParseHandler(base BaseHandler) *ParseHandler {
	return &ParseHandler{
		BaseHandler: base,
		Parser:      parser.NewETCCSVParser(),
	}
}

// ParseCSV はアップロードされたCSVファイルを解析
func (h *ParseHandler) ParseCSV(w http.ResponseWriter, r *http.Request) {
	// multipart/form-dataの解析
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "PARSE_ERROR", "Failed to parse multipart form", err.Error())
		return
	}

	// ファイルの取得
	file, header, err := r.FormFile("file")
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "FILE_ERROR", "Failed to get file from form", err.Error())
		return
	}
	defer file.Close()

	// アカウントタイプの取得
	accountType := r.FormValue("account_type")
	if accountType == "" {
		accountType = "corporate" // デフォルト値
	}

	// 一時ファイルに保存
	tempDir := "./downloads"
	os.MkdirAll(tempDir, 0755)
	tempFile := filepath.Join(tempDir, fmt.Sprintf("upload_%d_%s", time.Now().Unix(), header.Filename))

	dst, err := os.Create(tempFile)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "FILE_SAVE_ERROR", "Failed to save uploaded file", err.Error())
		return
	}
	defer dst.Close()
	defer os.Remove(tempFile) // 処理後に削除

	// ファイルの内容をコピー
	if _, err := io.Copy(dst, file); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "FILE_COPY_ERROR", "Failed to copy file content", err.Error())
		return
	}

	// CSVファイルの解析
	records, err := h.Parser.ParseCSVFile(tempFile, accountType == "corporate")
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "CSV_PARSE_ERROR", "Failed to parse CSV file", err.Error())
		return
	}

	response := ParseResponse{
		Success:     true,
		RecordCount: len(records),
		Records:     records,
	}

	h.RespondSuccess(w, response, fmt.Sprintf("Successfully parsed %d records", len(records)))
}