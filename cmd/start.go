package cmd

import (
	"flag"

	"github.com/spf13/cobra"

	"github.com/joshvanl/cert-manager-csi/pkg/driver"
	"github.com/joshvanl/cert-manager-csi/pkg/registrar"
)

var (
	Endpoint        string
	KubeletEndpoint string
	NodeID          string
	DriverName      string
	DataRoot        string
)

func init() {
	flag.Set("logtostderr", "true")
}

var RootCmd = &cobra.Command{
	Use:   "cert-manager-csi",
	Short: "Container Storage Interface driver to issue certificates from Cert-Manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := driver.New(DriverName, NodeID, Endpoint, DataRoot)
		if err != nil {
			return err
		}

		r := registrar.New(DriverName, KubeletEndpoint, d.NodeServer())
		if err := r.Run(); err != nil {
			return err
		}

		d.Run()
		return nil
	},
}
