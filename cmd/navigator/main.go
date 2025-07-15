package main

import (
	"os"

	"github.com/liamawhite/navigator/internal/cli"
	_ "github.com/liamawhite/navigator/ui" // Import to trigger embed
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
