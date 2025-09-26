package repositories

import (
	"context"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ETCMappingRepositoryClient wraps the gRPC client for ETCMappingRepository
type ETCMappingRepositoryClient struct {
	client pb.ETCMappingRepositoryClient
	conn   *grpc.ClientConn
}

// NewETCMappingRepositoryClient creates a new repository client
func NewETCMappingRepositoryClient(address string) (*ETCMappingRepositoryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &ETCMappingRepositoryClient{
		client: pb.NewETCMappingRepositoryClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying connection
func (c *ETCMappingRepositoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Create creates a new mapping
func (c *ETCMappingRepositoryClient) Create(ctx context.Context, mapping *pb.ETCMapping) (*pb.ETCMapping, error) {
	return c.client.Create(ctx, mapping)
}

// GetByID retrieves a mapping by ID
func (c *ETCMappingRepositoryClient) GetByID(ctx context.Context, id int64) (*pb.ETCMapping, error) {
	req := &pb.GetByIDRequest{Id: id}
	return c.client.GetByID(ctx, req)
}

// Update updates a mapping
func (c *ETCMappingRepositoryClient) Update(ctx context.Context, mapping *pb.ETCMapping) (*pb.ETCMapping, error) {
	return c.client.Update(ctx, mapping)
}

// Delete deletes a mapping
func (c *ETCMappingRepositoryClient) Delete(ctx context.Context, id int64) error {
	req := &pb.GetByIDRequest{Id: id}
	_, err := c.client.Delete(ctx, req)
	return err
}

// List lists mappings with pagination
func (c *ETCMappingRepositoryClient) List(ctx context.Context, limit, offset int32) (*pb.ListMappingsResponse, error) {
	// Convert limit/offset to page/page_size
	pageSize := limit
	if pageSize <= 0 {
		pageSize = 10
	}
	page := (offset / pageSize) + 1

	req := &pb.ListMappingsRequest{
		Page:     page,
		PageSize: pageSize,
	}
	return c.client.List(ctx, req)
}

// GetByETCRecordID retrieves mappings by ETC record ID
func (c *ETCMappingRepositoryClient) GetByETCRecordID(ctx context.Context, etcRecordID int64) (*pb.GetMappingsByRecordResponse, error) {
	req := &pb.GetByETCRecordIDRequest{EtcRecordId: etcRecordID}
	return c.client.GetByETCRecordID(ctx, req)
}

// GetByMappedEntity retrieves mappings by mapped entity
func (c *ETCMappingRepositoryClient) GetByMappedEntity(ctx context.Context, entityID int64, entityType string) (*pb.ListMappingsResponse, error) {
	req := &pb.GetByMappedEntityRequest{
		MappedEntityId:   entityID,
		MappedEntityType: entityType,
	}
	return c.client.GetByMappedEntity(ctx, req)
}

// UpdateStatus updates the status of a mapping
func (c *ETCMappingRepositoryClient) UpdateStatus(ctx context.Context, id int64, status pb.MappingStatus) (*pb.ETCMapping, error) {
	req := &pb.UpdateStatusRequest{
		Id:     id,
		Status: status,
	}
	return c.client.UpdateStatus(ctx, req)
}

// BulkCreate creates multiple mappings
func (c *ETCMappingRepositoryClient) BulkCreate(ctx context.Context, mappings []*pb.ETCMapping) (*pb.BulkCreateMappingsResponse, error) {
	req := &pb.BulkCreateMappingsRequest{Mappings: mappings}
	return c.client.BulkCreate(ctx, req)
}

// BulkUpdateStatus updates status for multiple mappings
func (c *ETCMappingRepositoryClient) BulkUpdateStatus(ctx context.Context, ids []int64, status pb.MappingStatus) (*pb.BulkUpdateStatusResponse, error) {
	req := &pb.BulkUpdateStatusRequest{
		Ids:    ids,
		Status: status,
	}
	return c.client.BulkUpdateStatus(ctx, req)
}

// GetPendingMappings retrieves pending mappings
func (c *ETCMappingRepositoryClient) GetPendingMappings(ctx context.Context, limit, offset int32) (*pb.ListMappingsResponse, error) {
	req := &pb.GetPendingMappingsRequest{
		Limit:  limit,
		Offset: offset,
	}
	return c.client.GetPendingMappings(ctx, req)
}

// GetActiveMappings retrieves active mappings
func (c *ETCMappingRepositoryClient) GetActiveMappings(ctx context.Context, limit, offset int32) (*pb.ListMappingsResponse, error) {
	req := &pb.GetActiveMappingsRequest{
		Limit:  limit,
		Offset: offset,
	}
	return c.client.GetActiveMappings(ctx, req)
}

// CountByStatus counts mappings by status
func (c *ETCMappingRepositoryClient) CountByStatus(ctx context.Context) (*pb.CountByStatusResponse, error) {
	req := &pb.CountByStatusRequest{}
	return c.client.CountByStatus(ctx, req)
}

// GetMappingStatistics retrieves mapping statistics
func (c *ETCMappingRepositoryClient) GetMappingStatistics(ctx context.Context, dateFrom, dateTo, mappingType *string) (*pb.MappingStatistics, error) {
	req := &pb.GetMappingStatisticsRequest{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		MappingType: mappingType,
	}
	return c.client.GetMappingStatistics(ctx, req)
}

// SearchMappings searches for mappings
func (c *ETCMappingRepositoryClient) SearchMappings(ctx context.Context, query string, statuses []pb.MappingStatus, mappingTypes []string, limit, offset int32) (*pb.ListMappingsResponse, error) {
	req := &pb.SearchMappingsRequest{
		Query:        query,
		Statuses:     statuses,
		MappingTypes: mappingTypes,
		Limit:        limit,
		Offset:       offset,
	}
	return c.client.SearchMappings(ctx, req)
}