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

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager-csi/pkg/certmanager"
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
	glog.V(4).Infof("renewer: starting discovery on %s", r.dataDir)

	files, err := ioutil.ReadDir(r.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data dir: %s", err)
	}

	var errs []string
	for _, f := range files {
		fPath := filepath.Join(r.dataDir, f.Name())

		glog.V(4).Infof("renewer: trying discovery on %s", fPath)

		// not a directory or not a csi directory
		base := filepath.Base(fPath)
		if !f.IsDir() ||
			!strings.HasPrefix(base, "cert-manager-csi") {
			glog.V(4).Infof("renewer: file not a directory or has 'cert-manger-csi' prefix: %s", f.Name())
			continue
		}

		metaPath := filepath.Join(fPath, csiapi.MetaDataFileName)
		b, err := ioutil.ReadFile(metaPath)
		if err != nil {
			// meta data file doesn't exist, move on
			if os.IsNotExist(err) {
				glog.V(4).Infof("renewer: metadata file not found: %s", metaPath)
				continue
			}

			return fmt.Errorf("failed to read metadata file: %s", err)
		}

		metaData := new(csiapi.MetaData)
		if err := json.Unmarshal(b, metaData); err != nil {
			errs = append(errs,
				fmt.Sprintf("failed to unmarshal metadata file for %s: %s", f.Name(), err.Error()))
			continue
		}

		// TODO (@joshvanl): do we really need to check the key?
		keyBytes, err := r.readFile(fPath, metaData.Attributes[csiapi.KeyFileKey])
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if _, err := pki.DecodePrivateKeyBytes(keyBytes); err != nil {
			errs = append(errs, fmt.Sprintf("%s: failed to parse key file: %s",
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
			errs = append(errs, fmt.Sprintf("%s: failed to parse cert file: %s",
				f.Name(), err))
			continue
		}

		glog.Infof("renewer: watching new volume for certificate renewal %s", base)

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

func (r *Renewer) WatchFile(metaData *csiapi.MetaData, notAfter time.Time) error {
	r.muVol.Lock()
	defer r.muVol.Unlock()

	if _, ok := r.watchingVols[metaData.Name]; ok {
		glog.Errorf("volume already being watched, aborting second watcher: %s",
			metaData.Name)
		return nil
	}

	renewBefore, err := time.ParseDuration(
		metaData.Attributes[csiapi.RenewBeforeKey])
	if err != nil {
		return fmt.Errorf("failed to parse renew before: %s", err)
	}

	ch := make(chan struct{})
	r.watchingVols[metaData.Name] = ch

	glog.Infof("renewer: starting to watch certificate for renewal: %s", metaData.Name)

	renewalTime := notAfter.Add(-renewBefore)
	timer := time.NewTimer(time.Until(renewalTime))

	go func() {
		select {
		case <-ch:
			timer.Stop()
			return
		case <-timer.C:
			cert, err := r.cm.RenewCertificate(metaData)
			if err != nil {
				glog.Errorf("renewer: failed to renew certificate %s: %s",
					metaData.Name, err)
				return
			}

			if err := r.WatchFile(metaData, cert.NotBefore); err != nil {
				glog.Errorf("renewer: failed to watch certificate %s: %s",
					metaData.Name, err)
			}
		}
	}()

	return nil
}

func (r *Renewer) KillWatcher(vol *csiapi.MetaData) {
	r.muVol.RLock()
	defer r.muVol.RUnlock()

	ch, ok := r.watchingVols[vol.Name]
	if ok {
		glog.Infof("renewer: killing watcher for %s", vol.Name)
		close(ch)
	}
}

func (r *Renewer) readFile(rootPath, path string) ([]byte, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("%s: read path is empty from attributes file",
			rootPath)
	}

	path = filepath.Join(rootPath, "data", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read file: %s",
			path, err)
	}

	return b, nil
}
