package main

import (
	"fmt"
	"os"

	"github.com/joshvanl/cert-manager-csi/cmd/app"
)

func main() {
	if err := app.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
