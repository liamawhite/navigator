package integration

import (
	"testing"
)

// RunServiceRegistryTests runs all integration tests using table-driven approach
func RunServiceRegistryTests(t *testing.T, env TestEnvironment) {
	testCases := GetTestCases()
	for _, testCase := range testCases {
		RunTestCase(t, env, testCase)
	}
}
