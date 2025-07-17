// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
