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

// Package enrich provides Istio-specific enrichment capabilities.
// This package enriches generic Envoy data structures with Istio service mesh
// specific information and interpretations.
package enrich

import (
	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/envoy/configdump"
)

// ProxyConfigSummary enriches a proxy config summary with Istio-specific information
func ProxyConfigSummary(summary *configdump.ParsedSummary) error {
	// Enrich bootstrap
	if summary.Bootstrap != nil {
		if err := enrichBootstrapProxyMode()(summary.Bootstrap); err != nil {
			return err
		}
	}

	// Enrich listeners
	for _, listener := range summary.Listeners {
		if err := enrichListenerType()(listener); err != nil {
			return err
		}
	}

	// Enrich clusters
	for _, cluster := range summary.Clusters {
		if err := enrichClusterNameComponents()(cluster); err != nil {
			return err
		}
	}

	// Enrich routes
	for _, route := range summary.Routes {
		if err := enrichRouteType()(route); err != nil {
			return err
		}
	}

	return nil
}

// EndpointSummaries enriches endpoint summaries with Istio-specific information
func EndpointSummaries(endpoints []*v1alpha1.EndpointSummary) error {
	for _, endpoint := range endpoints {
		if err := enrichEndpointClusterName()(endpoint); err != nil {
			return err
		}
		if err := enrichEndpointClusterType()(endpoint); err != nil {
			return err
		}
	}
	return nil
}
