package options

import (
	"github.com/spf13/cobra"
)

type Options struct {
	NodeID     string
	DriverName string

	// CSI driver endpoint.
	Endpoint string

	// Root directory to write data and mount from.
	DataRoot string

	// Endpoint that Kubelet should connect to driver.
	KubeletRegistrationEndpoint string
}

func AddFlags(cmd *cobra.Command) *Options {
	var opts Options

	cmd.PersistentFlags().StringVar(&opts.NodeID, "node-id", "", "node ID")
	cmd.MarkPersistentFlagRequired("node-id")

	cmd.PersistentFlags().StringVar(&opts.Endpoint, "endpoint", "", "CSI endpoint")
	cmd.MarkPersistentFlagRequired("endpoint")

	cmd.PersistentFlags().StringVar(&opts.KubeletRegistrationEndpoint,
		"kubelet-registration-path", "", "Path of the CSI driver socket on the Kubernetes host machine.")
	cmd.MarkPersistentFlagRequired("kubelet-registration-path")

	cmd.PersistentFlags().StringVar(&opts.DriverName, "driver-name",
		"cert-manager-csi", "name of the drive")

	cmd.PersistentFlags().StringVar(&opts.DataRoot, "data-root",
		"/csi-data-dir", "directory to store ephemeral data")

	return &opts
}
