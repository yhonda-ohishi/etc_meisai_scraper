package repositories

import (
	"context"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ImportRepositoryClient wraps the gRPC client for ImportRepository
type ImportRepositoryClient struct {
	client pb.ImportRepositoryClient
	conn   *grpc.ClientConn
}

// NewImportRepositoryClient creates a new repository client
func NewImportRepositoryClient(address string) (*ImportRepositoryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &ImportRepositoryClient{
		client: pb.NewImportRepositoryClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying connection
func (c *ImportRepositoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateSession creates a new import session
func (c *ImportRepositoryClient) CreateSession(ctx context.Context, session *pb.ImportSession) (*pb.ImportSession, error) {
	return c.client.CreateSession(ctx, session)
}

// GetSession retrieves a session by ID
func (c *ImportRepositoryClient) GetSession(ctx context.Context, sessionID string) (*pb.ImportSession, error) {
	req := &pb.GetSessionRequest{SessionId: sessionID}
	return c.client.GetSession(ctx, req)
}

// UpdateSession updates a session
func (c *ImportRepositoryClient) UpdateSession(ctx context.Context, session *pb.ImportSession) (*pb.ImportSession, error) {
	return c.client.UpdateSession(ctx, session)
}

// ListSessions lists import sessions with pagination
func (c *ImportRepositoryClient) ListSessions(ctx context.Context, limit, offset int32) (*pb.ListImportSessionsResponse, error) {
	// Convert limit/offset to page/page_size
	pageSize := limit
	if pageSize <= 0 {
		pageSize = 10
	}
	page := (offset / pageSize) + 1

	req := &pb.ListImportSessionsRequest{
		Page:     page,
		PageSize: pageSize,
	}
	return c.client.ListSessions(ctx, req)
}

// AddError adds an error to a session
func (c *ImportRepositoryClient) AddError(ctx context.Context, sessionID string, importError *pb.ImportError) error {
	req := &pb.AddErrorRequest{
		SessionId: sessionID,
		Error:     importError,
	}
	_, err := c.client.AddError(ctx, req)
	return err
}

// GetSessionStatistics retrieves session statistics
func (c *ImportRepositoryClient) GetSessionStatistics(ctx context.Context, accountID, dateFrom, dateTo *string) (*pb.SessionStatistics, error) {
	req := &pb.GetSessionStatisticsRequest{
		AccountId: accountID,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
	}
	return c.client.GetSessionStatistics(ctx, req)
}
