package grpc

import (
	"context"
	"sync"

	"github.com/google/uuid"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImportRepositoryServer implements the ImportRepository service
type ImportRepositoryServer struct {
	pb.UnimplementedImportRepositoryServer
	mu        sync.RWMutex
	sessions  map[string]*pb.ImportSession
	errors    map[string][]*pb.ImportError
	errorsSeq map[string]int64 // Sequential ID generator for errors per session
}

// NewImportRepositoryServer creates a new import repository server instance
func NewImportRepositoryServer() *ImportRepositoryServer {
	return &ImportRepositoryServer{
		sessions:  make(map[string]*pb.ImportSession),
		errors:    make(map[string][]*pb.ImportError),
		errorsSeq: make(map[string]int64),
	}
}

// CreateSession creates a new import session
func (s *ImportRepositoryServer) CreateSession(ctx context.Context, req *pb.ImportSession) (*pb.ImportSession, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "session is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate session ID if not provided
	if req.Id == "" {
		req.Id = uuid.New().String()
	}

	// Check if session already exists
	if _, exists := s.sessions[req.Id]; exists {
		return nil, status.Error(codes.AlreadyExists, "session already exists")
	}

	// Set timestamps
	now := timestamppb.Now()
	req.StartedAt = now
	req.CreatedAt = now

	// Initialize status if not set
	if req.Status == pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED {
		req.Status = pb.ImportStatus_IMPORT_STATUS_PENDING
	}

	// Store session
	s.sessions[req.Id] = req
	s.errors[req.Id] = make([]*pb.ImportError, 0)
	s.errorsSeq[req.Id] = 1

	return req, nil
}

// GetSession retrieves a session by ID
func (s *ImportRepositoryServer) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.ImportSession, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[req.SessionId]
	if !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	// Include error count in session
	if errors, ok := s.errors[req.SessionId]; ok {
		session.ErrorRows = int32(len(errors))
	}

	return session, nil
}

// UpdateSession updates an existing session
func (s *ImportRepositoryServer) UpdateSession(ctx context.Context, req *pb.ImportSession) (*pb.ImportSession, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "valid session with ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.sessions[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	// Preserve immutable fields
	req.CreatedAt = existing.CreatedAt
	req.StartedAt = existing.StartedAt

	// Set completed_at if status changed to completed or failed
	if (req.Status == pb.ImportStatus_IMPORT_STATUS_COMPLETED || req.Status == pb.ImportStatus_IMPORT_STATUS_FAILED) &&
		req.CompletedAt == nil {
		req.CompletedAt = timestamppb.Now()
	}

	// Update session
	s.sessions[req.Id] = req

	return req, nil
}

// ListSessions lists import sessions with optional filtering
func (s *ImportRepositoryServer) ListSessions(ctx context.Context, req *pb.ListImportSessionsRequest) (*pb.ListImportSessionsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filteredSessions []*pb.ImportSession

	for _, session := range s.sessions {
		// Apply filters
		if req.AccountId != nil && session.AccountId != *req.AccountId {
			continue
		}
		if req.Status != nil && session.Status != *req.Status {
			continue
		}

		// Note: Date filtering not implemented in this simple in-memory repository

		// Include error count
		if errors, ok := s.errors[session.Id]; ok {
			session.ErrorRows = int32(len(errors))
		}

		filteredSessions = append(filteredSessions, session)
	}

	// Apply pagination
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 100
	}
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	start := offset
	if start > len(filteredSessions) {
		start = len(filteredSessions)
	}
	end := start + pageSize
	if end > len(filteredSessions) {
		end = len(filteredSessions)
	}

	return &pb.ListImportSessionsResponse{
		Sessions:   filteredSessions[start:end],
		TotalCount: int32(len(filteredSessions)),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}

// AddError adds an error to a session
func (s *ImportRepositoryServer) AddError(ctx context.Context, req *pb.AddErrorRequest) (*emptypb.Empty, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}
	if req.Error == nil {
		return nil, status.Error(codes.InvalidArgument, "error is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify session exists
	if _, ok := s.sessions[req.SessionId]; !ok {
		return nil, status.Error(codes.NotFound, "session not found")
	}

	// Generate sequential row number for this session
	if req.Error.RowNumber == 0 {
		req.Error.RowNumber = int32(s.errorsSeq[req.SessionId])
		s.errorsSeq[req.SessionId]++
	}

	// Add error to session
	s.errors[req.SessionId] = append(s.errors[req.SessionId], req.Error)

	// Update session error count
	if session, ok := s.sessions[req.SessionId]; ok {
		session.ErrorRows = int32(len(s.errors[req.SessionId]))
	}

	return &emptypb.Empty{}, nil
}

// GetSessionStatistics retrieves import session statistics
func (s *ImportRepositoryServer) GetSessionStatistics(ctx context.Context, req *pb.GetSessionStatisticsRequest) (*pb.SessionStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &pb.SessionStatistics{}

	for _, session := range s.sessions {
		// Apply filters
		if req.AccountId != nil && session.AccountId != *req.AccountId {
			continue
		}

		// Note: Date filtering not implemented in this simple in-memory repository

		// Count sessions by status
		stats.TotalSessions++
		switch session.Status {
		case pb.ImportStatus_IMPORT_STATUS_COMPLETED:
			stats.SuccessfulSessions++
			stats.TotalRecordsImported += session.SuccessRows
		case pb.ImportStatus_IMPORT_STATUS_FAILED:
			stats.FailedSessions++
		}

		// Count duplicates and errors
		stats.TotalDuplicates += session.DuplicateRows
		if errors, ok := s.errors[session.Id]; ok {
			stats.TotalErrors += int32(len(errors))
		}
	}

	return stats, nil
}