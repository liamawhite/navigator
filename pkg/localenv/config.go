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

package localenv

// ServiceSpec defines how to create a service in the local environment
type ServiceSpec struct {
	// Name is the name of the service
	Name string

	// Replicas is the number of replicas for the service (ignored for external services)
	Replicas int32

	// Type determines the type of service to create
	Type ServiceType

	// ExternalIPs is used for external services to define manual endpoints
	ExternalIPs []string

	// NextService is used for microservice topology chaining - this service will call the next one
	NextService string
}

// ServiceType represents different types of services that can be deployed
type ServiceType string

const (
	// ServiceTypeWeb creates a regular ClusterIP service with deployment
	ServiceTypeWeb ServiceType = "web"

	// ServiceTypeHeadless creates a headless service (ClusterIP: None)
	ServiceTypeHeadless ServiceType = "headless"

	// ServiceTypeExternal creates a service with manual endpoints (no deployment)
	ServiceTypeExternal ServiceType = "external"

	// ServiceTypeTopology creates a service that can proxy to other services in a chain
	ServiceTypeTopology ServiceType = "topology"
)

// Scenario defines a collection of services that represent a specific demo scenario
type Scenario struct {
	// Name is the human-readable name of the scenario
	Name string

	// Description provides details about what this scenario demonstrates
	Description string

	// Services is the list of services to deploy for this scenario
	Services []ServiceSpec

	// IstioEnabled determines if this scenario should enable Istio injection
	IstioEnabled bool

	// Namespaces is a list of additional namespaces to create (beyond the default)
	Namespaces []string
}
