package cli

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
