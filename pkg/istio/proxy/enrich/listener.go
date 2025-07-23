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
	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// enrichListenerType classifies listener type based on Istio-specific patterns
func enrichListenerType(proxyMode v1alpha1.ProxyMode) func(*v1alpha1.ListenerSummary) error {
	return func(listener *v1alpha1.ListenerSummary) error {
		if listener == nil {
			return nil
		}

		listener.Type = inferIstioListenerType(listener.Name, listener.Address, listener.Port, listener.UseOriginalDst, proxyMode)
		return nil
	}
}

// inferIstioListenerType applies Istio-specific listener type detection
func inferIstioListenerType(name, address string, port uint32, useOriginalDst bool, proxyMode v1alpha1.ProxyMode) v1alpha1.ListenerType {
	// Check for Istio virtual listeners by name
	if name == "virtualInbound" {
		return v1alpha1.ListenerType_VIRTUAL_INBOUND
	}
	if name == "virtualOutbound" {
		return v1alpha1.ListenerType_VIRTUAL_OUTBOUND
	}

	// Check for Istio-specific ports on 0.0.0.0
	if address == "0.0.0.0" {
		switch port {
		case 15090:
			// Prometheus metrics endpoint
			return v1alpha1.ListenerType_PROXY_METRICS
		case 15021:
			// Health check endpoint
			return v1alpha1.ListenerType_PROXY_HEALTHCHECK
		case 15001, 15006:
			// Istio virtual outbound listeners
			if useOriginalDst {
				return v1alpha1.ListenerType_VIRTUAL_OUTBOUND
			}
		}

		// Gateway-specific logic: 0.0.0.0 listeners without useOriginalDst
		if proxyMode == v1alpha1.ProxyMode_GATEWAY && !useOriginalDst {
			return v1alpha1.ListenerType_GATEWAY_INBOUND
		}

		// Other 0.0.0.0 listeners are port-based (for sidecars)
		return v1alpha1.ListenerType_PORT_OUTBOUND
	}

	// Specific IP addresses in Istio typically indicate service-specific outbound listeners
	return v1alpha1.ListenerType_SERVICE_OUTBOUND
}
