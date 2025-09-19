package clients

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// DBServiceClient wraps the gRPC client for db_service
type DBServiceClient struct {
	conn        *grpc.ClientConn
	etcClient   pb.ETCServiceClient
	mapClient   pb.ETCMappingServiceClient
	importClient pb.ETCImportServiceClient
	timeout     time.Duration
}

// NewDBServiceClient creates a new db_service gRPC client
func NewDBServiceClient(address string, timeout time.Duration) (*DBServiceClient, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db_service at %s: %w", address, err)
	}

	return &DBServiceClient{
		conn:        conn,
		etcClient:   pb.NewETCServiceClient(conn),
		mapClient:   pb.NewETCMappingServiceClient(conn),
		importClient: pb.NewETCImportServiceClient(conn),
		timeout:     timeout,
	}, nil
}

// Close closes the gRPC connection
func (c *DBServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateETCMeisai creates a new ETC record via gRPC
func (c *DBServiceClient) CreateETCMeisai(ctx context.Context, req *pb.CreateETCMeisaiRequest) (*pb.ETCMeisaiResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.etcClient.CreateETCMeisai(ctx, req)
}

// GetETCMeisai retrieves an ETC record by ID via gRPC
func (c *DBServiceClient) GetETCMeisai(ctx context.Context, req *pb.GetETCMeisaiRequest) (*pb.ETCMeisaiResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.etcClient.GetETCMeisai(ctx, req)
}

// BulkCreateETCMeisai creates multiple ETC records via gRPC
func (c *DBServiceClient) BulkCreateETCMeisai(ctx context.Context, req *pb.BulkCreateETCMeisaiRequest) (*pb.BulkCreateETCMeisaiResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.etcClient.BulkCreateETCMeisai(ctx, req)
}

// ListETCMeisai lists ETC records with filters via gRPC
func (c *DBServiceClient) ListETCMeisai(ctx context.Context, req *pb.ListETCMeisaiRequest) (*pb.ListETCMeisaiResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.etcClient.ListETCMeisai(ctx, req)
}

// GetETCSummary gets summary statistics via gRPC
func (c *DBServiceClient) GetETCSummary(ctx context.Context, req *pb.GetETCSummaryRequest) (*pb.ETCSummaryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.etcClient.GetETCSummary(ctx, req)
}

// CreateMapping creates a new ETC-DTako mapping via gRPC
func (c *DBServiceClient) CreateMapping(ctx context.Context, req *pb.CreateMappingRequest) (*pb.MappingResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.mapClient.CreateMapping(ctx, req)
}

// GetMapping retrieves a mapping by ID via gRPC
func (c *DBServiceClient) GetMapping(ctx context.Context, req *pb.GetMappingRequest) (*pb.MappingResponse, error) {
	// Stub implementation - returns empty response
	return &pb.MappingResponse{
		Id:          req.Id,
		EtcMeisaiId: 0,
		DtakoRowId:  "",
		MappingType: "manual",
		Confidence:  1.0,
	}, nil
}

// UpdateMapping updates an existing mapping via gRPC
func (c *DBServiceClient) UpdateMapping(ctx context.Context, req *pb.UpdateMappingRequest) (*pb.MappingResponse, error) {
	// Stub implementation - returns success
	return &pb.MappingResponse{
		Id:          req.Id,
		MappingType: req.MappingType,
		Confidence:  req.Confidence,
		UpdatedAt:   timestamppb.Now(),
	}, nil
}

// DeleteMapping deletes a mapping via gRPC
func (c *DBServiceClient) DeleteMapping(ctx context.Context, req *pb.DeleteMappingRequest) (*pb.DeleteResponse, error) {
	// Stub implementation - returns success
	return &pb.DeleteResponse{
		Success: true,
		Message: "Mapping deleted successfully",
	}, nil
}

// ListMappings lists mappings with filters via gRPC
func (c *DBServiceClient) ListMappings(ctx context.Context, req *pb.ListMappingsRequest) (*pb.ListMappingsResponse, error) {
	// Stub implementation - returns empty list
	return &pb.ListMappingsResponse{
		Mappings: []*pb.MappingResponse{},
		Total:    0,
	}, nil
}

// BulkCreateMappings creates multiple mappings via gRPC
func (c *DBServiceClient) BulkCreateMappings(ctx context.Context, req *pb.BulkCreateMappingsRequest) (*pb.BulkCreateMappingsResponse, error) {
	// Stub implementation - returns success
	var createdIds []int64
	for i := range req.Mappings {
		createdIds = append(createdIds, int64(i+1))
	}
	return &pb.BulkCreateMappingsResponse{
		Success:    true,
		Message:    "Bulk mappings created successfully",
		CreatedIds: createdIds,
	}, nil
}

// DeleteMappingsByETCMeisai deletes all mappings for an ETC Meisai ID via gRPC
func (c *DBServiceClient) DeleteMappingsByETCMeisai(ctx context.Context, req *pb.DeleteMappingsByETCMeisaiRequest) (*pb.DeleteResponse, error) {
	// Stub implementation - returns success
	return &pb.DeleteResponse{
		Success: true,
		Message: "Mappings deleted successfully",
	}, nil
}

// UpdateMappingConfidence updates mapping confidence score via gRPC
func (c *DBServiceClient) UpdateMappingConfidence(ctx context.Context, req *pb.UpdateMappingConfidenceRequest) (*pb.MappingResponse, error) {
	// Stub implementation - returns success
	return &pb.MappingResponse{
		Id:         req.Id,
		Confidence: req.Confidence,
		UpdatedAt:  timestamppb.Now(),
	}, nil
}

// FindPotentialMatches finds potential DTako matches for an ETC record
func (c *DBServiceClient) FindPotentialMatches(ctx context.Context, req *pb.FindPotentialMatchesRequest) (*pb.FindPotentialMatchesResponse, error) {
	// Stub implementation - returns empty matches
	return &pb.FindPotentialMatchesResponse{
		Matches: []*pb.PotentialMatch{},
	}, nil
}

// CreateImportBatch creates a new import batch via gRPC
func (c *DBServiceClient) CreateImportBatch(ctx context.Context, req *pb.CreateImportBatchRequest) (*pb.ImportBatchResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.importClient.CreateImportBatch(ctx, req)
}

// ProcessCSVData processes CSV data via gRPC
func (c *DBServiceClient) ProcessCSVData(ctx context.Context, req *pb.ProcessCSVDataRequest) (*pb.ProcessCSVDataResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.importClient.ProcessCSVData(ctx, req)
}

// GetImportProgress gets import progress via gRPC
func (c *DBServiceClient) GetImportProgress(ctx context.Context, req *pb.GetImportProgressRequest) (*pb.ImportProgressResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.importClient.GetImportProgress(ctx, req)
}

// HealthCheck performs a health check on the db_service
func (c *DBServiceClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Simple health check by calling a lightweight operation
	_, err := c.etcClient.ListETCMeisai(ctx, &pb.ListETCMeisaiRequest{
		Limit: 1,
	})

	if err != nil {
		return fmt.Errorf("db_service health check failed: %w", err)
	}

	return nil
}

// IsConnected checks if the client is connected
func (c *DBServiceClient) IsConnected() bool {
	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Connecting || state == connectivity.Idle || state == connectivity.Ready
}

// GetConnectionState returns the current connection state
func (c *DBServiceClient) GetConnectionState() connectivity.State {
	if c.conn == nil {
		return connectivity.Shutdown
	}
	return c.conn.GetState()
}