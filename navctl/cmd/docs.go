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

	// Generate markdown docs for all commands
	err := doc.GenMarkdownTree(rootCmd, outputDir)
	if err != nil {
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
