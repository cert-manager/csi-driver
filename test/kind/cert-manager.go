package kind

import (
	log "github.com/sirupsen/logrus"
)

func (k *Kind) DeployCertManager(version string) error {
	log.Infof("kind: deploying cert-manager version %q", version)

	return nil
}
