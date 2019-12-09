package webhook

import (
	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

type Webhook struct {
	clients  []csiapi.WebhookClient
	driverID *csiapi.DriverID
}

func New(driverID *csiapi.DriverID, clients ...csiapi.WebhookClient) *Webhook {
	return &Webhook{
		clients: clients,
	}
}

func (w *Webhook) Register() error {
	for _, c := range w.clients {
		if err := c.Register(w.driverID); err != nil {
			return err
		}
	}

	return nil
}

func (w *Webhook) Create(meta *csiapi.MetaData) {
	for _, c := range w.clients {
		c.Create(meta)
	}
}

func (w *Webhook) Renew(meta *csiapi.MetaData) {
	for _, c := range w.clients {
		c.Renew(meta)
	}
}

func (w *Webhook) Destroy(meta *csiapi.MetaData) {
	for _, c := range w.clients {
		c.Destroy(meta)
	}
}
