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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	helmChartRegistry = "ghcr.io/liamawhite/microservice"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "This tool downloads the latest microservice Helm chart from the registry.\n")
		os.Exit(1)
	}

	fmt.Printf("Downloading latest microservice Helm chart...\n")

	if err := downloadLatestMicroserviceChart(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to download chart: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded microservice chart to pkg/localenv/microservice/charts/\n")
}

func downloadLatestMicroserviceChart() error {
	// Get current working directory and go up to charts directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create output directory in ../charts/ (single location)
	outputDir := filepath.Join(wd, "..", "charts")
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use a specific known tag (in real usage, this would be automated to find the latest)
	chartRef := fmt.Sprintf("%s:0.0.0-5acac59", helmChartRegistry)

	fmt.Printf("Downloading latest Helm chart from OCI registry: %s\n", chartRef)

	// Download chart from OCI registry
	chartTarPath := filepath.Join(outputDir, "microservice.tgz")
	if err := downloadFromOCI(chartRef, chartTarPath); err != nil {
		return fmt.Errorf("failed to download chart from OCI registry: %w", err)
	}

	fmt.Printf("Created chart tarball: %s\n", chartTarPath)
	return nil
}

func downloadFromOCI(chartRef, outputPath string) error {
	// Check if helm CLI is available
	if _, err := exec.LookPath("helm"); err != nil {
		return fmt.Errorf("helm CLI not found in PATH: %w", err)
	}

	// Create temporary directory for helm pull
	tempDir, err := os.MkdirTemp("", "microservice-chart-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp directory: %v\n", removeErr)
		}
	}()

	fmt.Printf("Using helm CLI to pull chart from OCI registry: %s\n", chartRef)

	// Use helm pull to download the chart from OCI registry
	// #nosec G204 -- chartRef is constructed from validated inputs
	cmd := exec.Command("helm", "pull",
		"oci://"+strings.TrimPrefix(chartRef, "oci://"),
		"--destination", tempDir,
		"--untar=false") // Keep as tarball

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull chart with helm: %w", err)
	}

	// Find the downloaded tarball in temp directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	var chartTarball string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tgz") {
			chartTarball = filepath.Join(tempDir, entry.Name())
			break
		}
	}

	if chartTarball == "" {
		return fmt.Errorf("no chart tarball found in temp directory")
	}

	// Move the downloaded tarball to the target location
	if err := os.Rename(chartTarball, outputPath); err != nil {
		return fmt.Errorf("failed to move chart tarball: %w", err)
	}

	fmt.Printf("Chart downloaded successfully: %s\n", outputPath)
	return nil
}
