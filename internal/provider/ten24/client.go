package ten24

import "net/http"

const providerID = "1024proxy"

type Provider struct {
	cfg        Config
	httpClient *http.Client
}

func New(cfg Config, httpClient *http.Client) *Provider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Provider{cfg: cfg, httpClient: httpClient}
}

func (p *Provider) Name() string {
	return providerID
}

func (p *Provider) RequiresSessionLease() bool {
	return false
}
