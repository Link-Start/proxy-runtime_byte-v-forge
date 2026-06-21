package dashboard

import (
	"encoding/json"
	"fmt"
)

const endpointID = "proxy-gateway-mihomo"

type BootstrapOptions struct {
	EndpointURL  string
	UIURL        string
	AuthRequired bool
}

func BootstrapHTML(opts BootstrapOptions) ([]byte, error) {
	payload, err := json.Marshal(struct {
		EndpointID   string `json:"endpointID"`
		EndpointURL  string `json:"endpointURL"`
		UIURL        string `json:"uiURL"`
		AuthRequired bool   `json:"authRequired"`
	}{
		EndpointID:   endpointID,
		EndpointURL:  opts.EndpointURL,
		UIURL:        opts.UIURL,
		AuthRequired: opts.AuthRequired,
	})
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>Proxy Runtime</title></head>
<body>
<script>
const config = %s;
const endpoint = {
  id: config.endpointID,
  url: new URL(config.endpointURL, window.location.origin).href.replace(/\/$/, ''),
  secret: ''
};
window.localStorage.setItem('proxyGatewayControlAuthRequired', config.authRequired ? 'true' : 'false');
window.localStorage.setItem('endpointList', JSON.stringify([endpoint]));
window.localStorage.setItem('selectedEndpoint', endpoint.id);
window.location.replace(new URL(config.uiURL, window.location.origin).href);
</script>
</body>
</html>`, payload)), nil
}
