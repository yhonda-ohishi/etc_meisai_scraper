package mocks

import (
	"testing"
	"github.com/stretchr/testify/mock"
)

// MockRegistry manages all mock objects in a test
type MockRegistry struct {
	mocks map[string]*mock.Mock
	t     *testing.T
}

// NewMockRegistry creates a new mock registry
func NewMockRegistry(t *testing.T) *MockRegistry {
	return &MockRegistry{
		mocks: make(map[string]*mock.Mock),
		t:     t,
	}
}

// Register adds a mock to the registry
func (r *MockRegistry) Register(name string, m *mock.Mock) {
	r.mocks[name] = m
}

// Get retrieves a mock from the registry
func (r *MockRegistry) Get(name string) *mock.Mock {
	return r.mocks[name]
}

// AssertAllExpectations verifies all registered mocks
func (r *MockRegistry) AssertAllExpectations() {
	for name, m := range r.mocks {
		if !m.AssertExpectations(r.t) {
			r.t.Errorf("Mock %s: expectations not met", name)
		}
	}
}

// Reset clears all mock expectations
func (r *MockRegistry) Reset() {
	for _, m := range r.mocks {
		m.ExpectedCalls = nil
		m.Calls = nil
	}
}