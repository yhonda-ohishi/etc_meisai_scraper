package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// BaseHandler は全ハンドラーの基底構造体
type BaseHandler struct {
	DB     *sql.DB
	Logger *log.Logger
}

// ErrorResponse は統一エラーレスポンス
type ErrorResponse struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

// SuccessResponse は統一成功レスポンス
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondJSON は JSONレスポンスを送信
func (h *BaseHandler) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.Logger.Printf("Failed to encode response: %v", err)
	}
}

// RespondError はエラーレスポンスを送信
func (h *BaseHandler) RespondError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	resp.Error.Details = details
	h.RespondJSON(w, status, resp)
}

// RespondSuccess は成功レスポンスを送信
func (h *BaseHandler) RespondSuccess(w http.ResponseWriter, data interface{}, message string) {
	resp := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	h.RespondJSON(w, http.StatusOK, resp)
}