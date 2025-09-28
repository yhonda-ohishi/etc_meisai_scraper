package integration_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/grpc"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCServer_DownloadService(t *testing.T) {
	// Setup server
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	server := grpc.NewServer(nil, logger)

	// Start server in background
	port := "50052" // Different port to avoid conflicts

	go func() {
		if err := server.Start(port); err != nil {
			t.Errorf("Server failed: %v", err)
		}
	}()
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	conn, err := grpclib.Dial("localhost:"+port, grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDownloadServiceClient(conn)

	// Test GetAllAccountIDs
	t.Run("GetAllAccountIDs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.GetAllAccountIDs(ctx, &pb.GetAllAccountIDsRequest{})
		if err != nil {
			t.Errorf("GetAllAccountIDs failed: %v", err)
		}

		if resp == nil {
			t.Error("Expected non-nil response")
		}
	})

	// Test DownloadAsync
	t.Run("DownloadAsync", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := &pb.DownloadRequest{
			Accounts: []string{"test1:pass1"},
			FromDate: "2024-01-01",
			ToDate:   "2024-01-31",
		}

		resp, err := client.DownloadAsync(ctx, req)
		if err != nil {
			t.Errorf("DownloadAsync failed: %v", err)
		}

		if resp == nil || resp.JobId == "" {
			t.Error("Expected job ID in response")
		}

		// Test GetJobStatus with the returned job ID
		if resp != nil && resp.JobId != "" {
			statusResp, err := client.GetJobStatus(ctx, &pb.GetJobStatusRequest{
				JobId: resp.JobId,
			})
			if err != nil {
				t.Errorf("GetJobStatus failed: %v", err)
			}

			if statusResp == nil || statusResp.JobId != resp.JobId {
				t.Error("Expected matching job ID in status response")
			}
		}
	})
}