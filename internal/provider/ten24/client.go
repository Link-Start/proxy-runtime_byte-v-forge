package ten24

import (
	"net/http"
	"time"

	"github.com/byte-v-forge/proxy-gateway/internal/runtimehttp"
)

const providerID = "1024proxy"

type Provider struct {
	cfg        Config
	httpClient *http.Client
}

func New(cfg Config, httpClient *http.Client) *Provider {
	if httpClient == nil {
		httpClient = runtimehttp.New(10 * time.Second)
	}
	return &Provider{cfg: cfg, httpClient: httpClient}
}

func (p *Provider) Name() string {
	return providerID
}
