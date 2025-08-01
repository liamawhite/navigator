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

package demo

import (
	"fmt"
	"strings"

	"github.com/liamawhite/navigator/pkg/localenv"
)

// ScenarioInfo provides display information for demo scenarios
type ScenarioInfo struct {
	Name         string
	Description  string
	ServiceCount int
	IstioEnabled bool
	Services     []string
}

// GetAvailableScenarios returns information about all available demo scenarios
func GetAvailableScenarios() []ScenarioInfo {
	scenarios := localenv.ListScenarios()
	var info []ScenarioInfo

	for _, scenario := range scenarios {
		var serviceNames []string
		for _, service := range scenario.Services {
			serviceNames = append(serviceNames, service.Name)
		}

		info = append(info, ScenarioInfo{
			Name:         scenario.Name,
			Description:  scenario.Description,
			ServiceCount: len(scenario.Services),
			IstioEnabled: scenario.IstioEnabled,
			Services:     serviceNames,
		})
	}

	return info
}

// GetScenarioInfo returns information about a specific scenario
func GetScenarioInfo(scenarioName string) (*ScenarioInfo, error) {
	scenario := localenv.GetScenarioByName(scenarioName)
	if scenario == nil {
		return nil, fmt.Errorf("scenario '%s' not found", scenarioName)
	}

	var serviceNames []string
	for _, service := range scenario.Services {
		serviceNames = append(serviceNames, service.Name)
	}

	return &ScenarioInfo{
		Name:         scenario.Name,
		Description:  scenario.Description,
		ServiceCount: len(scenario.Services),
		IstioEnabled: scenario.IstioEnabled,
		Services:     serviceNames,
	}, nil
}

// ValidateScenarioName validates that a scenario name exists
func ValidateScenarioName(scenarioName string) error {
	scenario := localenv.GetScenarioByName(scenarioName)
	if scenario == nil {
		availableNames := localenv.GetScenarioNames()
		return fmt.Errorf("scenario '%s' not found. Available scenarios: %s",
			scenarioName, strings.Join(availableNames, ", "))
	}
	return nil
}

// GetScenarioNames returns a list of all available scenario names
func GetScenarioNames() []string {
	return localenv.GetScenarioNames()
}

// FormatScenarioList returns a formatted string listing all scenarios
func FormatScenarioList() string {
	scenarios := GetAvailableScenarios()
	if len(scenarios) == 0 {
		return "No scenarios available"
	}

	var builder strings.Builder
	builder.WriteString("Available demo scenarios:\n\n")

	for _, info := range scenarios {
		builder.WriteString(fmt.Sprintf("  %s\n", info.Name))
		builder.WriteString(fmt.Sprintf("    Description: %s\n", info.Description))
		builder.WriteString(fmt.Sprintf("    Services: %d (%s)\n",
			info.ServiceCount, strings.Join(info.Services, ", ")))
		if info.IstioEnabled {
			builder.WriteString("    Istio: enabled\n")
		} else {
			builder.WriteString("    Istio: disabled\n")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
