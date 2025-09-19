package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/yhonda-ohishi/etc_meisai/models"
)

// HashService provides hash calculation and comparison services
type HashService struct {
	recordConfig  models.HashConfig
	contentConfig models.HashConfig
}

// NewHashService creates a new hash service
func NewHashService() *HashService {
	return &HashService{
		recordConfig:  models.DefaultHashConfigs["record"],
		contentConfig: models.DefaultHashConfigs["file"],
	}
}

// CalculateRecordHash calculates xxHash128 for duplicate detection
// This hash is based only on identity fields that should not change
// when record content is updated (excludes amounts and other mutable data)
func (s *HashService) CalculateRecordHash(record *models.ETCMeisai) string {
	// Build hash input from identity fields only
	var parts []string

	parts = append(parts, record.Date)
	parts = append(parts, record.Time)
	parts = append(parts, s.normalize(record.ICEntry))
	parts = append(parts, s.normalize(record.ICExit))
	parts = append(parts, s.normalize(record.VehicleNo))
	parts = append(parts, record.CardNo)
	// Note: TotalAmount excluded to allow change detection

	// Join parts and calculate hash
	input := strings.Join(parts, "|")
	// Always convert to lowercase for consistency
	input = strings.ToLower(input)

	h := xxhash.Sum64String(input)
	return fmt.Sprintf("%016x", h)
}

// CalculateContentHash calculates SHA256 for change detection
func (s *HashService) CalculateContentHash(record *models.ETCMeisai) (string, error) {
	// Create a normalized representation excluding system fields
	normalized := struct {
		Date           string  `json:"date"`
		Time           string  `json:"time"`
		ICEntry        string  `json:"ic_entry"`
		ICExit         string  `json:"ic_exit"`
		VehicleNo      string  `json:"vehicle_no"`
		CardNo         string  `json:"card_no"`
		Amount         int     `json:"amount"`
		DiscountAmount int     `json:"discount_amount"`
		TotalAmount    int     `json:"total_amount"`
		UsageType      string  `json:"usage_type"`
		PaymentMethod  string  `json:"payment_method"`
		RouteCode      string  `json:"route_code"`
		Distance       float64 `json:"distance"`
		UnkoNo         string  `json:"unko_no"`
	}{
		Date:           record.Date,
		Time:           record.Time,
		ICEntry:        record.ICEntry,
		ICExit:         record.ICExit,
		VehicleNo:      record.VehicleNo,
		CardNo:         record.CardNo,
		Amount:         record.Amount,
		DiscountAmount: record.DiscountAmount,
		TotalAmount:    record.TotalAmount,
		UsageType:      record.UsageType,
		PaymentMethod:  record.PaymentMethod,
		RouteCode:      record.RouteCode,
		Distance:       record.Distance,
		UnkoNo:         record.UnkoNo,
	}

	// Serialize to JSON for consistent representation
	data, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("failed to marshal record for hashing: %w", err)
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// CompareRecords compares two records and returns changed fields
func (s *HashService) CompareRecords(old, new *models.ETCMeisai) []string {
	var changedFields []string

	if old.Date != new.Date {
		changedFields = append(changedFields, "Date")
	}
	if old.Time != new.Time {
		changedFields = append(changedFields, "Time")
	}
	if old.ICEntry != new.ICEntry {
		changedFields = append(changedFields, "ICEntry")
	}
	if old.ICExit != new.ICExit {
		changedFields = append(changedFields, "ICExit")
	}
	if old.VehicleNo != new.VehicleNo {
		changedFields = append(changedFields, "VehicleNo")
	}
	if old.CardNo != new.CardNo {
		changedFields = append(changedFields, "CardNo")
	}
	if old.Amount != new.Amount {
		changedFields = append(changedFields, "Amount")
	}
	if old.DiscountAmount != new.DiscountAmount {
		changedFields = append(changedFields, "DiscountAmount")
	}
	if old.TotalAmount != new.TotalAmount {
		changedFields = append(changedFields, "TotalAmount")
	}
	if old.UsageType != new.UsageType {
		changedFields = append(changedFields, "UsageType")
	}
	if old.PaymentMethod != new.PaymentMethod {
		changedFields = append(changedFields, "PaymentMethod")
	}
	if old.RouteCode != new.RouteCode {
		changedFields = append(changedFields, "RouteCode")
	}
	if old.Distance != new.Distance {
		changedFields = append(changedFields, "Distance")
	}
	if old.UnkoNo != new.UnkoNo {
		changedFields = append(changedFields, "UnkoNo")
	}

	return changedFields
}

// ProcessBatch processes a batch of records and calculates hashes
func (s *HashService) ProcessBatch(records []models.ETCMeisai) ([]models.ETCMeisaiWithHash, error) {
	results := make([]models.ETCMeisaiWithHash, 0, len(records))

	for _, record := range records {
		withHash := models.ETCMeisaiWithHash{
			ETCMeisai: record,
		}

		// Calculate hash (combining record and content hash into one)
		withHash.Hash = s.CalculateRecordHash(&record)

		results = append(results, withHash)
	}

	return results, nil
}

// DetectDuplicates detects duplicates in a batch of records
func (s *HashService) DetectDuplicates(records []models.ETCMeisaiWithHash, existingHashes map[string]models.HashIndex) *models.ImportDiff {
	diff := &models.ImportDiff{
		Added:   make([]models.ETCMeisai, 0),
		Updated: make([]models.ETCMeisai, 0),
		Deleted: make([]models.ETCMeisai, 0),
	}

	for _, record := range records {
		// Check if record already exists
		_, exists := existingHashes[record.Hash]

		if !exists {
			// New record
			diff.Added = append(diff.Added, record.ETCMeisai)
		} else {
			// Duplicate (already exists)
			// Could be added to Updated if content changed
			diff.Updated = append(diff.Updated, record.ETCMeisai)
		}
	}

	return diff
}

// normalize normalizes a string for consistent hashing
func (s *HashService) normalize(str string) string {
	// Always normalize for consistency

	// Remove extra spaces and normalize to consistent format
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "　", " ") // Replace full-width space

	// Normalize multiple spaces to single space
	for strings.Contains(str, "  ") {
		str = strings.ReplaceAll(str, "  ", " ")
	}

	return str
}