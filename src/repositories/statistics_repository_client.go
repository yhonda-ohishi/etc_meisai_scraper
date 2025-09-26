package repositories

import (
	"context"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// StatisticsRepositoryClient wraps the gRPC client for StatisticsRepository
type StatisticsRepositoryClient struct {
	client pb.StatisticsRepositoryClient
	conn   *grpc.ClientConn
}

// NewStatisticsRepositoryClient creates a new repository client
func NewStatisticsRepositoryClient(address string) (*StatisticsRepositoryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &StatisticsRepositoryClient{
		client: pb.NewStatisticsRepositoryClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying connection
func (c *StatisticsRepositoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetOverallStatistics retrieves overall statistics
func (c *StatisticsRepositoryClient) GetOverallStatistics(ctx context.Context, dateFrom, dateTo *string) (*pb.GetStatisticsResponse, error) {
	req := &pb.GetStatisticsRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}
	return c.client.GetOverallStatistics(ctx, req)
}

// GetDailyStatistics retrieves daily statistics
func (c *StatisticsRepositoryClient) GetDailyStatistics(ctx context.Context, dateFrom, dateTo string) (*pb.GetDailyStatisticsResponse, error) {
	req := &pb.GetDailyStatisticsRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}
	return c.client.GetDailyStatistics(ctx, req)
}

// GetICUsageStatistics retrieves IC usage statistics
func (c *StatisticsRepositoryClient) GetICUsageStatistics(ctx context.Context, dateFrom, dateTo *string, topN int32) (*pb.GetICUsageResponse, error) {
	req := &pb.GetICUsageRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
		TopN:     topN,
	}
	return c.client.GetICUsageStatistics(ctx, req)
}

// GetCarUsageStatistics retrieves car usage statistics
func (c *StatisticsRepositoryClient) GetCarUsageStatistics(ctx context.Context, dateFrom, dateTo *string) (*pb.GetCarUsageResponse, error) {
	req := &pb.GetCarUsageRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}
	return c.client.GetCarUsageStatistics(ctx, req)
}

// GetImportStatistics retrieves import statistics
func (c *StatisticsRepositoryClient) GetImportStatistics(ctx context.Context, accountID, dateFrom, dateTo *string) (*pb.GetImportStatisticsResponse, error) {
	req := &pb.GetImportStatisticsRequest{
		AccountId: accountID,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
	}
	return c.client.GetImportStatistics(ctx, req)
}

// GetMappingStatistics retrieves mapping statistics
func (c *StatisticsRepositoryClient) GetMappingStatistics(ctx context.Context, dateFrom, dateTo, mappingType *string) (*pb.MappingStatistics, error) {
	req := &pb.GetMappingStatisticsRequest{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		MappingType: mappingType,
	}
	return c.client.GetMappingStatistics(ctx, req)
}
