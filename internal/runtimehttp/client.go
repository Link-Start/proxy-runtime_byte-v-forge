package runtimehttp

import (
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/httpclient"
)

var HTTPProxySchemes = httpclient.HTTPProxySchemes
var CommonProxySchemes = httpclient.CommonProxySchemes

func New(timeout time.Duration) *http.Client {
	client, err := httpclient.New(normalizeTimeout(timeout), "")
	if err != nil {
		return &http.Client{Timeout: normalizeTimeout(timeout)}
	}
	return client
}

func NewWithProxy(timeout time.Duration, proxyRawURL string, schemes ...string) (*http.Client, error) {
	return httpclient.NewWithSchemes(normalizeTimeout(timeout), proxyRawURL, schemes...)
}

func normalizeTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 30 * time.Second
	}
	return timeout
}
