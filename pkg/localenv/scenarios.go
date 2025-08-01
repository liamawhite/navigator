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

// Predefined scenarios for local development and demos

var (
	// ScenarioBasic provides a simple single web service for basic testing
	ScenarioBasic = &Scenario{
		Name:        "basic",
		Description: "A single web service for basic Navigator functionality testing",
		Services: []ServiceSpec{
			{
				Name:     "web-service",
				Replicas: 2,
				Type:     ServiceTypeWeb,
			},
		},
		IstioEnabled: true,
	}

	// ScenarioMicroserviceTopology creates a chain of services that call each other
	ScenarioMicroserviceTopology = &Scenario{
		Name:        "microservice-topology",
		Description: "A chain of three microservices demonstrating service-to-service communication",
		Services: []ServiceSpec{
			{
				Name:        "frontend",
				Replicas:    1,
				Type:        ServiceTypeTopology,
				NextService: "backend",
			},
			{
				Name:        "backend",
				Replicas:    2,
				Type:        ServiceTypeTopology,
				NextService: "database",
			},
			{
				Name:     "database",
				Replicas: 1,
				Type:     ServiceTypeWeb,
			},
		},
		IstioEnabled: true,
	}
)

// GetScenarioByName returns a scenario by its name
func GetScenarioByName(name string) *Scenario {
	scenarios := map[string]*Scenario{
		"basic":                 ScenarioBasic,
		"microservice-topology": ScenarioMicroserviceTopology,
	}

	return scenarios[name]
}

// ListScenarios returns all available scenarios
func ListScenarios() []*Scenario {
	return []*Scenario{
		ScenarioBasic,
		ScenarioMicroserviceTopology,
	}
}

// GetScenarioNames returns the names of all available scenarios
func GetScenarioNames() []string {
	scenarios := ListScenarios()
	names := make([]string, len(scenarios))
	for i, scenario := range scenarios {
		names[i] = scenario.Name
	}
	return names
}
