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
)

const (
	istioHelmRepoURL = "https://istio-release.storage.googleapis.com/charts"
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

	fmt.Printf("Downloading Istio Helm charts for version %s...\n", version)

	if err := downloadIstioCharts(version); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to download charts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded Istio charts to pkg/localenv/istio/charts/\n")
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
