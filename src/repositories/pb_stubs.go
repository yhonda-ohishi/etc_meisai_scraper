package repositories

import (
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// Stub types for missing protobuf messages that are used in tests
// These would normally be generated from proto files

// ETCRecord represents an ETC record in protobuf format
type ETCRecord struct {
	Id            uint64
	Date          string
	Time          string
	EntranceIc    string
	ExitIc        string
	TollAmount    int64
	CarNumber     string
	EtcCardNumber string
	CreatedAt     int64
	UpdatedAt     int64
}

// ETCMapping represents an ETC mapping in protobuf format
type ETCMapping struct {
	Id               uint64
	EtcRecordId      uint64
	MappingType      string
	MappedEntityId   uint64
	MappedEntityType string
	Confidence       float64
	Status           string
	CreatedAt        int64
	UpdatedAt        int64
}

// MappingStatsProto represents mapping statistics in protobuf format
type MappingStatsProto struct {
	TotalMappings    int64
	ActiveMappings   int64
	PendingMappings  int64
	RejectedMappings int64
	AvgConfidence    float64
	ByType           map[string]int64
}

// Bulk operations request/response stubs
type BulkCreateETCRecordsRequest struct {
	Records []*pb.CreateRecordRequest
}

type BulkCreateETCRecordsResponse struct {
	Success      bool
	CreatedCount int64
	Error        string
}

// Search operations request/response stubs
type SearchETCRecordsRequest struct {
	EtcCardNumber string
	CarNumber     string
	EntranceIc    string
	ExitIc        string
	DateFrom      string
	DateTo        string
	MinAmount     int64
	MaxAmount     int64
}

type SearchETCRecordsResponse struct {
	Success bool
	Records []*ETCRecord
	Total   int64
	Error   string
}

// Health check request/response stubs
type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status string
	Error  string
}

// Find mappings by ETC record request/response stubs
type FindMappingsByETCRecordRequest struct {
	EtcRecordId uint64
}

type FindMappingsByETCRecordResponse struct {
	Success  bool
	Mappings []*ETCMapping
	Total    int64
	Error    string
}

// Update mapping status request/response stubs
type UpdateMappingStatusRequest struct {
	Id     uint64
	Status string
}

type UpdateMappingStatusResponse struct {
	Success bool
	Error   string
}

// Bulk update mapping status request/response stubs
type BulkUpdateMappingStatusRequest struct {
	MappingIds []uint64
	Status     string
}

type BulkUpdateMappingStatusResponse struct {
	Success      bool
	UpdatedCount int64
	Error        string
}

// Get mapping stats request/response stubs
type GetMappingStatsRequest struct{}

type GetMappingStatsResponse struct {
	Success bool
	Stats   *MappingStatsProto
	Error   string
}

// Create auto mapping request/response stubs
type CreateAutoMappingRequest struct {
	EtcRecordId uint64
}

type CreateAutoMappingResponse struct {
	Success bool
	Mapping *ETCMapping
	Error   string
}