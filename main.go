package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joshvanl/cert-manager-csi/cmd"
)

func main() {
	c := cmd.RootCmd

	flag.CommandLine.Parse([]string{})

	c.Flags().AddGoFlagSet(flag.CommandLine)

	c.PersistentFlags().StringVar(&cmd.NodeID, "node-id", "", "node ID")
	c.MarkPersistentFlagRequired("node-id")

	c.PersistentFlags().StringVar(&cmd.Endpoint, "endpoint", "", "CSI endpoint")
	c.MarkPersistentFlagRequired("endpoint")

	c.PersistentFlags().StringVar(&cmd.DriverName, "driver-name", "cert-manager-csi", "name of the drive")

	c.PersistentFlags().StringVar(&cmd.DataRoot, "data-root", "cert-manager-csi", "directory to store ephemeral data")

	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
