package mihomonative

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func applyVLESSSecurityOptions(config map[string]any, query url.Values, security string) {
	if security == "tls" || security == "reality" {
		config["tls"] = true
	}
	if serverName := appcore.FirstNonEmpty(query.Get("sni"), query.Get("servername")); serverName != "" {
		config["servername"] = serverName
	}
	if security != "reality" {
		return
	}
	reality := map[string]any{}
	if value := appcore.FirstNonEmpty(query.Get("pbk"), query.Get("public-key")); value != "" {
		reality["public-key"] = value
	}
	if value := appcore.FirstNonEmpty(query.Get("sid"), query.Get("short-id")); value != "" {
		reality["short-id"] = value
	}
	if len(reality) > 0 {
		config["reality-opts"] = reality
	}
}

func applyVLESSNetworkOptions(config map[string]any, query url.Values, network string) error {
	switch network {
	case "ws", "websocket":
		config["network"] = "ws"
		opts := map[string]any{}
		if value := strings.TrimSpace(query.Get("path")); value != "" {
			opts["path"] = value
		}
		if value := strings.TrimSpace(query.Get("host")); value != "" {
			opts["headers"] = map[string]string{"Host": value}
		}
		if len(opts) > 0 {
			config["ws-opts"] = opts
		}
	case "grpc":
		opts := map[string]any{}
		if value := appcore.FirstNonEmpty(query.Get("serviceName"), query.Get("service-name")); value != "" {
			opts["grpc-service-name"] = value
		}
		if len(opts) > 0 {
			config["grpc-opts"] = opts
		}
	case "tcp":
	default:
		return fmt.Errorf("unsupported vless network %q", network)
	}
	return nil
}
