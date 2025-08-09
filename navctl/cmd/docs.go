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

package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate CLI reference documentation",
	Hidden: true,
	Long: `Generate markdown documentation for all navctl commands.

This command generates comprehensive CLI reference documentation
in markdown format from the cobra command definitions.`,
	RunE: runDocs,
}

func init() {
	docsCmd.Flags().StringP("output", "o", "docs/reference/cli", "Output directory for generated documentation")
	rootCmd.AddCommand(docsCmd)
}

func runDocs(cmd *cobra.Command, args []string) error {
	outputDir, _ := cmd.Flags().GetString("output")

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return err
	}

	// Disable auto-generated tag for cleaner output
	rootCmd.DisableAutoGenTag = true

	// Generate markdown docs for all commands
	err := doc.GenMarkdownTree(rootCmd, outputDir)
	if err != nil {
		return err
	}

	// Normalize environment-specific paths in generated docs
	if err := normalizeGeneratedDocs(outputDir); err != nil {
		return err
	}

	// Rename the main navctl.md to cli-reference.md to replace the manual version
	oldPath := outputDir + "/navctl.md"
	newPath := outputDir + "/cli-reference.md"
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	cmd.Printf("CLI documentation generated successfully in %s\n", outputDir)
	cmd.Printf("Main CLI reference available at %s\n", newPath)
	return nil
}

// normalizeGeneratedDocs normalizes environment-specific paths in generated documentation
func normalizeGeneratedDocs(outputDir string) error {
	// Regex patterns to normalize environment-specific kubeconfig paths
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`/Users/[^/\s]+/\.kube/config`),
		regexp.MustCompile(`/home/[^/\s]+/\.kube/config`),
	}

	return filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .md files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Ensure path is within output directory to prevent path traversal
		relPath, err := filepath.Rel(outputDir, path)
		if err != nil || strings.Contains(relPath, "..") {
			return nil // Skip files outside output directory
		}

		// Read file content
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}

		// Apply normalization patterns
		normalized := string(content)
		for _, pattern := range patterns {
			normalized = pattern.ReplaceAllString(normalized, "~/.kube/config")
		}

		// Write back if content changed
		if normalized != string(content) {
			if err := os.WriteFile(path, []byte(normalized), info.Mode()); err != nil {
				return err
			}
		}

		return nil
	})
}
