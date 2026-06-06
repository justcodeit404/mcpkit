package main

import (
	"os"

	"github.com/justcodeit404/mcpkit/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
