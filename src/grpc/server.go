package grpc

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	grpcRepo "github.com/yhonda-ohishi/etc_meisai/src/repositories/grpc"
)

// Server represents the gRPC server
type Server struct {
	grpcServer *grpc.Server
	logger     *log.Logger
	port       string
}

// NewServer creates a new gRPC server instance
func NewServer(port string, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.Default()
	}

	return &Server{
		port:   port,
		logger: logger,
	}
}

// Start initializes and starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	// Create gRPC server
	s.grpcServer = grpc.NewServer()

	// Register repository services
	s.registerRepositoryServices()

	// Register business services
	s.registerBusinessServices()

	// Register reflection service for grpcurl/debugging
	reflection.Register(s.grpcServer)

	s.logger.Printf("[gRPC] Server starting on port %s", s.port)

	// Start serving
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// registerRepositoryServices registers all repository services
func (s *Server) registerRepositoryServices() {
	// Create repository server instances
	etcMappingRepo := grpcRepo.NewETCMappingRepositoryServer()
	etcMeisaiRecordRepo := grpcRepo.NewETCMeisaiRecordRepositoryServer()
	importRepo := grpcRepo.NewImportRepositoryServer()

	// Create statistics repository with dependencies
	statsRepo := grpcRepo.NewStatisticsRepositoryServer(
		etcMappingRepo,
		etcMeisaiRecordRepo,
		importRepo,
	)

	// Register repository services with gRPC server
	pb.RegisterETCMappingRepositoryServer(s.grpcServer, etcMappingRepo)
	pb.RegisterETCMeisaiRecordRepositoryServer(s.grpcServer, etcMeisaiRecordRepo)
	pb.RegisterImportRepositoryServer(s.grpcServer, importRepo)
	pb.RegisterStatisticsRepositoryServer(s.grpcServer, statsRepo)

	s.logger.Println("[gRPC] Registered repository services:")
	s.logger.Println("  - ETCMappingRepository (15 methods)")
	s.logger.Println("  - ETCMeisaiRecordRepository (12 methods)")
	s.logger.Println("  - ImportRepository (6 methods)")
	s.logger.Println("  - StatisticsRepository (6 methods)")
}

// registerBusinessServices registers all business layer services
func (s *Server) registerBusinessServices() {
	// TODO: Implement proper business service registration
	// Currently using stub server that returns "not implemented" errors
	//
	// The business services require repository clients (not servers) and proper gRPC connections:
	// - NewMappingBusinessServiceServer expects *repositories.ETCMappingRepositoryClient
	// - NewMeisaiBusinessServiceServer expects repository clients with gRPC addresses
	//
	// For now, business services are handled by the stub server in etc_meisai_server.go

	s.logger.Println("[gRPC] Business services:")
	s.logger.Println("  - Using stub implementations (returns 'not implemented' errors)")
	s.logger.Println("  - TODO: Implement proper repository client connections")
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.logger.Println("[gRPC] Stopping server...")
		s.grpcServer.GracefulStop()
		s.logger.Println("[gRPC] Server stopped")
	}
}

// GetServer returns the underlying gRPC server instance
func (s *Server) GetServer() *grpc.Server {
	return s.grpcServer
}