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

package renew

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/jetstack/cert-manager/pkg/util/pki"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager-csi/pkg/util"
)

type Renewer struct {
	dataDir string

	watchingVols map[string]chan struct{}
	muVol        sync.RWMutex

	renewFunc RenewFunc
}

type certToWatch struct {
	base     string
	metaData *csiapi.MetaData

	notBefore, notAfter time.Time
}

type RenewFunc func(vol *csiapi.MetaData) (*x509.Certificate, error)

func New(dataDir string, renewFunc RenewFunc) *Renewer {
	return &Renewer{
		dataDir:      dataDir,
		watchingVols: make(map[string]chan struct{}),
		renewFunc:    renewFunc,
	}
}

func (r *Renewer) Discover() error {
	glog.Infof("renewer: starting discovery on %q", r.dataDir)

	certsToWatch, err := r.walkDir()
	if err != nil {
		return err
	}

	for v := range r.watchingVols {
		r.KillWatcher(v)
	}

	var errs []string
	for _, f := range certsToWatch {
		glog.Infof("renewer: watching new volume for certificate renewal %q", f.base)

		if err := r.WatchCert(f.metaData, f.notBefore, f.notAfter); err != nil {
			errs = append(errs, fmt.Sprintf("%q: %s",
				f.metaData.ID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to start watching certs: %s",
			strings.Join(errs, ", "))
	}

	return nil
}

func (r *Renewer) walkDir() ([]certToWatch, error) {
	files, err := ioutil.ReadDir(r.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data dir: %s", err)
	}

	var errs []string
	var certsToWatch []certToWatch
	for _, f := range files {
		fPath := filepath.Join(r.dataDir, f.Name())

		glog.V(4).Infof("renewer: trying discovery on %q", fPath)

		// not a directory or not a csi directory
		base := filepath.Base(fPath)
		if !f.IsDir() ||
			!strings.HasPrefix(base, "csi-") {
			glog.V(4).Infof("renewer: file not a directory or doesn't have \"cert-manger-csi\" prefix: %q", base)
			continue
		}

		metaPath := filepath.Join(fPath, csiapi.MetaDataFileName)
		b, err := ioutil.ReadFile(metaPath)
		if err != nil {
			// meta data file doesn't exist, move on
			if os.IsNotExist(err) {
				glog.V(4).Infof("renewer: metadata file not found: %q", metaPath)
				continue
			}

			errs = append(errs,
				fmt.Sprintf("failed to read metadata file: %s", err))
			continue
		}

		metaData := new(csiapi.MetaData)
		if err := json.Unmarshal(b, metaData); err != nil {
			errs = append(errs,
				fmt.Sprintf("failed to unmarshal metadata file for %q: %s", f.Name(), err.Error()))
			continue
		}

		keyBytes, err := r.readFile(fPath, metaData.Attributes[csiapi.KeyFileKey])
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if _, err := pki.DecodePrivateKeyBytes(keyBytes); err != nil {
			errs = append(errs, fmt.Sprintf("%q: failed to parse key file: %s",
				f.Name(), err))
			continue
		}

		certBytes, err := r.readFile(fPath, metaData.Attributes[csiapi.CertFileKey])
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		cert, err := pki.DecodeX509CertificateBytes(certBytes)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%q: failed to parse cert file: %s",
				f.Name(), err))
			continue
		}

		certsToWatch = append(certsToWatch, certToWatch{
			base:      base,
			metaData:  metaData,
			notBefore: cert.NotBefore,
			notAfter:  cert.NotAfter,
		})
	}

	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ", "))
	}

	return certsToWatch, nil
}

func (r *Renewer) WatchCert(metaData *csiapi.MetaData, notBefore, notAfter time.Time) error {
	r.muVol.Lock()
	defer r.muVol.Unlock()

	if _, ok := r.watchingVols[metaData.ID]; ok {
		glog.Errorf("volume already being watched, aborting second watcher: %s",
			metaData.ID)
		return nil
	}

	ch := make(chan struct{})
	r.watchingVols[metaData.ID] = ch

	renewalTime, err := util.RenewTimeFromNotAfter(notBefore, notAfter, metaData.Attributes[csiapi.RenewBeforeKey])
	if err != nil {
		return fmt.Errorf("failed to watch certificate %q: %s",
			metaData.ID, err)
	}

	timer := time.NewTimer(renewalTime)

	glog.Infof("renewer: renewal set for certificate in %s: %q", renewalTime, metaData.ID)

	go func() {
		defer timer.Stop()

		select {
		case <-ch:
			return
		case <-timer.C:
			cert, err := r.renewFunc(metaData)
			if err != nil {
				glog.Errorf("renewer: failed to renew certificate %q: %s",
					metaData.ID, err)
				return
			}

			delete(r.watchingVols, metaData.ID)

			if err := r.WatchCert(metaData, cert.NotBefore, cert.NotAfter); err != nil {
				glog.Errorf("renewer: failed to watch certificate %q: %s",
					metaData.ID, err)
			}
		}
	}()

	return nil
}

func (r *Renewer) KillWatcher(volID string) {
	r.muVol.RLock()
	defer r.muVol.RUnlock()

	ch, ok := r.watchingVols[volID]
	if ok {
		glog.Infof("renewer: killing watcher for %q", volID)
		close(ch)
	}
}

func (r *Renewer) readFile(rootPath, path string) ([]byte, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("%q: read path is empty from attributes file",
			rootPath)
	}

	path = filepath.Join(rootPath, "data", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%q: failed to read file: %s",
			path, err)
	}

	return b, nil
}
