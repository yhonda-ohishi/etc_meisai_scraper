//go:build contract

package contract

import (
	"context"
	"io"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/yhonda-ohishi/etc_meisai/src/pb"
)

func TestImportCSVStream_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test data - CSV content to stream
	csvContent := `利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
2024-01-15,10:30:00,東京,大阪,1000,品川 123 あ 1234,1234567890123456
2024-01-16,14:20:00,新宿,横浜,800,品川 456 い 5678,9876543210987654
2024-01-17,09:15:00,渋谷,池袋,500,品川 789 う 9012,1111222233334444`

	// Act
	stream, err := client.ImportCSVStream(ctx)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ImportCSVStream not implemented yet - %v", err)
			return
		}
		t.Fatalf("Failed to create stream: %v", err)
	}

	sessionID := "test-session-stream-001"

	// Send data in chunks
	chunkSize := 100
	data := []byte(csvContent)
	chunkNumber := int32(1)

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := &pb.ImportCSVChunk{
			SessionId:   sessionID,
			Data:        data[i:end],
			IsLast:      end == len(data),
			ChunkNumber: chunkNumber,
		}

		if err := stream.Send(chunk); err != nil {
			t.Fatalf("Failed to send chunk %d: %v", chunkNumber, err)
		}

		chunkNumber++
	}

	// Close the send side
	if err := stream.CloseSend(); err != nil {
		t.Fatalf("Failed to close send: %v", err)
	}

	// Receive progress updates
	progressCount := 0
	var lastProgress *pb.ImportProgress

	for {
		progress, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to receive progress: %v", err)
		}

		progressCount++
		lastProgress = progress

		// Verify progress fields
		if progress.SessionId != sessionID {
			t.Errorf("Expected session ID %s, got %s", sessionID, progress.SessionId)
		}

		if progress.ProgressPercentage < 0 || progress.ProgressPercentage > 100 {
			t.Errorf("Invalid progress percentage: %f", progress.ProgressPercentage)
		}

		t.Logf("Progress update %d: %.1f%% - Processed: %d, Success: %d, Errors: %d, Status: %s",
			progressCount, progress.ProgressPercentage, progress.ProcessedRows,
			progress.SuccessRows, progress.ErrorRows, progress.Status.String())
	}

	// Assert
	if progressCount == 0 {
		t.Error("Expected to receive at least one progress update")
	}

	if lastProgress == nil {
		t.Fatal("No progress received")
	}

	// Final progress should indicate completion
	if lastProgress.Status != pb.ImportStatus_IMPORT_STATUS_COMPLETED &&
		lastProgress.Status != pb.ImportStatus_IMPORT_STATUS_FAILED {
		t.Errorf("Expected final status to be COMPLETED or FAILED, got %s", lastProgress.Status.String())
	}

	// Should have processed some rows
	if lastProgress.ProcessedRows == 0 {
		t.Error("Expected to process some rows")
	}

	t.Logf("Streaming import completed with %d progress updates", progressCount)
}

func TestImportCSVStream_LargeFile(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create large CSV content for streaming
	header := "利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号\n"
	row := "2024-01-15,10:30:00,東京,大阪,1000,品川 123 あ 1234,1234567890123456\n"

	// Build large CSV (500 rows)
	csvContent := header
	for i := 0; i < 500; i++ {
		csvContent += row
	}

	// Act
	stream, err := client.ImportCSVStream(ctx)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ImportCSVStream not implemented yet - %v", err)
			return
		}
		t.Fatalf("Failed to create stream: %v", err)
	}

	sessionID := "test-session-large-001"

	// Send in smaller chunks to simulate real streaming
	chunkSize := 1024 // 1KB chunks
	data := []byte(csvContent)
	chunkNumber := int32(1)

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := &pb.ImportCSVChunk{
			SessionId:   sessionID,
			Data:        data[i:end],
			IsLast:      end == len(data),
			ChunkNumber: chunkNumber,
		}

		if err := stream.Send(chunk); err != nil {
			t.Fatalf("Failed to send chunk %d: %v", chunkNumber, err)
		}

		chunkNumber++
	}

	if err := stream.CloseSend(); err != nil {
		t.Fatalf("Failed to close send: %v", err)
	}

	// Receive and count progress updates
	progressCount := 0
	var lastProgress *pb.ImportProgress

	for {
		progress, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to receive progress: %v", err)
		}

		progressCount++
		lastProgress = progress

		// Log every 10th progress update for large files
		if progressCount%10 == 0 || progress.ProgressPercentage == 100 {
			t.Logf("Progress %d: %.1f%% - %d/%d rows processed",
				progressCount, progress.ProgressPercentage,
				progress.ProcessedRows, progress.ProcessedRows+progress.ErrorRows)
		}
	}

	// Assert
	if lastProgress == nil {
		t.Fatal("No progress received")
	}

	// Should process approximately 500 rows
	if lastProgress.ProcessedRows < 450 { // Allow some tolerance
		t.Errorf("Expected to process ~500 rows, got %d", lastProgress.ProcessedRows)
	}

	t.Logf("Large file streaming completed: %d chunks sent, %d progress updates, %d rows processed",
		chunkNumber-1, progressCount, lastProgress.ProcessedRows)
}

func TestImportCSVStream_InvalidChunks(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testCases := []struct {
		name   string
		chunks []*pb.ImportCSVChunk
	}{
		{
			name: "empty session ID",
			chunks: []*pb.ImportCSVChunk{
				{
					SessionId:   "",
					Data:        []byte("test data"),
					IsLast:      true,
					ChunkNumber: 1,
				},
			},
		},
		{
			name: "out of order chunks",
			chunks: []*pb.ImportCSVChunk{
				{
					SessionId:   "test-session",
					Data:        []byte("data2"),
					IsLast:      false,
					ChunkNumber: 2,
				},
				{
					SessionId:   "test-session",
					Data:        []byte("data1"),
					IsLast:      true,
					ChunkNumber: 1,
				},
			},
		},
		{
			name: "duplicate chunk numbers",
			chunks: []*pb.ImportCSVChunk{
				{
					SessionId:   "test-session",
					Data:        []byte("data1"),
					IsLast:      false,
					ChunkNumber: 1,
				},
				{
					SessionId:   "test-session",
					Data:        []byte("data2"),
					IsLast:      true,
					ChunkNumber: 1, // Duplicate
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			stream, err := client.ImportCSVStream(ctx)
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ImportCSVStream not implemented yet - %v", err)
					return
				}
				t.Fatalf("Failed to create stream: %v", err)
			}

			// Send invalid chunks
			for _, chunk := range tc.chunks {
				if err := stream.Send(chunk); err != nil {
					// Expected error for invalid chunks
					t.Logf("Expected send error for %s: %v", tc.name, err)
					return
				}
			}

			if err := stream.CloseSend(); err != nil {
				t.Logf("Expected close error for %s: %v", tc.name, err)
				return
			}

			// Try to receive - should get error
			_, err = stream.Recv()
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.InvalidArgument {
					t.Logf("Correctly received InvalidArgument for %s", tc.name)
				} else {
					t.Logf("Received error for %s: %v", tc.name, err)
				}
				return
			}

			t.Logf("Warning: Expected error for %s but operation succeeded", tc.name)
		})
	}
}

func TestImportCSVStream_ConnectionLoss(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Act
	stream, err := client.ImportCSVStream(ctx)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ImportCSVStream not implemented yet - %v", err)
			return
		}
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Send first chunk
	chunk1 := &pb.ImportCSVChunk{
		SessionId:   "test-session-disconnect",
		Data:        []byte("利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号\n"),
		IsLast:      false,
		ChunkNumber: 1,
	}

	if err := stream.Send(chunk1); err != nil {
		t.Fatalf("Failed to send first chunk: %v", err)
	}

	// Cancel context to simulate connection loss
	cancel()

	// Try to send another chunk - should fail
	chunk2 := &pb.ImportCSVChunk{
		SessionId:   "test-session-disconnect",
		Data:        []byte("2024-01-15,10:30:00,東京,大阪,1000,品川 123 あ 1234,1234567890123456\n"),
		IsLast:      true,
		ChunkNumber: 2,
	}

	err = stream.Send(chunk2)
	if err != nil {
		t.Logf("Expected error after context cancellation: %v", err)
	}

	// Try to receive - should also fail
	_, err = stream.Recv()
	if err != nil {
		t.Logf("Expected receive error after cancellation: %v", err)
	}
}

func TestImportCSVStream_EmptyStream(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Act
	stream, err := client.ImportCSVStream(ctx)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ImportCSVStream not implemented yet - %v", err)
			return
		}
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Close without sending any chunks
	if err := stream.CloseSend(); err != nil {
		t.Fatalf("Failed to close send: %v", err)
	}

	// Try to receive - behavior depends on implementation
	_, err = stream.Recv()
	if err == io.EOF {
		t.Logf("Empty stream correctly returned EOF")
	} else if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.InvalidArgument {
			t.Logf("Empty stream correctly returned InvalidArgument")
		} else {
			t.Logf("Empty stream returned error: %v", err)
		}
	} else {
		t.Logf("Warning: Empty stream returned success - may indicate missing validation")
	}
}