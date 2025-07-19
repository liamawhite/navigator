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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/liamawhite/navigator/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Show version information including build details and Go version",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		info := version.GetInfo()

		switch output {
		case "json":
			jsonOutput, err := info.JSON()
			if err != nil {
				return fmt.Errorf("failed to marshal version info to JSON: %w", err)
			}
			fmt.Println(jsonOutput)
		case "text":
			fmt.Println(info.String())
		default:
			return fmt.Errorf("unsupported output format: %s (supported: text, json)", output)
		}

		return nil
	},
}

func init() {
	versionCmd.Flags().StringP("output", "o", "text", "Output format (text, json)")
}
