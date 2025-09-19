package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// BaseHandler は全ハンドラーの基底構造体
type BaseHandler struct {
	ServiceRegistry *services.ServiceRegistry
	Logger          *log.Logger
	ErrorHandler    *GRPCErrorHandler
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

// RespondGRPCError はgRPCエラーを適切なHTTPレスポンスに変換して送信
func (h *BaseHandler) RespondGRPCError(w http.ResponseWriter, err error, requestID string) {
	if h.ErrorHandler == nil {
		h.ErrorHandler = NewGRPCErrorHandler()
	}

	httpStatus, errorCode, message := h.ErrorHandler.HandleGRPCError(err)
	errorDetail := h.ErrorHandler.CreateErrorDetail(err, requestID)

	h.RespondError(w, httpStatus, errorCode, message, errorDetail)
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

// HealthCheck performs comprehensive health check including gRPC services
func (h *BaseHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if h.ServiceRegistry == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Service registry not initialized", nil)
		return
	}

	result := h.ServiceRegistry.HealthCheck(ctx)

	if result.IsHealthy() {
		h.RespondJSON(w, http.StatusOK, result)
	} else {
		h.RespondJSON(w, http.StatusServiceUnavailable, result)
	}
}

// NewBaseHandler creates a new base handler with service registry
func NewBaseHandler(serviceRegistry *services.ServiceRegistry, logger *log.Logger) *BaseHandler {
	return &BaseHandler{
		ServiceRegistry: serviceRegistry,
		Logger:          logger,
		ErrorHandler:    NewGRPCErrorHandler(),
	}
}