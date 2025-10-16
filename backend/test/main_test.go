package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	// Setup code can go here
	m.Run()
	// Teardown code can go here
}

// TestBasicAssertion is a simple sanity check
func TestBasicAssertion(t *testing.T) {
	assert.True(t, true, "This should always pass")
	assert.Equal(t, 1+1, 2, "Math should work")
}

// TestStringOperations tests basic string operations
func TestStringOperations(t *testing.T) {
	str := "Elang Backend"
	assert.Contains(t, str, "Backend")
	assert.NotEmpty(t, str)
	assert.Len(t, str, 13)
}

// TestSliceOperations tests basic slice operations
func TestSliceOperations(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	assert.Len(t, slice, 5)
	assert.Contains(t, slice, 3)
	assert.NotContains(t, slice, 10)
}

// TestMapOperations tests basic map operations
func TestMapOperations(t *testing.T) {
	m := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	assert.Len(t, m, 3)
	assert.Equal(t, 2, m["two"])
	_, exists := m["four"]
	assert.False(t, exists)
}
