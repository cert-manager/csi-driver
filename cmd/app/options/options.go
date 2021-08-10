/*
Copyright 2021 The cert-manager Authors.

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

	// Size in Mbytes to create the tmpfs file system to write and mount from.
	TmpfsSize string
}

func AddFlags(cmd *cobra.Command) *Options {
	var opts Options

	cmd.PersistentFlags().StringVar(&opts.NodeID, "node-id", "", "node ID")
	cmd.MarkPersistentFlagRequired("node-id")

	cmd.PersistentFlags().StringVar(&opts.Endpoint, "endpoint", "", "CSI endpoint")
	cmd.MarkPersistentFlagRequired("endpoint")

	cmd.PersistentFlags().StringVar(&opts.DriverName, "driver-name",
		"csi.cert-manager.io", "name of the driver")

	cmd.PersistentFlags().StringVar(&opts.DataRoot, "data-root",
		"/csi-data-dir", "directory to store ephemeral data")

	cmd.PersistentFlags().StringVar(&opts.TmpfsSize, "tmpfs-size",
		"100", "size in Mbytes to create the tmpfs file system to store ephemeral data")

	return &opts
}
