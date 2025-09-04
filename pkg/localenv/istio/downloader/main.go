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

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	istioHelmRepoURL = "https://istio-release.storage.googleapis.com/charts"
	// PrometheusNodePort matches the constant in pkg/localenv/kind/kind.go
	// This ensures Prometheus is accessible on localhost:30090 in Kind clusters
	prometheusNodePort = 30090
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <version>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s 1.25.4\n", os.Args[0])
		os.Exit(1)
	}

	version := os.Args[1]
	if err := validateVersion(version); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid version: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloading Istio Helm charts and addons for version %s...\n", version)

	if err := downloadIstioCharts(version); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to download charts: %v\n", err)
		os.Exit(1)
	}

	if err := downloadIstioAddons(version); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to download addons: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded Istio charts and addons to pkg/localenv/istio/charts/\n")
}

func validateVersion(version string) error {
	// Basic semantic version validation (e.g., 1.25.4, 1.20.0)
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+$`, version)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("version must be in format x.y.z (e.g., 1.25.4)")
	}
	return nil
}

func downloadIstioCharts(version string) error {
	// Get current working directory and go up to charts directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create output directory in ../charts/{version}/
	outputDir := filepath.Join(wd, "..", "charts", version)
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Download the main Istio charts
	charts := []string{
		"base",
		"istiod",
		"gateway",
	}

	for _, chart := range charts {
		if err := downloadChart(chart, version, outputDir); err != nil {
			return fmt.Errorf("failed to download %s chart: %w", chart, err)
		}
		fmt.Printf("Downloaded %s chart\n", chart)
	}

	return nil
}

func downloadIstioAddons(version string) error {
	// Get current working directory and go up to charts directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create output directory in ../charts/{version}/
	outputDir := filepath.Join(wd, "..", "charts", version)
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Download Prometheus addon
	if err := downloadPrometheusAddon(version, outputDir); err != nil {
		return fmt.Errorf("failed to download Prometheus addon: %w", err)
	}
	fmt.Printf("Downloaded Prometheus addon\n")

	return nil
}

func downloadPrometheusAddon(version, outputDir string) error {
	// Construct download URL for Prometheus addon
	prometheusURL := fmt.Sprintf("https://raw.githubusercontent.com/istio/istio/%s/samples/addons/prometheus.yaml", version)

	fmt.Printf("Downloading Prometheus addon from %s\n", prometheusURL)

	// Download the addon
	// #nosec G107 -- prometheusURL is constructed from validated inputs
	resp, err := http.Get(prometheusURL)
	if err != nil {
		return fmt.Errorf("failed to download Prometheus addon: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Prometheus addon: HTTP %d", resp.StatusCode)
	}

	// Create the YAML file path
	yamlFilePath := filepath.Join(outputDir, "prometheus.yaml")

	// Create the YAML file
	// #nosec G304 -- yamlFilePath is constructed from validated inputs
	yamlFile, err := os.Create(yamlFilePath)
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %w", err)
	}
	defer func() {
		if closeErr := yamlFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close YAML file: %v\n", closeErr)
		}
	}()

	// Read the downloaded content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Patch the Prometheus service to use NodePort for local Kind cluster access
	patchedContent := patchPrometheusServiceForNodePort(string(content))

	// Write the patched content to the YAML file
	if _, err := yamlFile.WriteString(patchedContent); err != nil {
		return fmt.Errorf("failed to write patched YAML file: %w", err)
	}

	return nil
}

func downloadChart(chartName, version, outputDir string) error {
	// Construct download URL for the chart
	chartURL := fmt.Sprintf("%s/%s-%s.tgz", istioHelmRepoURL, chartName, version)

	fmt.Printf("Downloading %s from %s\n", chartName, chartURL)

	// Download the chart
	// #nosec G107 -- chartURL is constructed from validated inputs
	resp, err := http.Get(chartURL)
	if err != nil {
		return fmt.Errorf("failed to download chart: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download chart: HTTP %d", resp.StatusCode)
	}

	// Create the tar file path
	tarFileName := fmt.Sprintf("%s-%s.tgz", chartName, version)
	tarFilePath := filepath.Join(outputDir, tarFileName)

	// Create the tar file
	// #nosec G304 -- tarFilePath is constructed from validated inputs
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer func() {
		if closeErr := tarFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close tar file: %v\n", closeErr)
		}
	}()

	// Copy the downloaded content to the tar file
	if _, err := io.Copy(tarFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save tar file: %w", err)
	}

	return nil
}

// patchPrometheusServiceForNodePort modifies the Prometheus service configuration
// to use NodePort instead of ClusterIP for direct access in Kind clusters
func patchPrometheusServiceForNodePort(yamlContent string) string {
	lines := strings.Split(yamlContent, "\n")
	var result []string
	var inPrometheusService bool
	var inServiceSpec bool
	var inServicePorts bool
	var serviceModified bool

	for i, line := range lines {
		// Detect the start of the Prometheus service
		if strings.Contains(line, "# Source: prometheus/templates/service.yaml") {
			inPrometheusService = true
			result = append(result, line)
			continue
		}

		// Detect the end of the Prometheus service (next resource starts)
		if inPrometheusService && strings.HasPrefix(line, "---") && i > 0 {
			inPrometheusService = false
			inServiceSpec = false
			inServicePorts = false
		}

		// If we're in the Prometheus service, look for specific fields to modify
		if inPrometheusService {
			// Detect service spec section
			if strings.HasPrefix(line, "spec:") {
				inServiceSpec = true
				result = append(result, line)
				continue
			}

			// Detect ports section within service spec
			if inServiceSpec && strings.HasPrefix(line, "  ports:") {
				inServicePorts = true
				result = append(result, line)
				continue
			}

			// Modify the service type from ClusterIP to NodePort
			if inServiceSpec && strings.Contains(line, `type: "ClusterIP"`) {
				result = append(result, "  # Modified by Navigator: Changed from ClusterIP to NodePort for local Kind cluster access")
				result = append(result, `  type: "NodePort"`)
				serviceModified = true
				continue
			}

			// Add nodePort to the http port configuration
			if inServicePorts && strings.Contains(line, "targetPort: 9090") {
				result = append(result, line)
				// Add the nodePort on the next line with proper indentation
				result = append(result, fmt.Sprintf("      # Added by Navigator: Fixed NodePort for consistent local access via localhost:%d", prometheusNodePort))
				result = append(result, fmt.Sprintf("      nodePort: %d", prometheusNodePort))
				continue
			}

			// End of service spec - reset flags
			if inServiceSpec && strings.HasPrefix(line, "---") {
				inServiceSpec = false
				inServicePorts = false
			}
		}

		// Add the line as-is
		result = append(result, line)
	}

	modifiedContent := strings.Join(result, "\n")

	// If we successfully modified the service, add a header comment
	if serviceModified {
		headerComment := `# This Prometheus manifest has been modified by Navigator for local Kind cluster usage:
# - Service type changed from ClusterIP to NodePort
# - Fixed nodePort 30090 added for consistent localhost access
# - Original source: https://raw.githubusercontent.com/istio/istio/VERSION/samples/addons/prometheus.yaml
#
`
		modifiedContent = headerComment + modifiedContent
	}

	return modifiedContent
}
