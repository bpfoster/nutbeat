package main

import (
	"os"

	"github.com/bpfoster/nutbeat/cmd"

	_ "github.com/bpfoster/nutbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
