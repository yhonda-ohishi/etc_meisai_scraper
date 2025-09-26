package repositories

import (
	"context"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ETCMeisaiRecordRepositoryClient wraps the gRPC client for ETCMeisaiRecordRepository
type ETCMeisaiRecordRepositoryClient struct {
	client pb.ETCMeisaiRecordRepositoryClient
	conn   *grpc.ClientConn
}

// NewETCMeisaiRecordRepositoryClient creates a new repository client
func NewETCMeisaiRecordRepositoryClient(address string) (*ETCMeisaiRecordRepositoryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &ETCMeisaiRecordRepositoryClient{
		client: pb.NewETCMeisaiRecordRepositoryClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying connection
func (c *ETCMeisaiRecordRepositoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Create creates a new ETC record
func (c *ETCMeisaiRecordRepositoryClient) Create(ctx context.Context, record *pb.ETCMeisaiRecord) (*pb.ETCMeisaiRecord, error) {
	return c.client.Create(ctx, record)
}

// GetByID retrieves a record by ID
func (c *ETCMeisaiRecordRepositoryClient) GetByID(ctx context.Context, id int64) (*pb.ETCMeisaiRecord, error) {
	req := &pb.GetByIDRequest{Id: id}
	return c.client.GetByID(ctx, req)
}

// GetByHash retrieves a record by hash
func (c *ETCMeisaiRecordRepositoryClient) GetByHash(ctx context.Context, hash string) (*pb.ETCMeisaiRecord, error) {
	req := &pb.GetByHashRequest{Hash: hash}
	return c.client.GetByHash(ctx, req)
}

// Update updates a record
func (c *ETCMeisaiRecordRepositoryClient) Update(ctx context.Context, record *pb.ETCMeisaiRecord) (*pb.ETCMeisaiRecord, error) {
	return c.client.Update(ctx, record)
}

// Delete deletes a record
func (c *ETCMeisaiRecordRepositoryClient) Delete(ctx context.Context, id int64) error {
	req := &pb.GetByIDRequest{Id: id}
	_, err := c.client.Delete(ctx, req)
	return err
}

// List lists records with pagination
func (c *ETCMeisaiRecordRepositoryClient) List(ctx context.Context, limit, offset int32) (*pb.ListRecordsResponse, error) {
	// Convert limit/offset to page/page_size
	pageSize := limit
	if pageSize <= 0 {
		pageSize = 10
	}
	page := (offset / pageSize) + 1

	req := &pb.ListRecordsRequest{
		Page:     page,
		PageSize: pageSize,
	}
	return c.client.List(ctx, req)
}

// GetByDateRange retrieves records by date range
func (c *ETCMeisaiRecordRepositoryClient) GetByDateRange(ctx context.Context, dateFrom, dateTo string, limit, offset int32) (*pb.ListRecordsResponse, error) {
	// Note: GetByDateRangeRequest uses limit/offset directly
	req := &pb.GetByDateRangeRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Limit:    limit,
		Offset:   offset,
	}
	return c.client.GetByDateRange(ctx, req)
}

// GetByCarNumber retrieves records by car number
func (c *ETCMeisaiRecordRepositoryClient) GetByCarNumber(ctx context.Context, carNumber string, limit, offset int32) (*pb.ListRecordsResponse, error) {
	// Note: GetByCarNumberRequest uses limit/offset directly
	req := &pb.GetByCarNumberRequest{
		CarNumber: carNumber,
		Limit:     limit,
		Offset:    offset,
	}
	return c.client.GetByCarNumber(ctx, req)
}

// GetByETCCard retrieves records by ETC card number
func (c *ETCMeisaiRecordRepositoryClient) GetByETCCard(ctx context.Context, etcCardNumber string, limit, offset int32) (*pb.ListRecordsResponse, error) {
	// Note: GetByETCCardRequest uses limit/offset directly
	req := &pb.GetByETCCardRequest{
		EtcCardNumber: etcCardNumber,
		Limit:         limit,
		Offset:        offset,
	}
	return c.client.GetByETCCard(ctx, req)
}

// BulkCreate creates multiple records
func (c *ETCMeisaiRecordRepositoryClient) BulkCreate(ctx context.Context, records []*pb.ETCMeisaiRecord) (*pb.BulkCreateRecordsResponse, error) {
	req := &pb.BulkCreateRecordsRequest{Records: records}
	return c.client.BulkCreate(ctx, req)
}

// CheckDuplicate checks for duplicate records
func (c *ETCMeisaiRecordRepositoryClient) CheckDuplicate(ctx context.Context, hash string) (*pb.CheckDuplicateResponse, error) {
	req := &pb.CheckDuplicateRequest{Hash: hash}
	return c.client.CheckDuplicate(ctx, req)
}

// GetRecordStatistics retrieves record statistics
func (c *ETCMeisaiRecordRepositoryClient) GetRecordStatistics(ctx context.Context, dateFrom, dateTo *string) (*pb.RecordStatistics, error) {
	req := &pb.GetRecordStatisticsRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}
	return c.client.GetRecordStatistics(ctx, req)
}
