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
	"math"
	"net/http"

	"github.com/cert-manager/csi-lib/driver"
	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/manager/util"
	"github.com/cert-manager/csi-lib/metadata"
	"github.com/cert-manager/csi-lib/storage"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/cert-manager/csi-driver/cmd/app/options"
	"github.com/cert-manager/csi-driver/internal/version"
	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
	"github.com/cert-manager/csi-driver/pkg/filestore"
	"github.com/cert-manager/csi-driver/pkg/keygen"
	"github.com/cert-manager/csi-driver/pkg/readinessgate"
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

			gates, err := readinessgate.Parse(opts.PodReadinessGates)
			if err != nil {
				return err
			}
			if len(gates) > 0 && !opts.ContinueOnNotReady {
				return fmt.Errorf("--pod-readiness-gate requires --continue-on-not-ready=true")
			}
			if len(gates) > 0 {
				if err := validateGateBackoff(opts); err != nil {
					return err
				}
			}

			mngrlog := opts.Logr.WithName("manager")
			mgrOpts := manager.Options{
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
			}
			if len(gates) > 0 {
				k8sClient, err := kubernetes.NewForConfig(opts.RestConfig)
				if err != nil {
					return fmt.Errorf("failed to build kubernetes client: %w", err)
				}

				// Scope the informer to pods on this node so cache memory is
				// bounded to the local pod count. The driver runs as a
				// DaemonSet, so a node-scoped informer is the right granularity.
				nodeSelector := fields.OneTermEqualSelector("spec.nodeName", opts.NodeID).String()
				podInformerFactory := informers.NewSharedInformerFactoryWithOptions(
					k8sClient,
					0, // no periodic resync; informer events are sufficient
					informers.WithTweakListOptions(func(o *metav1.ListOptions) {
						o.FieldSelector = nodeSelector
					}),
				)
				podLister := podInformerFactory.Core().V1().Pods().Lister()

				podInformerFactory.Start(ctx.Done())
				if !cache.WaitForCacheSync(ctx.Done(), podInformerFactory.Core().V1().Pods().Informer().HasSynced) {
					return fmt.Errorf("failed to sync pod informer cache")
				}
				log.Info("pod informer cache synced", "node", opts.NodeID)

				mgrOpts.ReadyToRequest = readinessgate.NewReadyToRequestFunc(podLister, gates)

				// Only build GateBackoffConfig if the operator explicitly set
				// one of the --gate-backoff-* flags. Our flag defaults happen
				// to mirror csi-lib's own GateBackoffConfig defaults today, but
				// hardcoding them here would silently pin csi-driver to stale
				// values if csi-lib ever changes its defaults. Leaving
				// GateBackoffConfig nil lets csi-lib apply its own (possibly
				// updated) defaults.
				if fs := cmd.Flags(); fs.Changed("gate-backoff-duration") ||
					fs.Changed("gate-backoff-factor") ||
					fs.Changed("gate-backoff-jitter") ||
					fs.Changed("gate-backoff-cap") {
					mgrOpts.GateBackoffConfig = &wait.Backoff{
						Duration: opts.GateBackoffDuration,
						Factor:   opts.GateBackoffFactor,
						Jitter:   opts.GateBackoffJitter,
						Cap:      opts.GateBackoffCap,
						Steps:    math.MaxInt32,
					}
				}
			}

			d, err := driver.New(ctx, opts.Endpoint, opts.Logr.WithName("driver"), driver.Options{
				DriverName:         opts.DriverName,
				DriverVersion:      version.AppVersion,
				NodeID:             opts.NodeID,
				Store:              store,
				ContinueOnNotReady: opts.ContinueOnNotReady,
				Manager:            manager.NewManagerOrDie(mgrOpts),
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

// validateGateBackoff sanity-checks the --gate-backoff-* flag values so a
// misconfigured operator gets a clear startup error rather than a runtime
// retry storm (e.g. Duration=0 → infinite no-wait spin, Jitter>1 → undefined
// behaviour in wait.Backoff).
func validateGateBackoff(opts *options.Options) error {
	if opts.GateBackoffDuration <= 0 {
		return fmt.Errorf("--gate-backoff-duration must be > 0, got %s", opts.GateBackoffDuration)
	}
	if opts.GateBackoffFactor < 1 {
		return fmt.Errorf("--gate-backoff-factor must be >= 1, got %v", opts.GateBackoffFactor)
	}
	if opts.GateBackoffJitter < 0 || opts.GateBackoffJitter > 1 {
		return fmt.Errorf("--gate-backoff-jitter must be in [0, 1], got %v", opts.GateBackoffJitter)
	}
	// wait.Backoff treats Cap == 0 as "no cap" (see its delay() implementation:
	// the cap is only applied when cap > 0), so it's a valid way to request
	// unbounded exponential growth. Any other non-positive or sub-duration cap
	// is nonsensical, since it would either immediately cap every step to
	// less than the base duration or be silently negative.
	if opts.GateBackoffCap != 0 && opts.GateBackoffCap < opts.GateBackoffDuration {
		return fmt.Errorf("--gate-backoff-cap (%s) must be 0 (uncapped) or >= --gate-backoff-duration (%s)", opts.GateBackoffCap, opts.GateBackoffDuration)
	}
	return nil
}
