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

// enrichBootstrapProxyMode infers and sets the proxy mode from node ID
func enrichBootstrapProxyMode() func(*v1alpha1.BootstrapSummary) error {
	return func(bootstrap *v1alpha1.BootstrapSummary) error {
		if bootstrap == nil || bootstrap.Node == nil {
			return nil
		}

		bootstrap.Node.ProxyMode = inferProxyMode(bootstrap.Node.Id)
		return nil
	}
}

// inferProxyMode infers the proxy mode from node ID
func inferProxyMode(nodeID string) v1alpha1.ProxyMode {
	if nodeID == "" {
		return v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE
	}

	// Istio node ID patterns:
	// - sidecar~<IP>~<pod>.<namespace>~<cluster>.svc.cluster.local
	// - router~<IP>~<gateway>.<namespace>~<cluster>.svc.cluster.local

	nodeID = strings.ToLower(nodeID)

	if strings.HasPrefix(nodeID, "sidecar~") {
		return v1alpha1.ProxyMode_SIDECAR
	}

	if strings.HasPrefix(nodeID, "router~") || strings.HasPrefix(nodeID, "gateway~") {
		return v1alpha1.ProxyMode_ROUTER
	}

	// Fallback for unknown patterns
	return v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE
}
