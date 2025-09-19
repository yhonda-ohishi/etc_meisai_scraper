package tests

import (
	"fmt"
	"testing"
	"time"
	"path/filepath"
	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/parser"
	"github.com/yhonda-ohishi/etc_meisai/services"
)

func BenchmarkCSVParser(b *testing.B) {
	p := parser.NewETCCSVParser()
	testFile := filepath.Join("..", "testdata", "sample_etc.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseFile(testFile)
		if err != nil {
			b.Fatalf("Failed to parse file: %v", err)
		}
	}
}

func BenchmarkHashCalculation(b *testing.B) {
	hashService := services.NewHashService()
	record := &models.ETCMeisai{
		Date:      "2025/09/01",
		Time:      "10:30",
		ICEntry:   "東京IC",
		ICExit:    "横浜IC",
		VehicleNo: "品川300あ1234",
		CardNo:    "1234567890123456",
		Amount:    1500,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hashService.CalculateRecordHash(record)
	}
}

func BenchmarkBatchProcessing(b *testing.B) {
	hashService := services.NewHashService()

	// Generate test data
	records := make([]models.ETCMeisai, 1000)
	for i := 0; i < 1000; i++ {
		records[i] = models.ETCMeisai{
			Date:      fmt.Sprintf("2025/09/%02d", (i%30)+1),
			Time:      fmt.Sprintf("%02d:30", i%24),
			ICEntry:   fmt.Sprintf("IC_%d", i%10),
			ICExit:    fmt.Sprintf("IC_%d", (i+1)%10),
			VehicleNo: fmt.Sprintf("品川300あ%04d", i),
			CardNo:    fmt.Sprintf("1234567890%06d", i),
			Amount:    1000 + (i * 100),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hashService.ProcessBatch(records)
		if err != nil {
			b.Fatalf("Failed to process batch: %v", err)
		}
	}
}

func TestLargeFileProcessing(t *testing.T) {
	// Generate large dataset
	records := make([]models.ETCMeisai, 10000)
	for i := 0; i < 10000; i++ {
		records[i] = models.ETCMeisai{
			Date:      fmt.Sprintf("2025/09/%02d", (i%30)+1),
			Time:      fmt.Sprintf("%02d:%02d", i%24, i%60),
			ICEntry:   fmt.Sprintf("Entry_IC_%d", i%100),
			ICExit:    fmt.Sprintf("Exit_IC_%d", i%100),
			VehicleNo: fmt.Sprintf("品川300あ%04d", i%1000),
			CardNo:    fmt.Sprintf("1234567890%06d", i%100),
			Amount:    1000 + (i * 10),
		}
	}

	start := time.Now()

	// Process with hash service
	hashService := services.NewHashService()
	processed, err := hashService.ProcessBatch(records)
	if err != nil {
		t.Fatalf("Failed to process batch: %v", err)
	}

	duration := time.Since(start)

	// Performance target: 10,000 records in under 5 seconds
	if duration > 5*time.Second {
		t.Errorf("Processing took too long: %v (target: < 5s)", duration)
	}

	if len(processed) != len(records) {
		t.Errorf("Expected %d processed records, got %d", len(records), len(processed))
	}

	t.Logf("Processed %d records in %v", len(records), duration)
	t.Logf("Rate: %.0f records/second", float64(len(records))/duration.Seconds())
}

func TestMemoryUsage(t *testing.T) {
	// This test checks memory efficiency
	// Generate records in batches to simulate streaming

	hashService := services.NewHashService()
	totalRecords := 0
	batchSize := 1000
	numBatches := 10

	start := time.Now()

	for batch := 0; batch < numBatches; batch++ {
		records := make([]models.ETCMeisai, batchSize)
		for i := 0; i < batchSize; i++ {
			idx := batch*batchSize + i
			records[i] = models.ETCMeisai{
				Date:      fmt.Sprintf("2025/09/%02d", (idx%30)+1),
				Time:      fmt.Sprintf("%02d:30", idx%24),
				ICEntry:   fmt.Sprintf("IC_%d", idx%10),
				ICExit:    fmt.Sprintf("IC_%d", (idx+1)%10),
				VehicleNo: fmt.Sprintf("品川300あ%04d", idx),
				CardNo:    fmt.Sprintf("1234567890%06d", idx),
				Amount:    1000 + (idx * 100),
			}
		}

		_, err := hashService.ProcessBatch(records)
		if err != nil {
			t.Fatalf("Failed to process batch %d: %v", batch, err)
		}

		totalRecords += len(records)
	}

	duration := time.Since(start)

	t.Logf("Processed %d records in %d batches", totalRecords, numBatches)
	t.Logf("Total time: %v", duration)
	t.Logf("Average batch time: %v", duration/time.Duration(numBatches))
	t.Logf("Overall rate: %.0f records/second", float64(totalRecords)/duration.Seconds())
}

func TestConcurrentProcessing(t *testing.T) {
	// Test concurrent processing capability
	sessionService := services.NewSessionService()

	// Start multiple sessions concurrently
	numSessions := 5
	done := make(chan bool, numSessions)

	start := time.Now()

	for i := 0; i < numSessions; i++ {
		go func(idx int) {
			accountType := "corporate"
			if idx%2 == 0 {
				accountType = "personal"
			}

			startDate := time.Now().AddDate(0, -1, 0)
			endDate := time.Now()

			session := sessionService.StartSession(accountType, startDate, endDate)

			// Simulate processing
			time.Sleep(100 * time.Millisecond)

			sessionService.UpdateSession(session.ID, 100+idx*10, "success", "")
			done <- true
		}(i)
	}

	// Wait for all sessions to complete
	for i := 0; i < numSessions; i++ {
		<-done
	}

	duration := time.Since(start)

	// Check all sessions were created
	stats := sessionService.GetSessionStats()
	if total, ok := stats["total"].(int); !ok || total != numSessions {
		t.Errorf("Expected %d sessions, got %v", numSessions, stats["total"])
	}

	t.Logf("Processed %d concurrent sessions in %v", numSessions, duration)
}