package lease

import (
	"net"
	"strings"
)

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
