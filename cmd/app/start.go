/*
Copyright 2019 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
		d, err := driver.New(opts.DriverName, opts.NodeID, opts.Endpoint, opts.DataRoot, opts.TmpfsSize)
		if err != nil {
			return err
		}

		d.Run()
		return nil
	},
}
