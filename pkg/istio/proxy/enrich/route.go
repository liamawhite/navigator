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

package enrich

import (
	"strings"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// enrichRouteType classifies route type based on Istio-specific patterns
func enrichRouteType() func(*v1alpha1.RouteConfigSummary) error {
	return func(route *v1alpha1.RouteConfigSummary) error {
		if route == nil {
			return nil
		}

		route.Type = inferIstioRouteType(route.Name, route.Type)
		return nil
	}
}

// inferIstioRouteType applies Istio-specific route type detection
func inferIstioRouteType(routeName string, currentType v1alpha1.RouteType) v1alpha1.RouteType {
	// If already classified as static, keep it
	if currentType == v1alpha1.RouteType_STATIC {
		return currentType
	}

	// Check for Istio-specific route patterns
	if strings.Contains(routeName, "|") {
		// Pipe-separated routes are typically service-specific in Istio
		return v1alpha1.RouteType_SERVICE_SPECIFIC
	}

	// Check for port-only routes (e.g., "80", "443", "15010")
	if isPortOnlyRoute(routeName) {
		return v1alpha1.RouteType_PORT_BASED
	}

	// Check for Istio static route patterns
	if isIstioStaticRoute(routeName) {
		return v1alpha1.RouteType_STATIC
	}

	return v1alpha1.RouteType_SERVICE_SPECIFIC
}

// isPortOnlyRoute checks if the route name is just a port number
func isPortOnlyRoute(routeName string) bool {
	if len(routeName) == 0 || len(routeName) > 5 {
		return false
	}

	// Simple check: all digits and valid port range
	for _, c := range routeName {
		if c < '0' || c > '9' {
			return false
		}
	}

	// Additional check for valid port range (1-65535)
	if routeName == "0" || (len(routeName) == 5 && routeName > "65535") {
		return false
	}

	return true
}

// isIstioStaticRoute checks if the route name matches Istio static patterns
func isIstioStaticRoute(routeName string) bool {
	// Empty names are static
	if strings.TrimSpace(routeName) == "" {
		return true
	}

	// Common Istio static route patterns
	staticPatterns := []string{
		"InboundPassthroughCluster",
		"BlackHoleCluster",
		"PassthroughCluster",
		"local_agent",
		"admin",
	}

	for _, pattern := range staticPatterns {
		if routeName == pattern {
			return true
		}
	}

	// Routes with localhost/127.0.0.1 are typically static
	if strings.Contains(routeName, "127.0.0.1") || strings.Contains(routeName, "localhost") {
		return true
	}

	return false
}
