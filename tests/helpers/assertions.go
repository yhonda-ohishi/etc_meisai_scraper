package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertNoError checks that error is nil with enhanced error message
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NoError(t, err, msgAndArgs...)
}

// AssertError checks that error is not nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Error(t, err, msgAndArgs...)
}

// AssertEqual checks equality with type-specific formatting
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotEqual checks inequality
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEqual(t, expected, actual, msgAndArgs...)
}

// AssertNil checks that value is nil
func AssertNil(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Nil(t, object, msgAndArgs...)
}

// AssertNotNil checks that value is not nil
func AssertNotNil(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotNil(t, object, msgAndArgs...)
}

// AssertTrue checks that value is true
func AssertTrue(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	assert.True(t, value, msgAndArgs...)
}

// AssertFalse checks that value is false
func AssertFalse(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	assert.False(t, value, msgAndArgs...)
}

// AssertContains checks that haystack contains needle
func AssertContains(t *testing.T, haystack, needle interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Contains(t, haystack, needle, msgAndArgs...)
}

// AssertLen checks that object has expected length
func AssertLen(t *testing.T, object interface{}, length int, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Len(t, object, length, msgAndArgs...)
}

// AssertNotEmpty checks that object is not empty
func AssertNotEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEmpty(t, object, msgAndArgs...)
}

// AssertNotContains checks that haystack does not contain needle
func AssertNotContains(t *testing.T, haystack, needle interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotContains(t, haystack, needle, msgAndArgs...)
}

// AssertEmpty checks that object is empty
func AssertEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Empty(t, object, msgAndArgs...)
}