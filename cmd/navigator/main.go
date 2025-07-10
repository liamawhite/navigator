package main

import (
	"os"

	"github.com/liamawhite/navigator/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
