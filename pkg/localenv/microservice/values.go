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

package microservice

// CreateThreeTierApplicationValues creates custom values for the three-tier application topology
func CreateThreeTierApplicationValues() map[string]interface{} {
	return map[string]interface{}{
		// Base chart configuration
		"base": map[string]interface{}{
			// Namespace configuration
			"namespace": map[string]interface{}{
				"create": false,
				"name":   "default",
				"labels": map[string]interface{}{
					"istio-injection": "enabled",
				},
			},

			// Default values for all services
			"defaults": map[string]interface{}{
				"replicaCount": 1,
				"image": map[string]interface{}{
					"repository": "ghcr.io/liamawhite/microservice",
					"pullPolicy": "IfNotPresent",
					"tag":        "latest",
				},
				"serviceAccount": map[string]interface{}{
					"create":    true,
					"automount": true,
				},
				"service": map[string]interface{}{
					"type": "ClusterIP",
				},
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"memory": "128Mi",
						"cpu":    "100m",
					},
					"limits": map[string]interface{}{
						"memory": "256Mi",
						"cpu":    "200m",
					},
				},
				"livenessProbe": map[string]interface{}{
					"httpGet": map[string]interface{}{
						"path": "/health",
						"port": "http",
					},
					"initialDelaySeconds": 30,
					"periodSeconds":       10,
				},
				"readinessProbe": map[string]interface{}{
					"httpGet": map[string]interface{}{
						"path": "/health",
						"port": "http",
					},
					"initialDelaySeconds": 5,
					"periodSeconds":       5,
				},
				"config": map[string]interface{}{
					"timeout":   "30s",
					"logLevel":  "info",
					"logFormat": "json",
				},
			},

			// Three-tier services configuration
			"services": []interface{}{
				// Frontend - Entry point service (web tier)
				map[string]interface{}{
					"name": "frontend",
					"config": map[string]interface{}{
						"serviceName": "frontend",
						"port":        8080,
					},
					"service": map[string]interface{}{
						"port": 8080,
					},
					"autoscaling": map[string]interface{}{
						"enabled": false,
					},
				},
				// Backend - Application logic service (app tier)
				map[string]interface{}{
					"name": "backend",
					"config": map[string]interface{}{
						"serviceName": "backend",
						"port":        8080,
					},
					"service": map[string]interface{}{
						"port": 8080,
					},
					"autoscaling": map[string]interface{}{
						"enabled": false,
					},
				},
				// Database - Data storage service (data tier)
				map[string]interface{}{
					"name": "database",
					"config": map[string]interface{}{
						"serviceName": "database",
						"port":        8080,
					},
					"service": map[string]interface{}{
						"port": 8080,
					},
					"autoscaling": map[string]interface{}{
						"enabled": false,
					},
				},
			},
		},
	}
}
