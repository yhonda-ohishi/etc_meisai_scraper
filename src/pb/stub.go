// Package pb provides protobuf stubs for gRPC communication
// This is a temporary stub - replace with actual protobuf generated files
package pb

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Temporary stubs - these should be replaced with actual protobuf generated types

// ETCServiceClient provides ETC service operations
type ETCServiceClient interface {
	CreateETCMeisai(ctx context.Context, req *CreateETCMeisaiRequest, opts ...grpc.CallOption) (*ETCMeisaiResponse, error)
	GetETCMeisai(ctx context.Context, req *GetETCMeisaiRequest, opts ...grpc.CallOption) (*ETCMeisaiResponse, error)
	BulkCreateETCMeisai(ctx context.Context, req *BulkCreateETCMeisaiRequest, opts ...grpc.CallOption) (*BulkCreateETCMeisaiResponse, error)
	ListETCMeisai(ctx context.Context, req *ListETCMeisaiRequest, opts ...grpc.CallOption) (*ListETCMeisaiResponse, error)
	GetETCSummary(ctx context.Context, req *GetETCSummaryRequest, opts ...grpc.CallOption) (*ETCSummaryResponse, error)
}

type ETCMappingServiceClient interface {
	CreateMapping(ctx context.Context, req *CreateMappingRequest, opts ...grpc.CallOption) (*MappingResponse, error)
	GetMapping(ctx context.Context, req *GetMappingRequest, opts ...grpc.CallOption) (*MappingResponse, error)
	UpdateMapping(ctx context.Context, req *UpdateMappingRequest, opts ...grpc.CallOption) (*MappingResponse, error)
	DeleteMapping(ctx context.Context, req *DeleteMappingRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	ListMappings(ctx context.Context, req *ListMappingsRequest, opts ...grpc.CallOption) (*ListMappingsResponse, error)
	BulkCreateMappings(ctx context.Context, req *BulkCreateMappingsRequest, opts ...grpc.CallOption) (*BulkCreateMappingsResponse, error)
	DeleteMappingsByETCMeisai(ctx context.Context, req *DeleteMappingsByETCMeisaiRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	UpdateMappingConfidence(ctx context.Context, req *UpdateMappingConfidenceRequest, opts ...grpc.CallOption) (*MappingResponse, error)
	FindPotentialMatches(ctx context.Context, req *FindPotentialMatchesRequest, opts ...grpc.CallOption) (*FindPotentialMatchesResponse, error)
}

type ETCImportServiceClient interface {
	CreateImportBatch(ctx context.Context, req *CreateImportBatchRequest, opts ...grpc.CallOption) (*ImportBatchResponse, error)
	ProcessCSVData(ctx context.Context, req *ProcessCSVDataRequest, opts ...grpc.CallOption) (*ProcessCSVDataResponse, error)
	GetImportProgress(ctx context.Context, req *GetImportProgressRequest, opts ...grpc.CallOption) (*ImportProgressResponse, error)
}

// Request/Response types (stubs)
type CreateETCMeisaiRequest struct {
	UseDate   *timestamppb.Timestamp
	UseTime   string
	EntryIc   string
	ExitIc    string
	Amount    int32
	CarNumber string
	EtcNumber string
}

type ETCMeisaiResponse struct {
	Id        int64
	UseDate   *timestamppb.Timestamp
	UseTime   string
	EntryIc   string
	ExitIc    string
	Amount    int32
	CarNumber string
	EtcNumber string
	Hash      string
	CreatedAt *timestamppb.Timestamp
	UpdatedAt *timestamppb.Timestamp
}

type GetETCMeisaiRequest struct {
	Id int64
}

type BulkCreateETCMeisaiRequest struct {
	Records []*CreateETCMeisaiRequest
}

type BulkCreateETCMeisaiResponse struct {
	Success      bool
	ImportedRows int32
	Message      string
	Errors       []string
}

type ListETCMeisaiRequest struct {
	Limit     int32
	Offset    int32
	FromDate  *timestamppb.Timestamp
	ToDate    *timestamppb.Timestamp
	CarNumber string
	EtcNumber string
}

type ListETCMeisaiResponse struct {
	Records []*ETCMeisaiResponse
	Total   int64
}

type GetETCSummaryRequest struct {
	FromDate *timestamppb.Timestamp
	ToDate   *timestamppb.Timestamp
}

type ETCSummaryResponse struct {
	TotalRecords int64
	TotalAmount  int64
	DateRange    string
}

type CreateMappingRequest struct {
	EtcMeisaiId int64
	DtakoRowId  string
	MappingType string
	Confidence  float32
	Notes       string
}

type MappingResponse struct {
	Id          int64
	EtcMeisaiId int64
	DtakoRowId  string
	MappingType string
	Confidence  float32
	Notes       string
	CreatedAt   *timestamppb.Timestamp
	UpdatedAt   *timestamppb.Timestamp
}

type FindPotentialMatchesRequest struct {
	EtcMeisaiId int64
	Threshold   float32
}

type FindPotentialMatchesResponse struct {
	Matches []*PotentialMatch
}

// Additional mapping request/response types
type GetMappingRequest struct {
	Id int64
}

type UpdateMappingRequest struct {
	Id          int64
	MappingType string
	Confidence  float32
	Notes       string
}

type DeleteMappingRequest struct {
	Id int64
}

type DeleteResponse struct {
	Success bool
	Message string
}

type ListMappingsRequest struct {
	Limit         int32
	Offset        int32
	EtcMeisaiId   int64
	DtakoRowId    string
	MappingType   string
	MinConfidence float32
}

type ListMappingsResponse struct {
	Mappings []*MappingResponse
	Total    int64
}

type BulkCreateMappingsRequest struct {
	Mappings []*CreateMappingRequest
}

type BulkCreateMappingsResponse struct {
	Success    bool
	Message    string
	CreatedIds []int64
}

type DeleteMappingsByETCMeisaiRequest struct {
	EtcMeisaiId int64
}

type UpdateMappingConfidenceRequest struct {
	Id         int64
	Confidence float32
}

type PotentialMatch struct {
	DtakoRowId    string
	Confidence    float32
	MatchReasons  []string
	DtakoData     map[string]interface{}
}

type CreateImportBatchRequest struct {
	FileName    string
	FileSize    int64
	AccountId   string
	ImportType  string
	Status      string
	TotalRows   int64
	ProcessedRows int64
}

type ImportBatchResponse struct {
	Id            int64
	FileName      string
	FileSize      int64
	AccountId     string
	ImportType    string
	Status        string
	TotalRows     int64
	ProcessedRows int64
	SuccessCount  int64
	ErrorCount    int64
	CreatedAt     *timestamppb.Timestamp
	UpdatedAt     *timestamppb.Timestamp
	CompletedAt   *timestamppb.Timestamp
}

type ProcessCSVDataRequest struct {
	BatchId    int64
	CsvContent string
	AccountId  string
}

type ProcessCSVDataResponse struct {
	Status        string
	TotalRows     int64
	ProcessedRows int64
	SuccessCount  int64
	ErrorCount    int64
	UpdatedAt     *timestamppb.Timestamp
}

type GetImportProgressRequest struct {
	BatchId int64
}

type ImportProgressResponse struct {
	BatchId      int64
	Status       string
	TotalRows    int64
	ProcessedRows int64
	SuccessCount int64
	ErrorCount   int64
	Percentage   float32
	Message      string
	UpdatedAt    *timestamppb.Timestamp
}

// Client constructors (stubs)
func NewETCServiceClient(conn grpc.ClientConnInterface) ETCServiceClient {
	return &stubETCServiceClient{}
}

func NewETCMappingServiceClient(conn grpc.ClientConnInterface) ETCMappingServiceClient {
	return &stubETCMappingServiceClient{}
}

func NewETCImportServiceClient(conn grpc.ClientConnInterface) ETCImportServiceClient {
	return &stubETCImportServiceClient{}
}

// Stub implementations
type stubETCServiceClient struct{}

func (c *stubETCServiceClient) CreateETCMeisai(ctx context.Context, req *CreateETCMeisaiRequest, opts ...grpc.CallOption) (*ETCMeisaiResponse, error) {
	return &ETCMeisaiResponse{
		Id:        1,
		UseDate:   req.UseDate,
		UseTime:   req.UseTime,
		EntryIc:   req.EntryIc,
		ExitIc:    req.ExitIc,
		Amount:    req.Amount,
		CarNumber: req.CarNumber,
		EtcNumber: req.EtcNumber,
		Hash:      "stub_hash",
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}, nil
}

func (c *stubETCServiceClient) GetETCMeisai(ctx context.Context, req *GetETCMeisaiRequest, opts ...grpc.CallOption) (*ETCMeisaiResponse, error) {
	return &ETCMeisaiResponse{
		Id: req.Id,
	}, nil
}

func (c *stubETCServiceClient) BulkCreateETCMeisai(ctx context.Context, req *BulkCreateETCMeisaiRequest, opts ...grpc.CallOption) (*BulkCreateETCMeisaiResponse, error) {
	return &BulkCreateETCMeisaiResponse{
		Success:      true,
		ImportedRows: int32(len(req.Records)),
		Message:      "Bulk import completed (stub)",
	}, nil
}

func (c *stubETCServiceClient) ListETCMeisai(ctx context.Context, req *ListETCMeisaiRequest, opts ...grpc.CallOption) (*ListETCMeisaiResponse, error) {
	return &ListETCMeisaiResponse{
		Records: []*ETCMeisaiResponse{},
		Total:   0,
	}, nil
}

func (c *stubETCServiceClient) GetETCSummary(ctx context.Context, req *GetETCSummaryRequest, opts ...grpc.CallOption) (*ETCSummaryResponse, error) {
	return &ETCSummaryResponse{
		TotalRecords: 0,
		TotalAmount:  0,
		DateRange:    "stub range",
	}, nil
}

type stubETCMappingServiceClient struct{}

func (c *stubETCMappingServiceClient) CreateMapping(ctx context.Context, req *CreateMappingRequest, opts ...grpc.CallOption) (*MappingResponse, error) {
	return &MappingResponse{
		Id:          1,
		EtcMeisaiId: req.EtcMeisaiId,
		DtakoRowId:  req.DtakoRowId,
		MappingType: req.MappingType,
		Confidence:  req.Confidence,
		Notes:       req.Notes,
		CreatedAt:   timestamppb.New(time.Now()),
		UpdatedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (c *stubETCMappingServiceClient) GetMapping(ctx context.Context, req *GetMappingRequest, opts ...grpc.CallOption) (*MappingResponse, error) {
	return &MappingResponse{
		Id:          req.Id,
		EtcMeisaiId: 0,
		DtakoRowId:  "",
		MappingType: "manual",
		Confidence:  1.0,
	}, nil
}

func (c *stubETCMappingServiceClient) UpdateMapping(ctx context.Context, req *UpdateMappingRequest, opts ...grpc.CallOption) (*MappingResponse, error) {
	return &MappingResponse{
		Id:          req.Id,
		MappingType: req.MappingType,
		Confidence:  req.Confidence,
		UpdatedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (c *stubETCMappingServiceClient) DeleteMapping(ctx context.Context, req *DeleteMappingRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	return &DeleteResponse{
		Success: true,
		Message: "Mapping deleted successfully",
	}, nil
}

func (c *stubETCMappingServiceClient) ListMappings(ctx context.Context, req *ListMappingsRequest, opts ...grpc.CallOption) (*ListMappingsResponse, error) {
	return &ListMappingsResponse{
		Mappings: []*MappingResponse{},
		Total:    0,
	}, nil
}

func (c *stubETCMappingServiceClient) BulkCreateMappings(ctx context.Context, req *BulkCreateMappingsRequest, opts ...grpc.CallOption) (*BulkCreateMappingsResponse, error) {
	var createdIds []int64
	for i := range req.Mappings {
		createdIds = append(createdIds, int64(i+1))
	}
	return &BulkCreateMappingsResponse{
		Success:    true,
		Message:    "Bulk mappings created successfully",
		CreatedIds: createdIds,
	}, nil
}

func (c *stubETCMappingServiceClient) DeleteMappingsByETCMeisai(ctx context.Context, req *DeleteMappingsByETCMeisaiRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	return &DeleteResponse{
		Success: true,
		Message: "Mappings deleted successfully",
	}, nil
}

func (c *stubETCMappingServiceClient) UpdateMappingConfidence(ctx context.Context, req *UpdateMappingConfidenceRequest, opts ...grpc.CallOption) (*MappingResponse, error) {
	return &MappingResponse{
		Id:         req.Id,
		Confidence: req.Confidence,
		UpdatedAt:  timestamppb.New(time.Now()),
	}, nil
}

func (c *stubETCMappingServiceClient) FindPotentialMatches(ctx context.Context, req *FindPotentialMatchesRequest, opts ...grpc.CallOption) (*FindPotentialMatchesResponse, error) {
	return &FindPotentialMatchesResponse{
		Matches: []*PotentialMatch{},
	}, nil
}

type stubETCImportServiceClient struct{}

func (c *stubETCImportServiceClient) CreateImportBatch(ctx context.Context, req *CreateImportBatchRequest, opts ...grpc.CallOption) (*ImportBatchResponse, error) {
	return &ImportBatchResponse{
		Id:          1,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		AccountId:   req.AccountId,
		ImportType:  req.ImportType,
		Status:      req.Status,
		TotalRows:   req.TotalRows,
		ProcessedRows: req.ProcessedRows,
		CreatedAt:   timestamppb.New(time.Now()),
		UpdatedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (c *stubETCImportServiceClient) ProcessCSVData(ctx context.Context, req *ProcessCSVDataRequest, opts ...grpc.CallOption) (*ProcessCSVDataResponse, error) {
	return &ProcessCSVDataResponse{
		Status:        "completed",
		TotalRows:     100,
		ProcessedRows: 100,
		SuccessCount:  100,
		ErrorCount:    0,
		UpdatedAt:     timestamppb.New(time.Now()),
	}, nil
}

func (c *stubETCImportServiceClient) GetImportProgress(ctx context.Context, req *GetImportProgressRequest, opts ...grpc.CallOption) (*ImportProgressResponse, error) {
	return &ImportProgressResponse{
		BatchId:       req.BatchId,
		Status:        "completed",
		TotalRows:     100,
		ProcessedRows: 100,
		SuccessCount:  100,
		ErrorCount:    0,
		Percentage:    100.0,
		Message:       "Import completed (stub)",
		UpdatedAt:     timestamppb.New(time.Now()),
	}, nil
}