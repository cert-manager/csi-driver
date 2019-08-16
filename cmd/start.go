package cmd

import (
	"flag"

	"github.com/spf13/cobra"

	"github.com/joshvanl/cert-manager-csi/pkg/driver"
)

var (
	Endpoint   string
	NodeID     string
	DriverName string
	DataRoot   string
)

func init() {
	flag.Set("logtostderr", "true")
}

var RootCmd = &cobra.Command{
	Use:   "cert-manager-csi",
	Short: "Container Storage Interface driver to issue certificates from Cert-Manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		d := driver.New(DriverName, NodeID, Endpoint, DataRoot)
		d.Run()
		return nil
	},
}
