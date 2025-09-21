package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"etc_meisai/src/pb"
)

// ChiToGRPCAdapter adapts Chi HTTP handlers to gRPC service calls
type ChiToGRPCAdapter struct {
	grpcConn   *grpc.ClientConn
	etcClient  pb.EtcServiceClient
	grpcAddr   string
	timeout    time.Duration
	maxRetries int
}

// AdapterConfig holds configuration for the adapter
type AdapterConfig struct {
	GRPCAddress string
	Timeout     time.Duration
	MaxRetries  int
}

// DefaultAdapterConfig returns default configuration
func DefaultAdapterConfig() *AdapterConfig {
	return &AdapterConfig{
		GRPCAddress: "localhost:9090",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
	}
}

// NewChiToGRPCAdapter creates a new adapter instance
func NewChiToGRPCAdapter(config *AdapterConfig) (*ChiToGRPCAdapter, error) {
	if config == nil {
		config = DefaultAdapterConfig()
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(
		config.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(32*1024*1024), // 32MB
			grpc.MaxCallSendMsgSize(32*1024*1024), // 32MB
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	adapter := &ChiToGRPCAdapter{
		grpcConn:   conn,
		etcClient:  pb.NewEtcServiceClient(conn),
		grpcAddr:   config.GRPCAddress,
		timeout:    config.Timeout,
		maxRetries: config.MaxRetries,
	}

	return adapter, nil
}

// Close closes the gRPC connection
func (a *ChiToGRPCAdapter) Close() error {
	if a.grpcConn != nil {
		return a.grpcConn.Close()
	}
	return nil
}

// Health returns a health check handler
func (a *ChiToGRPCAdapter) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Call gRPC health check (if available) or just check connection
		_, err := a.etcClient.GetStats(ctx, &emptypb.Empty{})
		if err != nil {
			log.Printf("Health check failed: %v", err)
			a.writeErrorResponse(w, http.StatusServiceUnavailable, "service unavailable", err)
			return
		}

		response := map[string]interface{}{
			"status":    "healthy",
			"service":   "etc-meisai-adapter",
			"grpc_addr": a.grpcAddr,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		a.writeJSONResponse(w, http.StatusOK, response)
	}
}

// ListAccounts returns a handler for listing ETC accounts
func (a *ChiToGRPCAdapter) ListAccounts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
		defer cancel()

		resp, err := a.etcClient.ListAccounts(ctx, &emptypb.Empty{})
		if err != nil {
			a.handleGRPCError(w, err, "failed to list accounts")
			return
		}

		// Convert gRPC response to legacy format
		accounts := make([]map[string]interface{}, len(resp.Accounts))
		for i, account := range resp.Accounts {
			accounts[i] = map[string]interface{}{
				"id":           account.Id,
				"card_number":  account.CardNumber,
				"account_type": account.AccountType,
				"description":  account.Description,
				"is_active":    account.IsActive,
				"created_at":   account.CreatedAt,
				"updated_at":   account.UpdatedAt,
			}
		}

		response := map[string]interface{}{
			"accounts": accounts,
			"total":    len(accounts),
		}

		a.writeJSONResponse(w, http.StatusOK, response)
	}
}

// CreateAccount returns a handler for creating new ETC accounts
func (a *ChiToGRPCAdapter) CreateAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
		defer cancel()

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			a.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
			return
		}

		// Convert to gRPC request
		req := &pb.CreateAccountRequest{
			CardNumber:  getStringField(reqBody, "card_number"),
			AccountType: getStringField(reqBody, "account_type"),
			Description: getStringField(reqBody, "description"),
		}

		resp, err := a.etcClient.CreateAccount(ctx, req)
		if err != nil {
			a.handleGRPCError(w, err, "failed to create account")
			return
		}

		// Convert response to legacy format
		response := map[string]interface{}{
			"id":           resp.Account.Id,
			"card_number":  resp.Account.CardNumber,
			"account_type": resp.Account.AccountType,
			"description":  resp.Account.Description,
			"is_active":    resp.Account.IsActive,
			"created_at":   resp.Account.CreatedAt,
			"updated_at":   resp.Account.UpdatedAt,
		}

		a.writeJSONResponse(w, http.StatusCreated, response)
	}
}

// GetAccount returns a handler for getting account details
func (a *ChiToGRPCAdapter) GetAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract account ID from URL path
		accountID := a.extractIDFromPath(r.URL.Path, "/accounts/")
		if accountID == "" {
			a.writeErrorResponse(w, http.StatusBadRequest, "account ID required", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
		defer cancel()

		req := &pb.GetAccountRequest{Id: accountID}
		resp, err := a.etcClient.GetAccount(ctx, req)
		if err != nil {
			a.handleGRPCError(w, err, "failed to get account")
			return
		}

		// Convert response to legacy format
		response := map[string]interface{}{
			"id":           resp.Account.Id,
			"card_number":  resp.Account.CardNumber,
			"account_type": resp.Account.AccountType,
			"description":  resp.Account.Description,
			"is_active":    resp.Account.IsActive,
			"created_at":   resp.Account.CreatedAt,
			"updated_at":   resp.Account.UpdatedAt,
		}

		a.writeJSONResponse(w, http.StatusOK, response)
	}
}

// ListTransactions returns a handler for listing ETC transactions
func (a *ChiToGRPCAdapter) ListTransactions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
		defer cancel()

		// Parse query parameters
		query := r.URL.Query()
		req := &pb.ListTransactionsRequest{
			AccountId: query.Get("account_id"),
			StartDate: query.Get("start_date"),
			EndDate:   query.Get("end_date"),
			Limit:     parseIntParam(query.Get("limit"), 100),
			Offset:    parseIntParam(query.Get("offset"), 0),
		}

		resp, err := a.etcClient.ListTransactions(ctx, req)
		if err != nil {
			a.handleGRPCError(w, err, "failed to list transactions")
			return
		}

		// Convert gRPC response to legacy format
		transactions := make([]map[string]interface{}, len(resp.Transactions))
		for i, tx := range resp.Transactions {
			transactions[i] = map[string]interface{}{
				"id":           tx.Id,
				"account_id":   tx.AccountId,
				"date":         tx.Date,
				"time":         tx.Time,
				"entrance":     tx.Entrance,
				"exit":         tx.Exit,
				"highway":      tx.Highway,
				"amount":       tx.Amount,
				"etc_number":   tx.EtcNumber,
				"dtako_row_id": tx.DtakoRowId,
				"created_at":   tx.CreatedAt,
				"updated_at":   tx.UpdatedAt,
			}
		}

		response := map[string]interface{}{
			"transactions": transactions,
			"total":        resp.Total,
			"limit":        req.Limit,
			"offset":       req.Offset,
		}

		a.writeJSONResponse(w, http.StatusOK, response)
	}
}

// DownloadStatements returns a handler for downloading ETC statements
func (a *ChiToGRPCAdapter) DownloadStatements() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			a.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute) // Longer timeout for downloads
		defer cancel()

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			a.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
			return
		}

		// Convert to gRPC request
		req := &pb.DownloadStatementsRequest{
			AccountIds: getStringSliceField(reqBody, "account_ids"),
			StartDate:  getStringField(reqBody, "start_date"),
			EndDate:    getStringField(reqBody, "end_date"),
		}

		resp, err := a.etcClient.DownloadStatements(ctx, req)
		if err != nil {
			a.handleGRPCError(w, err, "failed to download statements")
			return
		}

		// Convert response to legacy format
		response := map[string]interface{}{
			"job_id":     resp.JobId,
			"status":     resp.Status,
			"message":    resp.Message,
			"started_at": resp.StartedAt,
		}

		a.writeJSONResponse(w, http.StatusAccepted, response)
	}
}

// GetStats returns a handler for getting system statistics
func (a *ChiToGRPCAdapter) GetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
		defer cancel()

		resp, err := a.etcClient.GetStats(ctx, &emptypb.Empty{})
		if err != nil {
			a.handleGRPCError(w, err, "failed to get stats")
			return
		}

		// Convert response to legacy format
		response := map[string]interface{}{
			"total_accounts":     resp.TotalAccounts,
			"total_transactions": resp.TotalTransactions,
			"last_download":      resp.LastDownload,
			"storage_used":       resp.StorageUsed,
		}

		a.writeJSONResponse(w, http.StatusOK, response)
	}
}

// Helper methods

func (a *ChiToGRPCAdapter) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func (a *ChiToGRPCAdapter) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"details": nil,
	}

	if err != nil {
		response["details"] = err.Error()
		log.Printf("API Error: %s - %v", message, err)
	}

	// Add deprecation warning for legacy endpoints
	w.Header().Set("X-API-Deprecated", "true")
	w.Header().Set("X-API-Deprecation-Info", "This endpoint is deprecated. Please use the gRPC API or grpc-gateway endpoints.")

	a.writeJSONResponse(w, statusCode, response)
}

func (a *ChiToGRPCAdapter) handleGRPCError(w http.ResponseWriter, err error, message string) {
	st, ok := status.FromError(err)
	if !ok {
		a.writeErrorResponse(w, http.StatusInternalServerError, message, err)
		return
	}

	// Map gRPC status codes to HTTP status codes
	var httpStatus int
	switch st.Code() {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.ResourceExhausted:
		httpStatus = http.StatusTooManyRequests
	case codes.FailedPrecondition:
		httpStatus = http.StatusPreconditionFailed
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		httpStatus = http.StatusRequestTimeout
	default:
		httpStatus = http.StatusInternalServerError
	}

	response := map[string]interface{}{
		"error":      message,
		"grpc_error": st.Message(),
		"grpc_code":  st.Code().String(),
		"status":     httpStatus,
	}

	a.writeJSONResponse(w, httpStatus, response)
}

func (a *ChiToGRPCAdapter) extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	id := strings.TrimPrefix(path, prefix)
	// Remove any trailing path segments
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}

	return strings.TrimSpace(id)
}

// Utility functions for request parsing

func getStringField(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSliceField(data map[string]interface{}, key string) []string {
	if val, ok := data[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return nil
}

func parseIntParam(value string, defaultValue int32) int32 {
	if value == "" {
		return defaultValue
	}

	if parsed, err := strconv.ParseInt(value, 10, 32); err == nil {
		return int32(parsed)
	}

	return defaultValue
}

// Retry wrapper for resilient gRPC calls
func (a *ChiToGRPCAdapter) withRetry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < a.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := time.Duration(attempt*attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		if err := operation(); err != nil {
			lastErr = err

			// Check if error is retryable
			if st, ok := status.FromError(err); ok {
				switch st.Code() {
				case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
					// Retryable errors
					continue
				default:
					// Non-retryable errors
					return err
				}
			}

			// Non-gRPC errors - retry
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d retries: %w", a.maxRetries, lastErr)
}