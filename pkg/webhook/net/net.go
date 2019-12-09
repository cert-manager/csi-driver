package net

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

const (
	CreatePath  = "/create"
	RenewPath   = "/renew"
	DestroyPath = "/desroy"
)

type Net struct {
	url    *url.URL
	client *http.Client
}

func New(host string, skipTLSVerify bool) (csiapi.Webhook, error) {
	url, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	return &Net{
		url:    url,
		client: buildHTTPClient(skipTLSVerify),
	}, nil
}

func (n *Net) Create(meta *csiapi.MetaData) {
	n.post(meta, CreatePath)
}

func (n *Net) Renew(meta *csiapi.MetaData) {
	n.post(meta, RenewPath)
}

func (n *Net) Destroy(meta *csiapi.MetaData) {
	n.post(meta, DestroyPath)
}

func (n *Net) post(meta *csiapi.MetaData, path string) {
	timestamp := time.Now()

	b, err := json.Marshal(&csiapi.WebhookPost{
		MetaData:  meta,
		Timestamp: timestamp,
	})
	if err != nil {
		glog.Errorf("webhook/net: failed to marshal POST data: %s",
			err)
		return
	}

	urlPath := n.url.String() + CreatePath

	err = wait.PollImmediate(time.Second/4, time.Second*20,
		func() (bool, error) {
			resp, err := n.client.Post(urlPath, "application/json",
				bytes.NewReader(b))
			if err != nil {
				return false, err
			}

			if resp.StatusCode != 200 {
				glog.Errorf("webhook/net: got %d status code from %s",
					resp.StatusCode, urlPath)
				return false, nil
			}

			return true, nil
		},
	)

	if err != nil {
		glog.Errorf("webhook/net: failed to post %s: %s",
			urlPath, err)
		return
	}

	glog.V(4).Infof("webhook/net: POST %s %s:%s",
		path, meta.ID, timestamp)
}

func buildHTTPClient(skipTLSVerify bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext:           dialTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipTLSVerify},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Second * 30,
	}
}

func dialTimeout(ctx context.Context, network, addr string) (net.Conn, error) {
	d := net.Dialer{Timeout: time.Duration(5 * time.Second)}
	return d.DialContext(ctx, network, addr)
}
