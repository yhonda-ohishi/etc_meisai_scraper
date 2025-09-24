package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of LoggerInterface
type MockLogger struct {
	mock.Mock
}

// Printf formats according to a format specifier and writes to the logger
func (m *MockLogger) Printf(format string, v ...interface{}) {
	args := make([]interface{}, 0, len(v)+1)
	args = append(args, format)
	args = append(args, v...)
	m.Called(args...)
}

// Println writes to the logger with a newline
func (m *MockLogger) Println(v ...interface{}) {
	m.Called(v...)
}

// Print writes to the logger
func (m *MockLogger) Print(v ...interface{}) {
	m.Called(v...)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1)
func (m *MockLogger) Fatalf(format string, v ...interface{}) {
	args := make([]interface{}, 0, len(v)+1)
	args = append(args, format)
	args = append(args, v...)
	m.Called(args...)
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1)
func (m *MockLogger) Fatal(v ...interface{}) {
	m.Called(v...)
}

// Panicf is equivalent to Printf() followed by a call to panic()
func (m *MockLogger) Panicf(format string, v ...interface{}) {
	args := make([]interface{}, 0, len(v)+1)
	args = append(args, format)
	args = append(args, v...)
	m.Called(args...)
}

// Panic is equivalent to Print() followed by a call to panic()
func (m *MockLogger) Panic(v ...interface{}) {
	m.Called(v...)
}

// Ensure MockLogger implements the interface
var _ interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Print(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatal(v ...interface{})
	Panicf(format string, v ...interface{})
	Panic(v ...interface{})
} = (*MockLogger)(nil)