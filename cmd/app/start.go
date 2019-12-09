package app

import (
	"flag"

	"github.com/spf13/cobra"

	"github.com/jetstack/cert-manager-csi/cmd/app/options"
	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager-csi/pkg/driver"
	"github.com/jetstack/cert-manager-csi/pkg/webhook"
	"github.com/jetstack/cert-manager-csi/pkg/webhook/net"
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
		var whClients []csiapi.WebhookClient
		if len(opts.Webhook.NetHost) > 0 {
			// TODO: implement webhook net trust ca bundle
			nwh, err := net.New(opts.Webhook.NetHost, true)
			if err != nil {
				return err
			}

			whClients = append(whClients, nwh)
		}

		wh := webhook.New(&opts.DriverID, whClients...)

		d, err := driver.New(&opts.DriverID, opts.Endpoint,
			opts.DataRoot, opts.TmpfsSize, wh)
		if err != nil {
			return err
		}

		if err := wh.Register(); err != nil {
			return err
		}

		d.Run()
		return nil
	},
}
