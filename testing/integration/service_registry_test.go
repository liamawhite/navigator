package integration

import (
	"testing"
)

// TestServiceRegistry runs all integration tests using table-driven approach
func TestServiceRegistry(t *testing.T, env TestEnvironment) {
	testCases := GetTestCases()
	for _, testCase := range testCases {
		RunTestCase(t, env, testCase)
	}
}
