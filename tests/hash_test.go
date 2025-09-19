package tests

import (
	"testing"
	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/services"
)

func TestHashService(t *testing.T) {
	// Create hash service
	hashService := services.NewHashService()

	// Create test record
	record := &models.ETCMeisai{
		Date:      "2025/09/01",
		Time:      "10:30",
		ICEntry:   "東京IC",
		ICExit:    "横浜IC",
		VehicleNo: "品川300あ1234",
		CardNo:    "1234567890123456",
		Amount:    1500,
	}

	// Calculate record hash
	hash1 := hashService.CalculateRecordHash(record)
	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}

	// Same record should produce same hash
	hash2 := hashService.CalculateRecordHash(record)
	if hash1 != hash2 {
		t.Error("Expected same hash for same record")
	}

	// Different record should produce different hash
	record2 := &models.ETCMeisai{
		Date:      "2025/09/02",
		Time:      "11:30",
		ICEntry:   "新宿IC",
		ICExit:    "八王子IC",
		VehicleNo: "品川300い5678",
		CardNo:    "9876543210987654",
		Amount:    800,
	}
	hash3 := hashService.CalculateRecordHash(record2)
	if hash1 == hash3 {
		t.Error("Expected different hash for different record")
	}

	t.Logf("Hash 1: %s", hash1)
	t.Logf("Hash 2: %s", hash2)
	t.Logf("Hash 3: %s", hash3)
}

func TestHashServiceBatch(t *testing.T) {
	hashService := services.NewHashService()

	// Create test records
	records := []models.ETCMeisai{
		{
			Date:      "2025/09/01",
			Time:      "10:30",
			ICEntry:   "東京IC",
			ICExit:    "横浜IC",
			VehicleNo: "品川300あ1234",
			CardNo:    "1234567890123456",
		},
		{
			Date:      "2025/09/02",
			Time:      "11:30",
			ICEntry:   "新宿IC",
			ICExit:    "八王子IC",
			VehicleNo: "品川300い5678",
			CardNo:    "9876543210987654",
		},
	}

	// Process batch
	results, err := hashService.ProcessBatch(records)
	if err != nil {
		t.Fatalf("Failed to process batch: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check that hashes are different
	if results[0].Hash == results[1].Hash {
		t.Error("Expected different hashes for different records")
	}

	t.Logf("Processed %d records", len(results))
}

func TestDuplicateDetection(t *testing.T) {
	hashService := services.NewHashService()

	// Create test records
	records := []models.ETCMeisai{
		{
			Date:      "2025/09/01",
			Time:      "10:30",
			ICEntry:   "東京IC",
			ICExit:    "横浜IC",
			VehicleNo: "品川300あ1234",
			CardNo:    "1234567890123456",
		},
		{
			Date:      "2025/09/01",
			Time:      "10:30",
			ICEntry:   "東京IC",
			ICExit:    "横浜IC",
			VehicleNo: "品川300あ1234",
			CardNo:    "1234567890123456",
		}, // Duplicate
		{
			Date:      "2025/09/02",
			Time:      "11:30",
			ICEntry:   "新宿IC",
			ICExit:    "八王子IC",
			VehicleNo: "品川300い5678",
			CardNo:    "9876543210987654",
		},
	}

	// Process batch
	withHash, err := hashService.ProcessBatch(records)
	if err != nil {
		t.Fatalf("Failed to process batch: %v", err)
	}

	// Create existing hash index (simulate first record already exists)
	existingHashes := make(map[string]models.HashIndex)
	existingHashes[withHash[0].Hash] = models.HashIndex{
		Hash:     withHash[0].Hash,
		RecordID: 1,
	}

	// Detect duplicates
	diff := hashService.DetectDuplicates(withHash, existingHashes)

	if len(diff.Added) != 1 {
		t.Errorf("Expected 1 new record, got %d", len(diff.Added))
	}

	if len(diff.Updated) != 2 {
		t.Errorf("Expected 2 duplicate/updated records, got %d", len(diff.Updated))
	}

	t.Logf("Found %d new, %d duplicates", len(diff.Added), len(diff.Updated))
}