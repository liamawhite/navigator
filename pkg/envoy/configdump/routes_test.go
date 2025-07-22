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

	"github.com/stretchr/testify/assert"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestDetermineRouteType(t *testing.T) {
	tests := []struct {
		name               string
		routeName          string
		isFromStaticConfig bool
		expectedType       v1alpha1.RouteType
	}{
		// Generic Envoy behavior - most non-empty routes are service-specific
		{
			name:               "port 80",
			routeName:          "80",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		{
			name:               "port 443",
			routeName:          "443",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		{
			name:               "service route",
			routeName:          "backend.demo.svc.cluster.local:8080",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		{
			name:               "complex route pattern",
			routeName:          "outbound|8080||backend.demo.svc.cluster.local",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		// Static routes from config are always static
		{
			name:               "static config route",
			routeName:          "static-route-from-config",
			isFromStaticConfig: true,
			expectedType:       v1alpha1.RouteType_STATIC,
		},
		{
			name:               "static config with complex name",
			routeName:          "outbound|8080||backend.demo.svc.cluster.local",
			isFromStaticConfig: true,
			expectedType:       v1alpha1.RouteType_STATIC,
		},
		{
			name:               "static config empty name",
			routeName:          "",
			isFromStaticConfig: true,
			expectedType:       v1alpha1.RouteType_STATIC,
		},
		// Empty route names are static (basic Envoy behavior)
		{
			name:               "empty route name",
			routeName:          "",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_STATIC,
		},
		// Default behavior for generic Envoy deployments
		{
			name:               "unknown pattern",
			routeName:          "some-unknown-pattern",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineRouteType(tt.routeName, tt.isFromStaticConfig)
			assert.Equal(t, tt.expectedType, result, "Route %q (static=%v) should be categorized as %v, got %v", tt.routeName, tt.isFromStaticConfig, tt.expectedType, result)
		})
	}
}
