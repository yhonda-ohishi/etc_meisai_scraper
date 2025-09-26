package grpc

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// ETCMappingRepositoryServer implements the ETCMappingRepository gRPC service
type ETCMappingRepositoryServer struct {
	pb.UnimplementedETCMappingRepositoryServer
	mu sync.RWMutex

	// Storage
	mappings map[int64]*pb.ETCMapping

	// Indexes for faster lookups
	mappingsByRecord  map[int64][]int64
	mappingsByEntity  map[string][]int64
	mappingsByStatus  map[pb.MappingStatus][]int64

	// ID generator
	nextMappingID int64
}

// NewETCMappingRepositoryServer creates a new ETCMappingRepository server
func NewETCMappingRepositoryServer() *ETCMappingRepositoryServer {
	return &ETCMappingRepositoryServer{
		mappings:          make(map[int64]*pb.ETCMapping),
		mappingsByRecord:  make(map[int64][]int64),
		mappingsByEntity:  make(map[string][]int64),
		mappingsByStatus:  make(map[pb.MappingStatus][]int64),
		nextMappingID:     1,
	}
}

// Create creates a new mapping
func (s *ETCMappingRepositoryServer) Create(ctx context.Context, mapping *pb.ETCMapping) (*pb.ETCMapping, error) {
	if mapping == nil {
		return nil, status.Error(codes.InvalidArgument, "mapping cannot be nil")
	}

	// Validate required fields
	if mapping.EtcRecordId == 0 {
		return nil, status.Error(codes.InvalidArgument, "ETC record ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Assign new ID and timestamps
	mapping.Id = s.nextMappingID
	s.nextMappingID++

	now := timestamppb.Now()
	mapping.CreatedAt = now
	mapping.UpdatedAt = now

	// Set default status if not specified
	if mapping.Status == pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
		mapping.Status = pb.MappingStatus_MAPPING_STATUS_PENDING
	}

	// Store mapping
	s.mappings[mapping.Id] = copyMapping(mapping)

	// Update indexes
	s.mappingsByRecord[mapping.EtcRecordId] = append(s.mappingsByRecord[mapping.EtcRecordId], mapping.Id)

	entityKey := fmt.Sprintf("%d:%s", mapping.MappedEntityId, mapping.MappedEntityType)
	s.mappingsByEntity[entityKey] = append(s.mappingsByEntity[entityKey], mapping.Id)

	s.mappingsByStatus[mapping.Status] = append(s.mappingsByStatus[mapping.Status], mapping.Id)

	return copyMapping(mapping), nil
}

// GetByID retrieves a mapping by ID
func (s *ETCMappingRepositoryServer) GetByID(ctx context.Context, req *pb.GetByIDRequest) (*pb.ETCMapping, error) {
	if req == nil || req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	mapping, exists := s.mappings[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "mapping with ID %d not found", req.Id)
	}

	return copyMapping(mapping), nil
}

// Update updates an existing mapping
func (s *ETCMappingRepositoryServer) Update(ctx context.Context, mapping *pb.ETCMapping) (*pb.ETCMapping, error) {
	if mapping == nil || mapping.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "mapping with valid ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.mappings[mapping.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "mapping with ID %d not found", mapping.Id)
	}

	// Update fields while preserving immutable fields
	mapping.CreatedAt = existing.CreatedAt
	mapping.UpdatedAt = timestamppb.Now()

	// Update indexes if status changed
	if existing.Status != mapping.Status {
		s.removeFromStatusIndex(existing.Status, mapping.Id)
		s.mappingsByStatus[mapping.Status] = append(s.mappingsByStatus[mapping.Status], mapping.Id)
	}

	// Store updated mapping
	s.mappings[mapping.Id] = copyMapping(mapping)

	return copyMapping(mapping), nil
}

// Delete deletes a mapping
func (s *ETCMappingRepositoryServer) Delete(ctx context.Context, req *pb.GetByIDRequest) (*emptypb.Empty, error) {
	if req == nil || req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	mapping, exists := s.mappings[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "mapping with ID %d not found", req.Id)
	}

	// Remove from indexes
	s.removeFromRecordIndex(mapping.EtcRecordId, req.Id)
	s.removeFromStatusIndex(mapping.Status, req.Id)

	entityKey := fmt.Sprintf("%d:%s", mapping.MappedEntityId, mapping.MappedEntityType)
	s.removeFromEntityIndex(entityKey, req.Id)

	// Delete mapping
	delete(s.mappings, req.Id)

	return &emptypb.Empty{}, nil
}

// List lists mappings with pagination
func (s *ETCMappingRepositoryServer) List(ctx context.Context, req *pb.ListMappingsRequest) (*pb.ListMappingsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter mappings
	var filteredIDs []int64

	if req.Status != nil && *req.Status != pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
		filteredIDs = s.mappingsByStatus[*req.Status]
	} else {
		for id := range s.mappings {
			filteredIDs = append(filteredIDs, id)
		}
	}

	// Apply pagination
	totalCount := int32(len(filteredIDs))

	page := int(req.Page)
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 100 // Default page size
	}
	if page < 1 {
		page = 1 // Default to first page
	}

	offset := (page - 1) * pageSize
	if offset >= len(filteredIDs) {
		return &pb.ListMappingsResponse{
			Mappings:   []*pb.ETCMapping{},
			TotalCount: totalCount,
		}, nil
	}

	end := offset + pageSize
	if end > len(filteredIDs) {
		end = len(filteredIDs)
	}

	// Collect results
	results := make([]*pb.ETCMapping, 0, end-offset)
	for i := offset; i < end; i++ {
		if mapping, exists := s.mappings[filteredIDs[i]]; exists {
			results = append(results, copyMapping(mapping))
		}
	}

	return &pb.ListMappingsResponse{
		Mappings:   results,
		TotalCount: totalCount,
	}, nil
}

// GetByETCRecordID retrieves mappings by ETC record ID
func (s *ETCMappingRepositoryServer) GetByETCRecordID(ctx context.Context, req *pb.GetByETCRecordIDRequest) (*pb.GetMappingsByRecordResponse, error) {
	if req == nil || req.EtcRecordId == 0 {
		return nil, status.Error(codes.InvalidArgument, "ETC record ID is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	mappingIDs := s.mappingsByRecord[req.EtcRecordId]
	results := make([]*pb.ETCMapping, 0, len(mappingIDs))

	for _, id := range mappingIDs {
		if mapping, exists := s.mappings[id]; exists {
			results = append(results, copyMapping(mapping))
		}
	}

	return &pb.GetMappingsByRecordResponse{
		Mappings: results,
	}, nil
}

// GetByMappedEntity retrieves mappings by mapped entity
func (s *ETCMappingRepositoryServer) GetByMappedEntity(ctx context.Context, req *pb.GetByMappedEntityRequest) (*pb.ListMappingsResponse, error) {
	if req == nil || req.MappedEntityId == 0 {
		return nil, status.Error(codes.InvalidArgument, "mapped entity ID is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entityKey := fmt.Sprintf("%d:%s", req.MappedEntityId, req.MappedEntityType)
	mappingIDs := s.mappingsByEntity[entityKey]

	results := make([]*pb.ETCMapping, 0, len(mappingIDs))
	for _, id := range mappingIDs {
		if mapping, exists := s.mappings[id]; exists {
			results = append(results, copyMapping(mapping))
		}
	}

	return &pb.ListMappingsResponse{
		Mappings:   results,
		TotalCount: int32(len(results)),
	}, nil
}

// UpdateStatus updates the status of a mapping
func (s *ETCMappingRepositoryServer) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.ETCMapping, error) {
	if req == nil || req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	mapping, exists := s.mappings[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "mapping with ID %d not found", req.Id)
	}

	// Update status
	oldStatus := mapping.Status
	mapping.Status = req.Status
	mapping.UpdatedAt = timestamppb.Now()

	// Update status index
	s.removeFromStatusIndex(oldStatus, req.Id)
	s.mappingsByStatus[req.Status] = append(s.mappingsByStatus[req.Status], req.Id)

	return copyMapping(mapping), nil
}

// BulkCreate creates multiple mappings
func (s *ETCMappingRepositoryServer) BulkCreate(ctx context.Context, req *pb.BulkCreateMappingsRequest) (*pb.BulkCreateMappingsResponse, error) {
	if req == nil || len(req.Mappings) == 0 {
		return nil, status.Error(codes.InvalidArgument, "mappings are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	results := make([]*pb.ETCMapping, 0, len(req.Mappings))
	now := timestamppb.Now()

	for _, mapping := range req.Mappings {
		// Assign ID and timestamps
		mapping.Id = s.nextMappingID
		s.nextMappingID++
		mapping.CreatedAt = now
		mapping.UpdatedAt = now

		// Set default status
		if mapping.Status == pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
			mapping.Status = pb.MappingStatus_MAPPING_STATUS_PENDING
		}

		// Store mapping
		s.mappings[mapping.Id] = copyMapping(mapping)

		// Update indexes
		s.mappingsByRecord[mapping.EtcRecordId] = append(s.mappingsByRecord[mapping.EtcRecordId], mapping.Id)

		entityKey := fmt.Sprintf("%d:%s", mapping.MappedEntityId, mapping.MappedEntityType)
		s.mappingsByEntity[entityKey] = append(s.mappingsByEntity[entityKey], mapping.Id)

		s.mappingsByStatus[mapping.Status] = append(s.mappingsByStatus[mapping.Status], mapping.Id)

		results = append(results, copyMapping(mapping))
	}

	return &pb.BulkCreateMappingsResponse{
		Mappings:     results,
		CreatedCount: int32(len(results)),
	}, nil
}

// BulkUpdateStatus updates the status of multiple mappings
func (s *ETCMappingRepositoryServer) BulkUpdateStatus(ctx context.Context, req *pb.BulkUpdateStatusRequest) (*pb.BulkUpdateStatusResponse, error) {
	if req == nil || len(req.Ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "IDs are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	updatedCount := int32(0)
	now := timestamppb.Now()

	for _, id := range req.Ids {
		if mapping, exists := s.mappings[id]; exists {
			// Update status
			oldStatus := mapping.Status
			mapping.Status = req.Status
			mapping.UpdatedAt = now

			// Update status index
			s.removeFromStatusIndex(oldStatus, id)
			s.mappingsByStatus[req.Status] = append(s.mappingsByStatus[req.Status], id)

			updatedCount++
		}
	}

	return &pb.BulkUpdateStatusResponse{
		UpdatedCount: updatedCount,
	}, nil
}

// GetPendingMappings retrieves pending mappings
func (s *ETCMappingRepositoryServer) GetPendingMappings(ctx context.Context, req *pb.GetPendingMappingsRequest) (*pb.ListMappingsResponse, error) {
	status := pb.MappingStatus_MAPPING_STATUS_PENDING
	return s.List(ctx, &pb.ListMappingsRequest{
		Status: &status,
	})
}

// GetActiveMappings retrieves active mappings
func (s *ETCMappingRepositoryServer) GetActiveMappings(ctx context.Context, req *pb.GetActiveMappingsRequest) (*pb.ListMappingsResponse, error) {
	status := pb.MappingStatus_MAPPING_STATUS_ACTIVE
	return s.List(ctx, &pb.ListMappingsRequest{
		Status: &status,
	})
}

// CountByStatus counts mappings by status
func (s *ETCMappingRepositoryServer) CountByStatus(ctx context.Context, req *pb.CountByStatusRequest) (*pb.CountByStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[string]int32)

	for status, ids := range s.mappingsByStatus {
		statusName := statusToString(status)
		counts[statusName] = int32(len(ids))
	}

	return &pb.CountByStatusResponse{
		StatusCounts: counts,
	}, nil
}

// GetMappingStatistics retrieves mapping statistics
func (s *ETCMappingRepositoryServer) GetMappingStatistics(ctx context.Context, req *pb.GetMappingStatisticsRequest) (*pb.MappingStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalMappings := int32(len(s.mappings))
	activeMappings := int32(len(s.mappingsByStatus[pb.MappingStatus_MAPPING_STATUS_ACTIVE]))
	pendingMappings := int32(len(s.mappingsByStatus[pb.MappingStatus_MAPPING_STATUS_PENDING]))
	rejectedMappings := int32(len(s.mappingsByStatus[pb.MappingStatus_MAPPING_STATUS_REJECTED]))

	// Calculate average confidence
	var totalConfidence float32
	for _, mapping := range s.mappings {
		totalConfidence += mapping.Confidence
	}

	avgConfidence := float32(0)
	if totalMappings > 0 {
		avgConfidence = totalConfidence / float32(totalMappings)
	}

	// Count by type
	mappingsByType := make(map[string]int32)
	for _, mapping := range s.mappings {
		mappingsByType[mapping.MappingType]++
	}

	return &pb.MappingStatistics{
		TotalMappings:    totalMappings,
		ActiveMappings:   activeMappings,
		PendingMappings:  pendingMappings,
		RejectedMappings: rejectedMappings,
		MappingsByType:   mappingsByType,
		AverageConfidence: avgConfidence,
	}, nil
}

// SearchMappings searches mappings with advanced filters
func (s *ETCMappingRepositoryServer) SearchMappings(ctx context.Context, req *pb.SearchMappingsRequest) (*pb.ListMappingsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*pb.ETCMapping

	for _, mapping := range s.mappings {
		// Apply filters
		if len(req.Statuses) > 0 && !containsStatus(req.Statuses, mapping.Status) {
			continue
		}

		if len(req.MappingTypes) > 0 && !containsString(req.MappingTypes, mapping.MappingType) {
			continue
		}

		if req.Query != "" && !matchesQuery(mapping, req.Query) {
			continue
		}

		results = append(results, copyMapping(mapping))
	}

	// Apply pagination
	totalCount := int32(len(results))

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit == 0 {
		limit = 100
	}

	if offset >= len(results) {
		return &pb.ListMappingsResponse{
			Mappings:   []*pb.ETCMapping{},
			TotalCount: totalCount,
		}, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return &pb.ListMappingsResponse{
		Mappings:   results[offset:end],
		TotalCount: totalCount,
	}, nil
}

// Helper functions

func (s *ETCMappingRepositoryServer) removeFromRecordIndex(recordID, mappingID int64) {
	ids := s.mappingsByRecord[recordID]
	for i, id := range ids {
		if id == mappingID {
			s.mappingsByRecord[recordID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}

func (s *ETCMappingRepositoryServer) removeFromEntityIndex(entityKey string, mappingID int64) {
	ids := s.mappingsByEntity[entityKey]
	for i, id := range ids {
		if id == mappingID {
			s.mappingsByEntity[entityKey] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}

func (s *ETCMappingRepositoryServer) removeFromStatusIndex(status pb.MappingStatus, mappingID int64) {
	ids := s.mappingsByStatus[status]
	for i, id := range ids {
		if id == mappingID {
			s.mappingsByStatus[status] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}

func copyMapping(m *pb.ETCMapping) *pb.ETCMapping {
	if m == nil {
		return nil
	}
	return proto.Clone(m).(*pb.ETCMapping)
}

func statusToString(status pb.MappingStatus) string {
	switch status {
	case pb.MappingStatus_MAPPING_STATUS_ACTIVE:
		return "active"
	case pb.MappingStatus_MAPPING_STATUS_INACTIVE:
		return "inactive"
	case pb.MappingStatus_MAPPING_STATUS_PENDING:
		return "pending"
	case pb.MappingStatus_MAPPING_STATUS_REJECTED:
		return "rejected"
	default:
		return "unspecified"
	}
}

func containsStatus(statuses []pb.MappingStatus, status pb.MappingStatus) bool {
	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func containsString(strings []string, str string) bool {
	for _, s := range strings {
		if s == str {
			return true
		}
	}
	return false
}

func matchesQuery(mapping *pb.ETCMapping, query string) bool {
	// Simple query matching - can be enhanced
	return true
}