package lease

import (
	"fmt"
	"net"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

const (
	LabelProxyUsername = "proxy_username"
	LabelProxyPassword = "proxy_password"
)

func NewListenerEndpoint(listener Listener, advertisedHost string, fallbackProtocol string) (*proxygatewayv1.ProxyEndpoint, error) {
	hostPort, err := listenerHostPort(listener.Addr)
	if err != nil {
		return nil, err
	}
	host, portValue, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}
	port, err := listenerPort(portValue)
	if err != nil {
		return nil, err
	}
	if advertisedHost != "" && localEndpointHost(host) {
		host = advertisedHost
	}
	labels := appcore.CloneStringMap(listener.Labels)
	if listener.Username != "" || listener.Password != "" {
		if labels == nil {
			labels = map[string]string{}
		}
		labels[LabelProxyUsername] = listener.Username
		labels[LabelProxyPassword] = listener.Password
	}
	return &proxygatewayv1.ProxyEndpoint{Id: listener.ID, Protocol: listenerProtocol(listener, fallbackProtocol), Host: host, Port: port, Labels: labels}, nil
}

func listenerHostPort(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr, nil
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("invalid listener address")
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port), nil
}

func listenerPort(portValue string) (uint32, error) {
	var port uint32
	_, err := fmt.Sscanf(portValue, "%d", &port)
	return port, err
}

func listenerProtocol(listener Listener, fallback string) proxygatewayv1.ProxyProtocol {
	if strings.TrimSpace(listener.Protocol) == "socks5" {
		return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	}
	if strings.TrimSpace(fallback) == "socks5" {
		return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	}
	return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_HTTP
}

func localEndpointHost(host string) bool {
	switch strings.TrimSpace(strings.ToLower(host)) {
	case "127.0.0.1", "localhost", "0.0.0.0", "::1":
		return true
	default:
		return false
	}
}

func AdvertisedHost(configuredHost string, requestHost string, listener Listener) string {
	if host := strings.TrimSpace(configuredHost); host != "" {
		return host
	}
	bindHost := listenerBindHost(listener.Addr)
	if bindHost != "" && !localOnlyHost(bindHost) {
		return bindHost
	}
	if unspecifiedBindHost(bindHost) {
		if host := firstLocalAdvertisedIP(); host != "" {
			return host
		}
	}
	return strings.TrimSpace(requestHost)
}

func listenerBindHost(addr string) string {
	addr = strings.TrimSpace(addr)
	if strings.HasPrefix(addr, ":") {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return strings.Trim(host, "[]")
}

func localOnlyHost(host string) bool {
	host = strings.Trim(strings.TrimSpace(strings.ToLower(host)), "[]")
	if host == "" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip == nil || ip.IsUnspecified() || ip.IsLoopback()
}

func unspecifiedBindHost(host string) bool {
	host = strings.Trim(strings.TrimSpace(strings.ToLower(host)), "[]")
	if host == "" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsUnspecified()
}

func firstLocalAdvertisedIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	ipv6 := ""
	for _, addr := range addrs {
		ip := interfaceIP(addr)
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || !ip.IsGlobalUnicast() {
			continue
		}
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
		if ipv6 == "" {
			ipv6 = ip.String()
		}
	}
	return ipv6
}

func interfaceIP(addr net.Addr) net.IP {
	switch value := addr.(type) {
	case *net.IPNet:
		return value.IP
	case *net.IPAddr:
		return value.IP
	default:
		return nil
	}
}

type DynamicListenerInput struct {
	ID                  string
	Addr                string
	Protocol            string
	Route               string
	AccountID           string
	LeaseID             string
	DefaultUsername     string
	FallbackPassword    string
	IngressRules        []*proxygatewayv1.ProxyIngressRuleSettings
	PlaygroundAccountID string
	PlaygroundRuleID    string
	PlaygroundUsername  string
}

func NewDynamicListener(input DynamicListenerInput) (Listener, error) {
	return NewListener(ListenerInput{
		ID:        input.ID,
		Addr:      input.Addr,
		Protocol:  input.Protocol,
		Route:     input.Route,
		Username:  dynamicListenerUsername(input),
		Password:  dynamicListenerPassword(input),
		AccountID: input.AccountID,
		LeaseID:   input.LeaseID,
	})
}

func dynamicListenerUsername(input DynamicListenerInput) string {
	if input.AccountID == input.PlaygroundAccountID {
		if rule := PlaygroundIngressRule(input.IngressRules, input.PlaygroundRuleID, input.PlaygroundUsername); rule != nil {
			return rule.GetUsername()
		}
	}
	return input.DefaultUsername
}

func dynamicListenerPassword(input DynamicListenerInput) string {
	password := ListenerPassword(input.IngressRules, input.AccountID, input.FallbackPassword)
	if input.AccountID == input.PlaygroundAccountID {
		if rule := PlaygroundIngressRule(input.IngressRules, input.PlaygroundRuleID, input.PlaygroundUsername); rule != nil {
			return appcore.FirstNonEmpty(rule.GetPasswordValue(), password)
		}
	}
	return password
}
