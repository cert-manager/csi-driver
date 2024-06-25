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
	"time"

	"github.com/cert-manager/cert-manager/pkg/server"
	"github.com/cert-manager/csi-lib/driver"
	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/manager/util"
	"github.com/cert-manager/csi-lib/metadata"
	"github.com/cert-manager/csi-lib/storage"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/utils/clock"

	"github.com/cert-manager/csi-driver/cmd/app/options"
	"github.com/cert-manager/csi-driver/internal/metrics"
	"github.com/cert-manager/csi-driver/internal/version"
	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
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
		Use:   "cert-manager-csi-driver",
		Short: helpOutput,
		Long:  helpOutput,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Complete()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := opts.Logr.WithName("main")
			log.Info("Starting driver", "version", version.VersionInfo())

			store, err := storage.NewFilesystem(opts.Logr.WithName("storage"), opts.DataRoot)
			if err != nil {
				return fmt.Errorf("failed to setup filesystem: %w", err)
			}
			store.FSGroupVolumeAttributeKey = csiapi.FSGroupKey

			keyGenerator := keygen.Generator{Store: store}
			writer := filestore.Writer{Store: store}

			var clientForMeta manager.ClientForMetadataFunc
			if opts.UseTokenRequest {
				clientForMeta = util.ClientForMetadataTokenRequestEmptyAud(opts.RestConfig)
			}

			mngrlog := opts.Logr.WithName("manager")
			d, err := driver.New(opts.Endpoint, opts.Logr.WithName("driver"), driver.Options{
				DriverName:    opts.DriverName,
				DriverVersion: version.AppVersion,
				NodeID:        opts.NodeID,
				Store:         store,
				Manager: manager.NewManagerOrDie(manager.Options{
					Client:             opts.CMClient,
					ClientForMetadata:  clientForMeta,
					MetadataReader:     store,
					Clock:              clock.RealClock{},
					Log:                &mngrlog,
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

			// Start metrics server
			metricsLn, err := server.Listen("tcp", opts.MetricsListenAddress)
			if err != nil {
				return fmt.Errorf("failed to listen on prometheus address %s: %v", opts.MetricsListenAddress, err)
			}
			metricsServer := metrics.NewServer(metricsLn)

			g, _ := errgroup.WithContext(ctx)
			g.Go(func() error {
				<-ctx.Done()
				log.Info("shutting down driver", "context", ctx.Err())
				d.Stop()
				return nil
			})

			g.Go(func() error {
				log.Info("running driver")
				if err := d.Run(); err != nil {
					return fmt.Errorf("failed running driver: " + err.Error())
				}
				return nil
			})

			g.Go(func() error {
				<-ctx.Done()
				// allow a timeout for graceful shutdown
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// nolint: contextcheck
				return metricsServer.Shutdown(shutdownCtx)
			})

			g.Go(func() error {
				log.V(3).Info("starting metrics server", "address", metricsLn.Addr())
				return metricsServer.Serve(metricsLn)
			})

			return g.Wait()
		},
	}

	opts = opts.Prepare(cmd)

	return cmd
}

// signRequest will sign an X.509 certificate signing request with the provided
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
