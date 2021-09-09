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

package app

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/cert-manager/csi-lib/driver"
	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/metadata"
	"github.com/cert-manager/csi-lib/storage"
	"github.com/spf13/cobra"
	"k8s.io/utils/clock"

	"github.com/cert-manager/csi-driver/cmd/app/options"
	"github.com/cert-manager/csi-driver/pkg/filestore"
	"github.com/cert-manager/csi-driver/pkg/keygen"
	"github.com/cert-manager/csi-driver/pkg/requestgen"
)

const (
	helpOutput = "Container Storage Interface driver to issue certificates from cert-manager"
)

// NewCommand will return a new command instance for the cert-manager CSI driver.
func NewCommand(ctx context.Context) *cobra.Command {
	opts := options.New()

	cmd := &cobra.Command{
		Use:   "cert-manager-csi",
		Short: helpOutput,
		Long:  helpOutput,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Complete()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := opts.Logr.WithName("main")
			log.Info("building driver")

			store, err := storage.NewFilesystem(opts.Logr.WithName("storage"), opts.DataRoot)
			if err != nil {
				return fmt.Errorf("failed to setup filesystem: %w")
			}

			keyGenerator := keygen.Generator{Store: store}
			writer := filestore.Writer{Store: store}

			d, err := driver.New(opts.Endpoint, opts.Logr.WithName("driver"), driver.Options{
				DriverName:    opts.DriverName,
				DriverVersion: "v0.1.0",
				NodeID:        opts.NodeID,
				Store:         store,
				Manager: manager.NewManagerOrDie(manager.Options{
					Client:             opts.CMClient,
					MetadataReader:     store,
					Clock:              clock.RealClock{},
					Log:                opts.Logr.WithName("manager"),
					NodeID:             opts.NodeID,
					GeneratePrivateKey: keyGenerator.KeyForMetadata,
					GenerateRequest:    requestgen.RequestForMetadata,
					SignRequest:        signRequest,
					WriteKeypair:       writer.WriteKeypair,
				}),
			})
			if err != nil {
				return fmt.Errorf("failed to setup driver: " + err.Error())
			}

			go func() {
				<-ctx.Done()
				log.Info("shutting down driver", "context", ctx.Err())
				d.Stop()
			}()

			log.Info("running driver")
			if err := d.Run(); err != nil {
				return fmt.Errorf("failed running driver: " + err.Error())
			}

			return nil
		},
	}

	opts = opts.Prepare(cmd)

	return cmd
}

// signRequest will sign a X.509 certificate signing request with the provided
// private key.
func signRequest(_ metadata.Metadata, key crypto.PrivateKey, request *x509.CertificateRequest) ([]byte, error) {
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, request, key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDer,
	}), nil
}
