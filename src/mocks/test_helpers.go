package mocks

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

// TestFixture manages test data and setup/teardown operations
type TestFixture struct {
	Name     string
	Setup    func() error
	Teardown func() error
	Data     map[string]interface{}
}

// NewTestFixture creates a new test fixture
func NewTestFixture(name string) *TestFixture {
	return &TestFixture{
		Name: name,
		Data: make(map[string]interface{}),
	}
}

// Run executes the fixture setup
func (f *TestFixture) Run() error {
	if f.Setup != nil {
		return f.Setup()
	}
	return nil
}

// Cleanup executes the fixture teardown
func (f *TestFixture) Cleanup() error {
	if f.Teardown != nil {
		return f.Teardown()
	}
	return nil
}

// AssertHelper provides custom assertion utilities
type AssertHelper struct {
	t *testing.T
}

// NewAssertHelper creates a new assertion helper
func NewAssertHelper(t *testing.T) *AssertHelper {
	return &AssertHelper{t: t}
}

// AssertErrorContains checks if error contains substring
func (a *AssertHelper) AssertErrorContains(err error, substring string) {
	if err == nil {
		a.t.Errorf("Expected error containing '%s', but got nil", substring)
		return
	}
	assert.Contains(a.t, err.Error(), substring)
}

// AssertPanics checks if function panics
func (a *AssertHelper) AssertPanics(fn func()) {
	defer func() {
		if r := recover(); r == nil {
			a.t.Error("Expected function to panic, but it didn't")
		}
	}()
	fn()
}

// AssertEventually checks if condition becomes true within timeout
func (a *AssertHelper) AssertEventually(condition func() bool, timeout time.Duration, interval time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}
	a.t.Error("Condition was not met within timeout")
}

// AssertJSONEqual compares JSON strings for equality
func (a *AssertHelper) AssertJSONEqual(expected, actual string) {
	assert.JSONEq(a.t, expected, actual)
}