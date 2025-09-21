package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/parser"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// ImportService handles CSV import operations
type ImportService struct {
	dbClient    *clients.DBServiceClient
	etcRepo     repositories.ETCRepository
	mappingRepo repositories.MappingRepository
	parser      *parser.ETCCSVParser
}

// NewImportService creates a new import service
func NewImportService(dbClient *clients.DBServiceClient, etcRepo repositories.ETCRepository, mappingRepo repositories.MappingRepository) *ImportService {
	return &ImportService{
		dbClient:    dbClient,
		etcRepo:     etcRepo,
		mappingRepo: mappingRepo,
		parser:      parser.NewETCCSVParser(),
	}
}

// ProcessCSVFile processes a CSV file and imports the data
func (s *ImportService) ProcessCSVFile(ctx context.Context, filePath string, accountID string, importType string) (*models.ETCImportBatch, error) {
	// Read the CSV file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	// Create import batch via gRPC
	batchReq := &pb.CreateImportBatchRequest{
		FileName:    filepath.Base(filePath),
		FileSize:    int64(len(data)),
		AccountId:   accountID,
		ImportType:  importType,
		Status:      "processing",
		TotalRows:   0,
		ProcessedRows: 0,
	}

	batchResp, err := s.dbClient.CreateImportBatch(ctx, batchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create import batch: %w", err)
	}

	// Process CSV data via gRPC
	processReq := &pb.ProcessCSVDataRequest{
		BatchId:    batchResp.Id,
		CsvContent: string(data),
		AccountId:  accountID,
	}

	processResp, err := s.dbClient.ProcessCSVData(ctx, processReq)
	if err != nil {
		return nil, fmt.Errorf("failed to process CSV data: %w", err)
	}

	// Convert response to model
	batch := &models.ETCImportBatch{
		ID:            batchResp.Id,
		FileName:      batchResp.FileName,
		FileSize:      batchResp.FileSize,
		AccountID:     batchResp.AccountId,
		ImportType:    batchResp.ImportType,
		Status:        processResp.Status,
		TotalRows:     processResp.TotalRows,
		ProcessedRows: processResp.ProcessedRows,
		SuccessCount:  processResp.SuccessCount,
		ErrorCount:    processResp.ErrorCount,
		CreatedAt:     batchResp.CreatedAt.AsTime(),
		UpdatedAt:     processResp.UpdatedAt.AsTime(),
	}

	if batchResp.CompletedAt != nil {
		completedAt := batchResp.CompletedAt.AsTime()
		batch.CompletedAt = &completedAt
	}

	return batch, nil
}

// ParseAndValidateCSV parses and validates CSV content without importing
func (s *ImportService) ParseAndValidateCSV(ctx context.Context, content string, accountID string) (*parser.ParseResult, error) {
	// Parse CSV content
	result, err := s.parser.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Validate each record
	for _, record := range result.Records {
		// Generate hash for duplicate detection
		hash := s.generateRecordHash(record)
		record.Hash = hash

		// Check for duplicates via repository
		existing, err := s.etcRepo.GetByHash(hash)
		if err == nil && existing != nil {
			result.DuplicateCount++
		}
	}

	return result, nil
}

// ImportParsedRecords imports pre-parsed records
func (s *ImportService) ImportParsedRecords(ctx context.Context, records []*models.ETCMeisai, batchID int64) error {
	// Bulk create via repository
	if err := s.etcRepo.BulkInsert(records); err != nil {
		return fmt.Errorf("failed to bulk create records: %w", err)
	}

	return nil
}

// GetImportProgress gets the progress of an import batch
func (s *ImportService) GetImportProgress(ctx context.Context, batchID int64) (*models.ImportProgress, error) {
	req := &pb.GetImportProgressRequest{
		BatchId: batchID,
	}

	resp, err := s.dbClient.GetImportProgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get import progress: %w", err)
	}

	return &models.ImportProgress{
		BatchID:       resp.BatchId,
		Status:        resp.Status,
		TotalRows:     resp.TotalRows,
		ProcessedRows: resp.ProcessedRows,
		SuccessCount:  resp.SuccessCount,
		ErrorCount:    resp.ErrorCount,
		Percentage:    resp.Percentage,
		Message:       resp.Message,
		UpdatedAt:     resp.UpdatedAt.AsTime(),
	}, nil
}

// generateRecordHash generates a SHA256 hash for duplicate detection
func (s *ImportService) generateRecordHash(record *models.ETCMeisai) string {
	// Create a unique string from key fields
	hashInput := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d",
		record.ETCNumber,
		record.UseDate.Format("2006-01-02"),
		record.UseTime,
		record.EntryIC,
		record.ExitIC,
		record.CarNumber,
		record.Amount,
	)

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

// ValidateImportFile validates an import file before processing
func (s *ImportService) ValidateImportFile(filePath string) error {
	// Check file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Check file size (max 100MB)
	if info.Size() > 100*1024*1024 {
		return fmt.Errorf("file too large: %d bytes (max 100MB)", info.Size())
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".csv" {
		return fmt.Errorf("invalid file type: %s (expected .csv)", ext)
	}

	return nil
}

// CancelImport cancels an ongoing import
func (s *ImportService) CancelImport(ctx context.Context, batchID int64) error {
	// Update batch status via gRPC
	// This would require adding a CancelImport RPC to the proto definition
	// For now, we'll return an error indicating it's not implemented
	return fmt.Errorf("cancel import not yet implemented in gRPC service")
}

// GetImportHistory retrieves import history for an account
func (s *ImportService) GetImportHistory(ctx context.Context, accountID string, limit int) ([]*models.ETCImportBatch, error) {
	// This would require adding a GetImportHistory RPC to the proto definition
	// For now, return empty list
	return []*models.ETCImportBatch{}, nil
}

// RetryImport retries a failed import batch
func (s *ImportService) RetryImport(ctx context.Context, batchID int64) (*models.ETCImportBatch, error) {
	// Get the original batch details
	_, err := s.GetImportProgress(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch details: %w", err)
	}

	// Create a new batch for retry
	// This would require storing the original CSV content or file path
	return nil, fmt.Errorf("retry import requires stored CSV content: not yet implemented")
}

// HealthCheck performs a health check on the import service
func (s *ImportService) HealthCheck(ctx context.Context) error {
	// Check gRPC client connectivity
	if s.dbClient == nil {
		return fmt.Errorf("db client not initialized")
	}

	// Try to get progress for a non-existent batch to test connectivity
	_, err := s.GetImportProgress(ctx, -1)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("import service health check failed: %w", err)
	}

	return nil
}