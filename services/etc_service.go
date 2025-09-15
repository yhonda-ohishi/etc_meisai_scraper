package services

import (
	"fmt"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/repositories"
)

// ETCService handles business logic for ETC meisai
type ETCService struct {
	repo *repositories.ETCRepository
}

// NewETCService creates a new ETC service
func NewETCService(repo *repositories.ETCRepository) *ETCService {
	return &ETCService{repo: repo}
}

// ImportData imports ETC meisai data for a date range
func (s *ETCService) ImportData(req models.ETCImportRequest) (*models.ETCImportResult, error) {
	_, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid from_date format: %w", err)
	}

	_, err = time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return nil, fmt.Errorf("invalid to_date format: %w", err)
	}

	// Here you would implement the actual import logic
	// For now, we'll return a mock response
	result := &models.ETCImportResult{
		Success:      true,
		ImportedRows: 0,
		Message:      fmt.Sprintf("Import request received for %s to %s", req.FromDate, req.ToDate),
		ImportedAt:   time.Now(),
	}

	return result, nil
}

// GetMeisaiByDateRange retrieves ETC meisai for a date range
func (s *ETCService) GetMeisaiByDateRange(fromDate, toDate string) ([]models.ETCMeisai, error) {
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid from_date format: %w", err)
	}

	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return nil, fmt.Errorf("invalid to_date format: %w", err)
	}

	return s.repo.GetByDateRange(from, to)
}

// GetMeisaiByUnkoNo retrieves ETC meisai by unko_no
func (s *ETCService) GetMeisaiByUnkoNo(unkoNo string) ([]models.ETCMeisai, error) {
	if unkoNo == "" {
		return nil, fmt.Errorf("unko_no is required")
	}

	return s.repo.GetByUnkoNo(unkoNo)
}

// CreateMeisai creates a new ETC meisai record
func (s *ETCService) CreateMeisai(m *models.ETCMeisai) error {
	// Validate required fields
	if m.UnkoNo == "" {
		return fmt.Errorf("unko_no is required")
	}
	if m.VehicleNo == "" {
		return fmt.Errorf("vehicle_no is required")
	}
	if m.CardNo == "" {
		return fmt.Errorf("card_no is required")
	}

	return s.repo.Insert(m)
}

// GetSummary gets summary statistics for a date range
func (s *ETCService) GetSummary(fromDate, toDate string) ([]models.ETCSummary, error) {
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid from_date format: %w", err)
	}

	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return nil, fmt.Errorf("invalid to_date format: %w", err)
	}

	return s.repo.GetSummaryByDateRange(from, to)
}

// BulkImport imports multiple ETC meisai records
func (s *ETCService) BulkImport(records []models.ETCMeisai) (*models.ETCImportResult, error) {
	// Validate records
	for i, r := range records {
		if r.UnkoNo == "" {
			return nil, fmt.Errorf("record %d: unko_no is required", i)
		}
		if r.VehicleNo == "" {
			return nil, fmt.Errorf("record %d: vehicle_no is required", i)
		}
		if r.CardNo == "" {
			return nil, fmt.Errorf("record %d: card_no is required", i)
		}
	}

	err := s.repo.BulkInsert(records)
	if err != nil {
		return &models.ETCImportResult{
			Success:    false,
			Message:    fmt.Sprintf("Failed to import records: %v", err),
			ImportedAt: time.Now(),
			Errors:     []string{err.Error()},
		}, err
	}

	return &models.ETCImportResult{
		Success:      true,
		ImportedRows: len(records),
		Message:      fmt.Sprintf("Successfully imported %d records", len(records)),
		ImportedAt:   time.Now(),
	}, nil
}