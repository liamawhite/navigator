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

package configdump

import (
	"testing"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestParser_DetermineListenerType(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name           string
		listenerName   string
		address        string
		port           uint32
		useOriginalDst bool
		expected       v1alpha1.ListenerType
	}{
		// Generic 0.0.0.0 listeners
		{
			name:         "0.0.0.0 listener without original dst",
			listenerName: "listener-80",
			address:      "0.0.0.0",
			port:         80,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
		},
		{
			name:           "0.0.0.0 listener with original dst",
			listenerName:   "catch-all",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			expected:       v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
		},
		// Specific IP listeners
		{
			name:         "specific IP listener",
			listenerName: "service-listener",
			address:      "10.96.1.1",
			port:         8080,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
		},
		{
			name:         "another specific IP listener",
			listenerName: "another-service",
			address:      "192.168.1.10",
			port:         443,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.determineListenerType(tt.listenerName, tt.address, tt.port, tt.useOriginalDst)
			assert.Equal(t, tt.expected, result, "Listener type should match for %s", tt.name)
		})
	}
}
