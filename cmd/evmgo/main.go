package main

import (
	"os"

	"github.com/itzfelixv/evmgo/internal/cli"
)

func main() {
	if err := cli.Execute(os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
