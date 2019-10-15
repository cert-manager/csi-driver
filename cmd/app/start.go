package app

import (
	"flag"

	"github.com/spf13/cobra"

	"github.com/jetstack/cert-manager-csi/cmd/app/options"
	"github.com/jetstack/cert-manager-csi/pkg/driver"
)

var (
	opts *options.Options
)

func init() {
	flag.Set("logtostderr", "true")

	flag.CommandLine.Parse([]string{})

	RootCmd.Flags().AddGoFlagSet(flag.CommandLine)

	opts = options.AddFlags(RootCmd)
}

var RootCmd = &cobra.Command{
	Use:   "cert-manager-csi",
	Short: "Container Storage Interface driver to issue certificates from Cert-Manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := driver.New(opts.DriverName, opts.NodeID, opts.Endpoint, opts.DataRoot)
		if err != nil {
			return err
		}

		d.Run()
		return nil
	},
}
