/*
Copyright 2026 The cert-manager Authors.

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
	"math"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/cert-manager/csi-driver/cmd/app/options"
)

// newGateBackoffFlagSet registers just the four --gate-backoff-* flags
// (mirroring options.addAppFlags) so gateBackoffConfigFromFlags can be
// exercised standalone, including its FlagSet.Changed() checks.
func newGateBackoffFlagSet(opts *options.Options) *pflag.FlagSet {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.DurationVar(&opts.GateBackoffDuration, "gate-backoff-duration", time.Second, "")
	fs.Float64Var(&opts.GateBackoffFactor, "gate-backoff-factor", 2.0, "")
	fs.Float64Var(&opts.GateBackoffJitter, "gate-backoff-jitter", 0.5, "")
	fs.DurationVar(&opts.GateBackoffCap, "gate-backoff-cap", 10*time.Second, "")
	return fs
}

func TestValidateGateBackoff(t *testing.T) {
	tests := map[string]struct {
		opts    options.Options
		wantErr string
	}{
		"valid defaults": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   0.5,
				GateBackoffCap:      10 * time.Second,
			},
		},
		"zero duration": {
			opts: options.Options{
				GateBackoffCap: time.Second,
			},
			wantErr: "--gate-backoff-duration must be > 0",
		},
		"negative duration": {
			opts: options.Options{
				GateBackoffDuration: -time.Second,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-duration must be > 0",
		},
		"factor below 1": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   0.5,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-factor must be >= 1",
		},
		"jitter negative": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   -0.1,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-jitter must be in [0, 1]",
		},
		"jitter above 1": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   1.1,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-jitter must be in [0, 1]",
		},
		"cap below duration": {
			opts: options.Options{
				GateBackoffDuration: 30 * time.Second,
				GateBackoffFactor:   1,
				GateBackoffCap:      10 * time.Second,
			},
			wantErr: "--gate-backoff-cap (10s) must be 0 (uncapped) or >= --gate-backoff-duration (30s)",
		},
		"cap zero is uncapped and valid": {
			opts: options.Options{
				GateBackoffDuration: 30 * time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   0.5,
				GateBackoffCap:      0,
			},
		},
		"cap negative is invalid": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   0.5,
				GateBackoffCap:      -time.Second,
			},
			wantErr: "--gate-backoff-cap (-1s) must be 0 (uncapped) or >= --gate-backoff-duration (1s)",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateGateBackoff(&test.opts)
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

// TestGateBackoffConfigFromFlags exercises the partial-flag scenarios flagged
// in review: setting only *some* of the --gate-backoff-* flags must still
// produce a fully-populated wait.Backoff (today's known, documented
// limitation - see gateBackoffConfigFromFlags), while leaving *all* of them
// untouched must defer to csi-lib entirely by returning nil.
func TestGateBackoffConfigFromFlags(t *testing.T) {
	tests := map[string]struct {
		setFlags map[string]string // flag name -> value to explicitly set
		want     *wait.Backoff
	}{
		"no flags set defers to csi-lib (nil)": {
			setFlags: nil,
			want:     nil,
		},
		"only duration set pins all four fields": {
			setFlags: map[string]string{"gate-backoff-duration": "5s"},
			want: &wait.Backoff{
				Duration: 5 * time.Second,
				Factor:   2.0, // untouched: pflag's registered default, not csi-lib's
				Jitter:   0.5, // untouched: pflag's registered default, not csi-lib's
				Cap:      10 * time.Second,
				Steps:    math.MaxInt32,
			},
		},
		"only cap set pins all four fields": {
			setFlags: map[string]string{"gate-backoff-cap": "0s"},
			want: &wait.Backoff{
				Duration: time.Second,
				Factor:   2.0,
				Jitter:   0.5,
				Cap:      0,
				Steps:    math.MaxInt32,
			},
		},
		"duration and cap set (typical Helm per-field override) pins all four fields": {
			setFlags: map[string]string{"gate-backoff-duration": "2s", "gate-backoff-cap": "0s"},
			want: &wait.Backoff{
				Duration: 2 * time.Second,
				Factor:   2.0,
				Jitter:   0.5,
				Cap:      0,
				Steps:    math.MaxInt32,
			},
		},
		"all four flags set": {
			setFlags: map[string]string{
				"gate-backoff-duration": "3s",
				"gate-backoff-factor":   "1.5",
				"gate-backoff-jitter":   "0.2",
				"gate-backoff-cap":      "30s",
			},
			want: &wait.Backoff{
				Duration: 3 * time.Second,
				Factor:   1.5,
				Jitter:   0.2,
				Cap:      30 * time.Second,
				Steps:    math.MaxInt32,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opts := &options.Options{}
			fs := newGateBackoffFlagSet(opts)
			for name, value := range test.setFlags {
				require.NoError(t, fs.Set(name, value))
			}

			got := gateBackoffConfigFromFlags(fs, opts)
			assert.Equal(t, test.want, got)
		})
	}
}
