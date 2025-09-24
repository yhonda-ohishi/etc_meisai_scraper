package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockRegistry manages all mock objects for coordinated testing
type MockRegistry struct {
	mocks []interface{}
}

// NewMockRegistry creates a new mock registry
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		mocks: make([]interface{}, 0),
	}
}

// Register adds a mock to the registry
func (r *MockRegistry) Register(mockObj interface{}) {
	r.mocks = append(r.mocks, mockObj)
}

// AssertExpectations verifies all registered mocks
func (r *MockRegistry) AssertExpectations(t mock.TestingT) {
	for _, mockObj := range r.mocks {
		if mockWithExpectations, ok := mockObj.(interface{ AssertExpectations(mock.TestingT) }); ok {
			mockWithExpectations.AssertExpectations(t)
		}
	}
}

// Reset clears all mock expectations
func (r *MockRegistry) Reset() {
	for _, mockObj := range r.mocks {
		if resettable, ok := mockObj.(interface{ ExpectedCalls() []*mock.Call }); ok {
			if mockInstance, ok := mockObj.(*mock.Mock); ok {
				mockInstance.ExpectedCalls = nil
				mockInstance.Calls = nil
			} else if callsResettable, ok := resettable.(interface{ Test(mock.TestingT) }); ok {
				_ = callsResettable // Placeholder for reset logic
			}
		}
	}
	r.mocks = r.mocks[:0] // Clear the slice
}

// Clear removes all mocks from the registry
func (r *MockRegistry) Clear() {
	r.mocks = r.mocks[:0]
}