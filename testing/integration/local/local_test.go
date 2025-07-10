package local

import (
	"testing"

	"github.com/liamawhite/navigator/testing/integration"
)

var sharedCluster *SharedCluster

// TestLocalServiceRegistry runs all integration tests using local Kind environment
func TestLocalServiceRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup shared cluster once for all tests
	if sharedCluster == nil {
		var err error
		sharedCluster, err = NewSharedCluster("navigator-integration-tests")
		if err != nil {
			t.Fatalf("Failed to create shared cluster: %v", err)
		}
		t.Cleanup(func() {
			if sharedCluster != nil {
				sharedCluster.Cleanup()
			}
		})
	}

	// Run all test cases with shared cluster
	testCases := integration.GetTestCases()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			env := sharedCluster.NewEnvironment()
			integration.RunTestCase(t, env, testCase)
		})
	}
}
