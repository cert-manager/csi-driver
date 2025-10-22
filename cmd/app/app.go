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
	"net/http"
	"time"

	"github.com/cert-manager/cert-manager/pkg/client/informers/externalversions"
	"github.com/cert-manager/csi-lib/driver"
	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/manager/util"
	"github.com/cert-manager/csi-lib/metadata"
	csimetrics "github.com/cert-manager/csi-lib/metrics"
	"github.com/cert-manager/csi-lib/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/cert-manager/csi-driver/cmd/app/options"
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
			// Set the controller-runtime logger so that we get the
			// controller-runtime metricsserver logs.
			ctrl.SetLogger(log)

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

			// DRAFT: demo for csi-lib metrics feature
			certRequestInformerFactory := externalversions.NewSharedInformerFactory(opts.CMClient, 5*time.Second)
			certRequestInformer := certRequestInformerFactory.Certmanager().V1().CertificateRequests()
			metricsHandler := csimetrics.New(
				opts.NodeID,
				&opts.Logr,
				ctrlmetrics.Registry.(*prometheus.Registry),
				store,
				certRequestInformer.Lister(),
			)

			mngrlog := opts.Logr.WithName("manager")
			d, err := driver.New(ctx, opts.Endpoint, opts.Logr.WithName("driver"), driver.Options{
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
					Metrics:            metricsHandler,
				}),
			})
			if err != nil {
				return fmt.Errorf("failed to setup driver: %w", err)
			}

			g, gCTX := errgroup.WithContext(ctx)
			g.Go(func() error {
				<-ctx.Done()
				log.Info("shutting down driver", "context", ctx.Err())
				d.Stop()
				return nil
			})

			g.Go(func() error {
				log.Info("running driver")
				if err := d.Run(); err != nil {
					return fmt.Errorf("failed running driver: %w", err)
				}
				return nil
			})

			// Start a metrics server if the --metrics-bind-address is not "0".
			//
			// By default this will serve all the metrics that are registered by
			// controller-runtime to its global metrics registry. Including:
			// * Go Runtime metrics
			// * Process metrics
			// * Various controller-runtime controller metrics
			//   (not updated by csi-driver because it doesn't use controller-runtime)
			// * Leader election metrics
			//   (not updated by csi-driver because it doesn't use leader-election)
			//
			// The full list is here:
			// https://github.com/kubernetes-sigs/controller-runtime/blob/700befecdffa803d19830a6a43adc5779ed01e26/pkg/internal/controller/metrics/metrics.go#L73-L86
			//
			// The advantages of using the controller-runtime metricsserver are:
			// * It already exists and is actively maintained.
			// * Provides optional features for securing the metrics endpoint by
			//   TLS and by authentication with a K8S service account token,
			//   should that be requested by users in the future.
			// * Consistency with cert-manager/approver-policy, which also uses
			//   this library and therefore publishes the same set of
			//   controller-runtime base metrics.
			// Disadvantages:
			// * It introduces a dependency on controller-runtime, which often
			//   introduces breaking changes.
			// * It uses a global metrics registry, which has the usual risks
			//   associated with globals and makes it difficult for us to control
			//   which metrics are published for csi-driver.
			//   https://github.com/kubernetes-sigs/controller-runtime/issues/210
			var unusedHttpClient *http.Client
			metricsServer, err := metricsserver.NewServer(
				metricsserver.Options{
					BindAddress: opts.MetricsBindAddress,
				},
				opts.RestConfig,
				unusedHttpClient,
			)
			if err != nil {
				return err
			}
			if metricsServer != nil {
				g.Go(func() error {
					return metricsServer.Start(gCTX)
				})

				// DRAFT: demo for csi-lib metrics feature
				g.Go(func() error {
					certRequestInformerFactory.Start(gCTX.Done())
					certRequestInformerFactory.WaitForCacheSync(gCTX.Done())
					return nil
				})
			}
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
