//go:build !solution

package main

import (
	"os"

	"gitlab.com/slon/shad-go/gitfame/internal/cli"
)

func main() {
	if err := cli.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
