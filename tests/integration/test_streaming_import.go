//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestStreamingCSVImport tests streaming CSV import for large files
// This integration test verifies:
// 1. Stream large CSV in chunks
// 2. Monitor progress updates
// 3. Handle interruption and resume
// 4. Verify all records imported
// 5. Performance validation (10,000 rows in 5 seconds)
func TestStreamingCSVImport(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Generate large CSV content (1000 rows for testing)
	largeCSVContent := generateLargeCSVContent(1000)
	sessionId := fmt.Sprintf("stream-test-%d", time.Now().Unix())

	t.Run("StreamLargeCSVImport", func(t *testing.T) {
		// Start streaming import
		stream, err := client.ImportCSVStream(ctx)
		if err != nil {
			t.Fatalf("Failed to start streaming import: %v", err)
		}

		// Send CSV data in chunks
		chunkSize := 4096 // 4KB chunks
		csvBytes := []byte(largeCSVContent)
		totalChunks := (len(csvBytes) + chunkSize - 1) / chunkSize

		for i := 0; i < totalChunks; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(csvBytes) {
				end = len(csvBytes)
			}

			chunk := &pb.ImportCSVChunk{
				SessionId:   sessionId,
				Data:        csvBytes[start:end],
				IsLast:      i == totalChunks-1,
				ChunkNumber: int32(i + 1),
			}

			err := stream.Send(chunk)
			if err != nil {
				t.Fatalf("Failed to send chunk %d: %v", i+1, err)
			}

			t.Logf("Sent chunk %d/%d (%d bytes)", i+1, totalChunks, len(chunk.Data))
		}

		// Monitor progress updates
		progressCount := 0
		var finalProgress *pb.ImportProgressUpdate

		for {
			progress, err := stream.Recv()
			if err == io.EOF {
				t.Log("Stream completed")
				break
			}
			if err != nil {
				t.Fatalf("Failed to receive progress update: %v", err)
			}

			progressCount++
			finalProgress = progress

			t.Logf("Progress update %d: %.2f%% (Processed: %d, Success: %d, Errors: %d)",
				progressCount,
				progress.ProgressPercentage,
				progress.ProcessedRows,
				progress.SuccessRows,
				progress.ErrorRows)

			// Verify progress is reasonable
			if progress.ProgressPercentage < 0 || progress.ProgressPercentage > 100 {
				t.Errorf("Invalid progress percentage: %.2f", progress.ProgressPercentage)
			}

			if progress.ProcessedRows > progress.TotalRows {
				t.Errorf("Processed rows (%d) cannot exceed total rows (%d)",
					progress.ProcessedRows, progress.TotalRows)
			}
		}

		// Verify we received progress updates
		if progressCount == 0 {
			t.Error("Expected to receive progress updates, got none")
		}

		// Verify final progress
		if finalProgress != nil {
			if finalProgress.ProgressPercentage != 100.0 {
				t.Errorf("Expected final progress to be 100%%, got %.2f%%", finalProgress.ProgressPercentage)
			}

			if finalProgress.ProcessedRows != finalProgress.TotalRows {
				t.Errorf("Expected all rows to be processed: %d/%d",
					finalProgress.ProcessedRows, finalProgress.TotalRows)
			}

			t.Logf("Final result: %d total rows, %d success, %d errors",
				finalProgress.TotalRows, finalProgress.SuccessRows, finalProgress.ErrorRows)
		}

		// Close the send stream
		err = stream.CloseSend()
		if err != nil {
			t.Fatalf("Failed to close send stream: %v", err)
		}
	})

	// Verify imported records
	t.Run("VerifyStreamImportedRecords", func(t *testing.T) {
		// Wait a bit for final processing
		time.Sleep(2 * time.Second)

		// Count records by date range
		listReq := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 100, // Get a sample of records
			DateFrom: stringPtr("2025-09-20"),
			DateTo:   stringPtr("2025-09-25"),
		}

		listResp, err := client.ListRecords(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list imported records: %v", err)
		}

		if len(listResp.Records) == 0 {
			t.Fatal("Expected to find imported records from streaming import")
		}

		// Verify some records have the expected pattern
		foundStreamRecords := 0
		for _, record := range listResp.Records {
			if strings.Contains(record.CarNumber, "品川 300 あ") {
				foundStreamRecords++
			}
		}

		if foundStreamRecords == 0 {
			t.Error("Expected to find records from streaming import")
		}

		t.Logf("Found %d records from streaming import (sample)", foundStreamRecords)
	})
}

// TestStreamingImportPerformance tests performance requirements for streaming import
func TestStreamingImportPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Generate large dataset (10,000 rows)
	targetRows := 10000
	largeCSVContent := generateLargeCSVContent(targetRows)
	sessionId := fmt.Sprintf("perf-test-%d", time.Now().Unix())

	t.Run("Performance10kRows", func(t *testing.T) {
		startTime := time.Now()

		// Start streaming import
		stream, err := client.ImportCSVStream(ctx)
		if err != nil {
			t.Fatalf("Failed to start streaming import: %v", err)
		}

		// Send data in larger chunks for better performance
		chunkSize := 16384 // 16KB chunks
		csvBytes := []byte(largeCSVContent)
		totalChunks := (len(csvBytes) + chunkSize - 1) / chunkSize

		// Send chunks
		for i := 0; i < totalChunks; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(csvBytes) {
				end = len(csvBytes)
			}

			chunk := &pb.ImportCSVChunk{
				SessionId:   sessionId,
				Data:        csvBytes[start:end],
				IsLast:      i == totalChunks-1,
				ChunkNumber: int32(i + 1),
			}

			err := stream.Send(chunk)
			if err != nil {
				t.Fatalf("Failed to send chunk %d: %v", i+1, err)
			}
		}

		// Monitor progress and measure time
		var finalProgress *pb.ImportProgressUpdate
		for {
			progress, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Failed to receive progress: %v", err)
			}

			finalProgress = progress
		}

		err = stream.CloseSend()
		if err != nil {
			t.Fatalf("Failed to close stream: %v", err)
		}

		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// Performance validation: 10,000 rows in 5 seconds
		maxDuration := 5 * time.Second
		if duration > maxDuration {
			t.Errorf("Performance requirement not met: %v > %v for %d rows",
				duration, maxDuration, targetRows)
		}

		if finalProgress != nil {
			successRate := float64(finalProgress.SuccessRows) / float64(finalProgress.TotalRows) * 100
			t.Logf("Performance test completed in %v", duration)
			t.Logf("Processed %d rows with %.1f%% success rate",
				finalProgress.TotalRows, successRate)
			t.Logf("Throughput: %.0f rows/second",
				float64(finalProgress.TotalRows)/duration.Seconds())

			// Verify high success rate
			if successRate < 95.0 {
				t.Errorf("Expected >95%% success rate, got %.1f%%", successRate)
			}
		}
	})
}

// TestStreamingImportInterruption tests handling of interrupted streams
func TestStreamingImportInterruption(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	sessionId := fmt.Sprintf("interrupt-test-%d", time.Now().Unix())
	csvContent := generateLargeCSVContent(500)

	t.Run("InterruptedStream", func(t *testing.T) {
		// Start streaming import
		stream, err := client.ImportCSVStream(ctx)
		if err != nil {
			t.Fatalf("Failed to start streaming import: %v", err)
		}

		chunkSize := 4096
		csvBytes := []byte(csvContent)
		totalChunks := (len(csvBytes) + chunkSize - 1) / chunkSize

		// Send only half the chunks, then simulate interruption
		chunksToSend := totalChunks / 2

		for i := 0; i < chunksToSend; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(csvBytes) {
				end = len(csvBytes)
			}

			chunk := &pb.ImportCSVChunk{
				SessionId:   sessionId,
				Data:        csvBytes[start:end],
				IsLast:      false, // Don't mark as last
				ChunkNumber: int32(i + 1),
			}

			err := stream.Send(chunk)
			if err != nil {
				t.Fatalf("Failed to send chunk %d: %v", i+1, err)
			}
		}

		// Simulate interruption by closing the stream without sending IsLast=true
		err = stream.CloseSend()
		if err != nil {
			t.Fatalf("Failed to close interrupted stream: %v", err)
		}

		// Check if we get any progress updates before EOF
		progressReceived := false
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				// Expected: stream was interrupted
				t.Logf("Stream interrupted as expected: %v", err)
				break
			}
			progressReceived = true
		}

		if progressReceived {
			t.Log("Received progress updates before interruption")
		}
	})

	// Test resume capability (if supported)
	t.Run("ResumeAfterInterruption", func(t *testing.T) {
		// Try to resume or start a new session
		// This depends on whether the service supports resume functionality

		newSessionId := fmt.Sprintf("resume-test-%d", time.Now().Unix())
		stream, err := client.ImportCSVStream(ctx)
		if err != nil {
			t.Fatalf("Failed to start resume stream: %v", err)
		}

		// Send complete data in new session
		chunkSize := 4096
		csvBytes := []byte(csvContent)
		totalChunks := (len(csvBytes) + chunkSize - 1) / chunkSize

		for i := 0; i < totalChunks; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(csvBytes) {
				end = len(csvBytes)
			}

			chunk := &pb.ImportCSVChunk{
				SessionId:   newSessionId,
				Data:        csvBytes[start:end],
				IsLast:      i == totalChunks-1,
				ChunkNumber: int32(i + 1),
			}

			err := stream.Send(chunk)
			if err != nil {
				t.Fatalf("Failed to send resume chunk %d: %v", i+1, err)
			}
		}

		// Complete the resumed session
		var finalProgress *pb.ImportProgressUpdate
		for {
			progress, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Failed to receive resume progress: %v", err)
			}
			finalProgress = progress
		}

		err = stream.CloseSend()
		if err != nil {
			t.Fatalf("Failed to close resume stream: %v", err)
		}

		if finalProgress != nil && finalProgress.ProgressPercentage == 100.0 {
			t.Log("Successfully resumed and completed import")
		}
	})
}

// TestStreamingImportErrorHandling tests error scenarios in streaming import
func TestStreamingImportErrorHandling(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	t.Run("InvalidSessionId", func(t *testing.T) {
		stream, err := client.ImportCSVStream(ctx)
		if err != nil {
			t.Fatalf("Failed to start stream: %v", err)
		}

		// Send chunk with invalid session ID format
		chunk := &pb.ImportCSVChunk{
			SessionId:   "", // Empty session ID
			Data:        []byte("test data"),
			IsLast:      true,
			ChunkNumber: 1,
		}

		err = stream.Send(chunk)
		if err != nil {
			t.Logf("Expected error for empty session ID: %v", err)
		}

		// Try to receive response
		_, err = stream.Recv()
		if err == nil {
			t.Error("Expected error for invalid session ID")
		}

		stream.CloseSend()
	})

	t.Run("MixedSessionIds", func(t *testing.T) {
		stream, err := client.ImportCSVStream(ctx)
		if err != nil {
			t.Fatalf("Failed to start stream: %v", err)
		}

		sessionId1 := "session-1"
		sessionId2 := "session-2"

		// Send chunk with first session ID
		chunk1 := &pb.ImportCSVChunk{
			SessionId:   sessionId1,
			Data:        []byte("header,data\n"),
			IsLast:      false,
			ChunkNumber: 1,
		}

		err = stream.Send(chunk1)
		if err != nil {
			t.Fatalf("Failed to send first chunk: %v", err)
		}

		// Send chunk with different session ID (should cause error)
		chunk2 := &pb.ImportCSVChunk{
			SessionId:   sessionId2,
			Data:        []byte("value1,value2\n"),
			IsLast:      true,
			ChunkNumber: 2,
		}

		err = stream.Send(chunk2)
		if err != nil {
			t.Logf("Expected error for mixed session IDs: %v", err)
		}

		// Should get error when trying to receive
		_, err = stream.Recv()
		if err == nil {
			t.Error("Expected error for mixed session IDs")
		}

		stream.CloseSend()
	})

	t.Run("OutOfOrderChunks", func(t *testing.T) {
		stream, err := client.ImportCSVStream(ctx)
		if err != nil {
			t.Fatalf("Failed to start stream: %v", err)
		}

		sessionId := "out-of-order-test"

		// Send chunk 2 before chunk 1
		chunk2 := &pb.ImportCSVChunk{
			SessionId:   sessionId,
			Data:        []byte("value1,value2\n"),
			IsLast:      false,
			ChunkNumber: 2,
		}

		err = stream.Send(chunk2)
		if err != nil {
			t.Fatalf("Failed to send chunk 2: %v", err)
		}

		chunk1 := &pb.ImportCSVChunk{
			SessionId:   sessionId,
			Data:        []byte("header,data\n"),
			IsLast:      true,
			ChunkNumber: 1,
		}

		err = stream.Send(chunk1)
		if err != nil {
			t.Fatalf("Failed to send chunk 1: %v", err)
		}

		// Should handle out-of-order chunks gracefully or return error
		for {
			progress, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Logf("Received error for out-of-order chunks: %v", err)
				break
			}
			t.Logf("Progress: %.2f%%", progress.ProgressPercentage)
		}

		stream.CloseSend()
	})
}

// generateLargeCSVContent generates CSV content with specified number of rows
func generateLargeCSVContent(rows int) string {
	var builder strings.Builder

	// CSV header
	builder.WriteString("利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号\n")

	// Generate data rows
	for i := 0; i < rows; i++ {
		// Vary the data to make it realistic
		day := 20 + (i % 5)      // Days 20-24
		hour := 8 + (i % 12)     // Hours 8-19
		minute := (i * 5) % 60   // Minutes 0-55
		toll := 1000 + (i % 3000) // Toll 1000-4000
		carNum := 1000 + (i % 9000) // Car numbers 1000-9999

		entranceICs := []string{"東京IC", "横浜IC", "静岡IC", "名古屋IC", "大阪IC"}
		exitICs := []string{"横浜IC", "静岡IC", "名古屋IC", "大阪IC", "福岡IC"}

		entrance := entranceICs[i%len(entranceICs)]
		exit := exitICs[i%len(exitICs)]

		line := fmt.Sprintf("2025-09-%02d,%02d:%02d:00,%s,%s,%d,品川 300 あ %04d,%016d\n",
			day, hour, minute, entrance, exit, toll, carNum, 1234567890123456+int64(i))

		builder.WriteString(line)
	}

	return builder.String()
}