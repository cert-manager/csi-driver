package renew

import (
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

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/joshvanl/cert-manager-csi/pkg/certmanager"
)

// TODO (@joshvanl): check for v1alpha1.DisableAutoRenewKey

const (
	metaFileName = "metadata.json"
)

type Renewer struct {
	dataDir string

	watchingVols map[string]chan struct{}
	muVol        sync.RWMutex

	cm *certmanager.CertManager
}

func New(dataDir string, cm *certmanager.CertManager) *Renewer {
	return &Renewer{
		dataDir:      dataDir,
		watchingVols: make(map[string]chan struct{}),
		cm:           cm,
	}
}

func (r *Renewer) Discover() error {
	files, err := ioutil.ReadDir(r.dataDir)
	if err != nil {
		return err
	}

	var errs []string
	for _, f := range files {
		// not a directory or not a csi directory
		base := filepath.Base(f.Name())
		if !f.IsDir() ||
			!strings.HasPrefix(base, "cert-manager-csi") {
			continue
		}

		metaPath := filepath.Join(f.Name(), metaFileName)

		b, err := ioutil.ReadFile(metaPath)
		if err != nil {
			// meta data file doesn't exist, move on
			if os.IsNotExist(err) {
				continue
			}

			return nil
		}

		metaData := new(v1alpha1.MetaData)
		if err := json.Unmarshal(b, metaData); err != nil {
			errs = append(errs,
				fmt.Sprintf("%s: %s", f.Name(), err.Error()))
			continue
		}

		// TODO (@joshvanl): do we really need to check the key?
		keyBytes, err := r.readFile(f.Name(), v1alpha1.KeyFileKey, metaData.Attributes)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if _, err := pki.DecodePrivateKeyBytes(keyBytes); err != nil {
			errs = append(errs, fmt.Sprintf("%s: failed to parse key file: %s",
				f.Name(), err))
			continue
		}

		certBytes, err := r.readFile(f.Name(), v1alpha1.CertFileKey, metaData.Attributes)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		cert, err := pki.DecodeX509CertificateBytes(certBytes)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: failed to parse cert file: %s",
				f.Name(), err))
			continue
		}

		glog.Info("cert-manager: watching volume for certificate renewal %s", base)

		if err := r.WatchFile(metaData, cert.NotAfter); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s",
				f.Name(), err))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func (r *Renewer) WatchFile(metaData *v1alpha1.MetaData, notAfter time.Time) error {
	r.muVol.Lock()
	defer r.muVol.Unlock()

	if _, ok := r.watchingVols[metaData.Name]; ok {
		return fmt.Errorf("volume already being watched, aborting second watcher: %s",
			metaData.Name)
	}

	renewBefore, err := time.ParseDuration(
		metaData.Attributes[v1alpha1.RenewBeforeKey])
	if err != nil {
		return fmt.Errorf("failed to parse renew before: %s", err)
	}

	ch := make(chan struct{})
	r.watchingVols[metaData.Name] = ch

	renewalTime := notAfter.Add(-renewBefore)
	timer := time.NewTimer(time.Until(renewalTime))

	go func() {
		select {
		case <-ch:
			timer.Stop()
			return
		case <-timer.C:
			// cm.RenewCert(metaData) (make sure to check for reuse key
		}
	}()

	return nil
}

func (r *Renewer) readFile(rootPath string, key v1alpha1.Attribute, attr v1alpha1.Attributes) ([]byte, error) {
	path, ok := attr[key]
	if !ok {
		return nil, fmt.Errorf("%s: %s not set in metadata file",
			rootPath, key)
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read key file: %s",
			rootPath, err)
	}

	return b, nil
}
