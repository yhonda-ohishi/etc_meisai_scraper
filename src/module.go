package etc_meisai

import (
	"database/sql"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// Module represents the ETC meisai module
type Module struct {
	Handler *handlers.ETCHandler
	Service *services.ETCService
	Repo    *repositories.ETCRepository
}

// NewModule creates a new ETC meisai module with all dependencies
func NewModule(db *sql.DB) *Module {
	repo := repositories.NewETCRepository(db)
	service := services.NewETCService(repo)
	handler := handlers.NewETCHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
		Repo:    repo,
	}
}

// InitializeWithRouter initializes the module with a chi router
func InitializeWithRouter(db *sql.DB, r *chi.Mux) (*Module, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	module := NewModule(db)

	// TODO: Setup routes if needed

	return module, nil
}

// GetHandler returns the HTTP handler for external use
func (m *Module) GetHandler() *handlers.ETCHandler {
	return m.Handler
}

// GetService returns the service layer for external use
func (m *Module) GetService() *services.ETCService {
	return m.Service
}

// GetRepository returns the repository layer for external use
func (m *Module) GetRepository() *repositories.ETCRepository {
	return m.Repo
}