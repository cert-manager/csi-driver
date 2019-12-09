package net

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

const (
	RegisterPath = "/register"
	CreatePath   = "/create"
	RenewPath    = "/renew"
	DestroyPath  = "/desroy"
)

type Net struct {
	url    *url.URL
	client *http.Client

	driverID *csiapi.DriverID
}

func New(host string, skipTLSVerify bool) (csiapi.WebhookClient, error) {
	url, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	return &Net{
		url:    url,
		client: buildHTTPClient(skipTLSVerify),
	}, nil
}

func (n *Net) Register(id *csiapi.DriverID) error {
	if n.driverID != nil {
		return errors.New("driver has already registered")
	}

	b, err := json.Marshal(id)
	if err != nil {
		fmt.Errorf("failed to marshal driver ID for registration: %s", err)
	}

	urlPath := n.url.String() + CreatePath
	if err := n.post(urlPath, b); err != nil {
		return err
	}

	n.driverID = id

	return nil
}

func (n *Net) Create(meta *csiapi.MetaData) {
	n.postMeta(meta, CreatePath)
}

func (n *Net) Renew(meta *csiapi.MetaData) {
	n.postMeta(meta, RenewPath)
}

func (n *Net) Destroy(meta *csiapi.MetaData) {
	n.postMeta(meta, DestroyPath)
}

func (n *Net) postMeta(meta *csiapi.MetaData, path string) {
	if n.driverID == nil {
		glog.Error("webhook/net: wehbook client not yet registered")
		return
	}

	timestamp := time.Now()

	b, err := json.Marshal(&csiapi.WebhookClientPost{
		DriverID:  n.driverID,
		Timestamp: timestamp,
		MetaData:  meta,
	})
	if err != nil {
		glog.Errorf("webhook/net: failed to marshal POST data: %s",
			err)
		return
	}

	urlPath := n.url.String() + CreatePath

	err = wait.PollImmediate(time.Second/4,
		time.Second*20, func() (bool, error) {
			err := n.post(urlPath, b)
			if isRetryError(err) {
				glog.Error(err.Error())
				return false, nil
			}

			if err != nil {
				return false, err
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

func (n *Net) post(url string, b []byte) error {
	resp, err := n.client.Post(url, "application/json",
		bytes.NewReader(b))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return newRetryError("webhook/net: failed to post %s and decode response (%d): %s",
				url, resp.StatusCode, err)
		}

		return newRetryError("webhook/net: failed to post %s (%d): %s",
			url, resp.StatusCode, respBody)
	}

	return nil
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

type retryError struct {
	error
}

func newRetryError(err string, formatting ...interface{}) *retryError {
	return &retryError{fmt.Errorf(err, formatting...)}
}

func isRetryError(err error) bool {
	_, ok := err.(*retryError)
	return ok
}
