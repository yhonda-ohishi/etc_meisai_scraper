package repositories

import (
	// "context" // Commented out - not used when clients package is deleted
	"fmt"
	"time"

	// "github.com/yhonda-ohishi/etc_meisai/src/clients" // Commented out - clients package deleted
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	// "github.com/yhonda-ohishi/etc_meisai/src/pb" // Commented out - not used when clients package is deleted
	// "google.golang.org/protobuf/types/known/timestamppb" // Commented out - not used when clients package is deleted
)

// GRPCRepository implements ETCRepository interface using gRPC client only
type GRPCRepository struct {
	client interface{} // TODO: Replace with proper type when clients package is restored
}

// NewGRPCRepository creates a new gRPC-only repository
func NewGRPCRepository(client interface{}) ETCRepository {
	return &GRPCRepository{
		client: client,
	}
}

// Create creates a new ETC record via gRPC
func (r *GRPCRepository) Create(etc *models.ETCMeisai) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.CreateETCMeisaiRequest{
	//	UseDate:   timestamppb.New(etc.UseDate),
	//	UseTime:   etc.UseTime,
	//	EntryIc:   etc.EntryIC,
	//	ExitIc:    etc.ExitIC,
	//	Amount:    etc.Amount,
	//	CarNumber: etc.CarNumber,
	//	EtcNumber: etc.ETCNumber,
	// }
	//
	// resp, err := r.client.CreateETCMeisai(ctx, req)
	// if err != nil {
	//	return fmt.Errorf("gRPC create failed: %w", err)
	// }
	//
	// // Update the model with response data
	// etc.ID = resp.Id
	// etc.Hash = resp.Hash
	// etc.CreatedAt = resp.CreatedAt.AsTime()
	// etc.UpdatedAt = resp.UpdatedAt.AsTime()
	return fmt.Errorf("CreateETCMeisai not available - clients package deleted")
}

// GetByID retrieves an ETC record by ID via gRPC
func (r *GRPCRepository) GetByID(id int64) (*models.ETCMeisai, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.GetETCMeisaiRequest{
	//	Id: id,
	// }
	//
	// resp, err := r.client.GetETCMeisai(ctx, req)
	// if err != nil {
	//	return nil, fmt.Errorf("gRPC get failed: %w", err)
	// }
	//
	// // Convert response to model
	// etc := &models.ETCMeisai{
	//	ID:        resp.Id,
	//	UseDate:   resp.UseDate.AsTime(),
	//	UseTime:   resp.UseTime,
	//	EntryIC:   resp.EntryIc,
	//	ExitIC:    resp.ExitIc,
	//	Amount:    resp.Amount,
	//	CarNumber: resp.CarNumber,
	//	ETCNumber: resp.EtcNumber,
	//	Hash:      resp.Hash,
	//	CreatedAt: resp.CreatedAt.AsTime(),
	//	UpdatedAt: resp.UpdatedAt.AsTime(),
	// }
	//
	// return etc, nil
	return nil, fmt.Errorf("GetETCMeisai not available - clients package deleted")
}

// Update updates an existing ETC record (not supported in gRPC mode)
func (r *GRPCRepository) Update(etc *models.ETCMeisai) error {
	// Update operations should be implemented in db_service
	return fmt.Errorf("update operation not supported in gRPC-only mode")
}

// Delete deletes an ETC record (not supported in gRPC mode)
func (r *GRPCRepository) Delete(id int64) error {
	// Delete operations should be implemented in db_service
	return fmt.Errorf("delete operation not supported in gRPC-only mode")
}

// GetByDateRange retrieves records within a date range via gRPC
func (r *GRPCRepository) GetByDateRange(from, to time.Time) ([]*models.ETCMeisai, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.ListETCMeisaiRequest{
	//	FromDate: timestamppb.New(from),
	//	ToDate:   timestamppb.New(to),
	//	Limit:    1000,
	// }
	return nil, fmt.Errorf("ListETCMeisai not available - clients package deleted")
}

// List retrieves records with pagination via gRPC
func (r *GRPCRepository) List(params *models.ETCListParams) ([]*models.ETCMeisai, int64, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.ListETCMeisaiRequest{
	//	Limit:     int32(params.Limit),
	//	Offset:    int32(params.Offset),
	//	CarNumber: params.CarNumber,
	//	EtcNumber: params.ETCNumber,
	// }
	//
	// if params.StartDate != nil {
	//	req.FromDate = timestamppb.New(*params.StartDate)
	// }
	// if params.EndDate != nil {
	//	req.ToDate = timestamppb.New(*params.EndDate)
	// }
	return nil, 0, fmt.Errorf("ListETCMeisai not available - clients package deleted")
}

// GetByHash retrieves a record by its hash via gRPC
func (r *GRPCRepository) GetByHash(hash string) (*models.ETCMeisai, error) {
	// Use GetETCMeisaiByHash if available in proto, otherwise use List with hash filter
	// For now, we'll implement using a stub
	// In a real implementation, this would call db_service's GetETCMeisaiByHash method
	return nil, fmt.Errorf("GetByHash not yet implemented in db_service")
}

// BulkInsert creates multiple records via gRPC
func (r *GRPCRepository) BulkInsert(records []*models.ETCMeisai) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// var pbRecords []*pb.CreateETCMeisaiRequest
	// for _, record := range records {
	//	pbRecord := &pb.CreateETCMeisaiRequest{
	//		UseDate:   timestamppb.New(record.UseDate),
	//		UseTime:   record.UseTime,
	//		EntryIc:   record.EntryIC,
	//		ExitIc:    record.ExitIC,
	//		Amount:    record.Amount,
	//		CarNumber: record.CarNumber,
	//		EtcNumber: record.ETCNumber,
	//	}
	//	pbRecords = append(pbRecords, pbRecord)
	// }
	//
	// req := &pb.BulkCreateETCMeisaiRequest{
	//	Records: pbRecords,
	// }
	//
	// resp, err := r.client.BulkCreateETCMeisai(ctx, req)
	// if err != nil {
	//	return fmt.Errorf("gRPC bulk create failed: %w", err)
	// }
	//
	// if !resp.Success {
	//	return fmt.Errorf("bulk insert failed: %s", resp.Message)
	// }
	return fmt.Errorf("BulkCreateETCMeisai not available - clients package deleted")
}

// CheckDuplicatesByHash checks for duplicates (delegated to db_service)
func (r *GRPCRepository) CheckDuplicatesByHash(hashes []string) (map[string]bool, error) {
	// This should be implemented in db_service
	// For now, return empty map (no duplicates)
	result := make(map[string]bool)
	for _, hash := range hashes {
		result[hash] = false
	}
	return result, nil
}

// CountByDateRange counts records in a date range via gRPC
func (r *GRPCRepository) CountByDateRange(from, to time.Time) (int64, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.GetETCSummaryRequest{
	//	FromDate: timestamppb.New(from),
	//	ToDate:   timestamppb.New(to),
	// }
	return 0, fmt.Errorf("GetETCSummary not available - clients package deleted")
}

// GetByETCNumber retrieves records by ETC number via gRPC
func (r *GRPCRepository) GetByETCNumber(etcNumber string, limit int) ([]*models.ETCMeisai, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.ListETCMeisaiRequest{
	//	EtcNumber: etcNumber,
	//	Limit:     int32(limit),
	// }
	return nil, fmt.Errorf("ListETCMeisai by ETC number not available - clients package deleted")
}

// GetByCarNumber retrieves records by car number via gRPC
func (r *GRPCRepository) GetByCarNumber(carNumber string, limit int) ([]*models.ETCMeisai, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.ListETCMeisaiRequest{
	//	CarNumber: carNumber,
	//	Limit:     int32(limit),
	// }
	return nil, fmt.Errorf("ListETCMeisai by car number not available - clients package deleted")
}

// GetSummaryByDateRange gets aggregated summary via gRPC
func (r *GRPCRepository) GetSummaryByDateRange(from, to time.Time) (*models.ETCSummary, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.GetETCSummaryRequest{
	//	FromDate: timestamppb.New(from),
	//	ToDate:   timestamppb.New(to),
	// }
	return nil, fmt.Errorf("GetETCSummary not available - clients package deleted")
}