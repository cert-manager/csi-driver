package options

import (
	"github.com/spf13/cobra"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

type Options struct {
	DriverID csiapi.DriverID

	// CSI driver endpoint.
	Endpoint string

	// Root directory to write data and mount from.
	DataRoot string

	// Size in Mbytes to create the tmpfs file system to write and mount from.
	TmpfsSize string

	// Optional Webhook configuration
	Webhook Webhook
}

// Optional Webhook configuration
type Webhook struct {
	// URL to server to consume Webhook Create, Renew, Destroy
	NetHost string
	//TODO: add New CA trust bundle
}

func AddFlags(cmd *cobra.Command) *Options {
	var opts Options

	cmd.PersistentFlags().StringVar(&opts.DriverID.NodeID, "node-id", "", "node ID")
	cmd.MarkPersistentFlagRequired("node-id")

	cmd.PersistentFlags().StringVar(&opts.DriverID.DriverName, "driver-name",
		"csi.cert-manager.io", "name of the driver")
	cmd.MarkPersistentFlagRequired("driver-name")

	cmd.PersistentFlags().StringVar(&opts.Endpoint, "endpoint", "", "CSI endpoint")
	cmd.MarkPersistentFlagRequired("endpoint")

	cmd.PersistentFlags().StringVar(&opts.DataRoot, "data-root",
		"/csi-data-dir", "directory to store ephemeral data")

	cmd.PersistentFlags().StringVar(&opts.TmpfsSize, "tmpfs-size",
		"100", "size in Mbytes to create the tmpfs file system to store ephemeral data")

	addWebhookFlags(cmd, &opts)

	return &opts
}

func addWebhookFlags(cmd *cobra.Command, opts *Options) {
	cmd.PersistentFlags().StringVar(&opts.Webhook.NetHost, "webhook-net-host",
		"", "optional URL to a server to consume Create,Renew,Destroy webhooks for certificates")
}
