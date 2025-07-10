package local

import (
	"testing"

	"github.com/liamawhite/navigator/testing/integration"
)

// TestLocalServiceRegistry runs all integration tests using local Kind environment
func TestLocalServiceRegistry(t *testing.T) {
	env := NewLocalEnvironment("navigator-local-tests")

	// Run all test cases directly
	testCases := integration.GetTestCases()
	for _, testCase := range testCases {
		integration.RunTestCase(t, env, testCase)
	}
}
