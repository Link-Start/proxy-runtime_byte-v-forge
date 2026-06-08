package proxyurl

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func Parse(raw string, defaultScheme string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, errors.New("proxy value is empty")
	}
	scheme := NormalizeScheme(defaultScheme)
	if strings.Contains(value, "://") {
		return parseURL(value, scheme)
	}
	if strings.Contains(value, "@") {
		return parseAtFormat(value, scheme)
	}
	if strings.Count(value, ":") == 3 {
		return parseHostPortUserPass(value, scheme)
	}
	return parseURL(scheme+"://"+value, scheme)
}

func parseURL(raw string, defaultScheme string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, errors.New("invalid proxy URL")
	}
	if parsed.Scheme == "" {
		parsed.Scheme = NormalizeScheme(defaultScheme)
	}
	if parsed.Host == "" || parsed.Hostname() == "" {
		return nil, errors.New("proxy URL host is required")
	}
	return parsed, nil
}

func parseAtFormat(value string, scheme string) (*url.URL, error) {
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return nil, errors.New("invalid proxy credential format")
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	if strings.Count(left, ":") == 1 && strings.Count(right, ":") == 1 {
		hostPort := left
		userPass := strings.SplitN(right, ":", 2)
		return buildURL(scheme, hostPort, userPass[0], userPass[1])
	}
	return parseURL(scheme+"://"+value, scheme)
}

func parseHostPortUserPass(value string, scheme string) (*url.URL, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 4 {
		return nil, errors.New("invalid host:port:user:password proxy format")
	}
	return buildURL(scheme, parts[0]+":"+parts[1], parts[2], parts[3])
}

func buildURL(scheme string, host string, username string, password string) (*url.URL, error) {
	if strings.TrimSpace(host) == "" {
		return nil, errors.New("proxy host is required")
	}
	return &url.URL{Scheme: NormalizeScheme(scheme), Host: strings.TrimSpace(host), User: url.UserPassword(strings.TrimSpace(username), strings.TrimSpace(password))}, nil
}

func NormalizeScheme(scheme string) string {
	switch strings.ToLower(strings.TrimSpace(scheme)) {
	case "socks5", "socks5h":
		return "socks5"
	case "https":
		return "https"
	case "http", "":
		return "http"
	default:
		return strings.ToLower(strings.TrimSpace(scheme))
	}
}

func RedactedString(value *url.URL) string {
	if value == nil {
		return ""
	}
	clone := *value
	if clone.User != nil {
		clone.User = url.UserPassword(clone.User.Username(), "***")
	}
	return clone.String()
}

func BrowserMap(value *url.URL) (map[string]string, error) {
	if value == nil {
		return nil, errors.New("proxy URL is required")
	}
	if value.Scheme == "" || value.Host == "" || value.Hostname() == "" {
		return nil, errors.New("proxy URL must include scheme and host")
	}
	if value.RawQuery != "" || value.Fragment != "" || (value.Path != "" && value.Path != "/") {
		return nil, errors.New("proxy URL must not include path, query, or fragment")
	}
	proxy := map[string]string{"server": value.Scheme + "://" + value.Host}
	if value.User != nil {
		if username := value.User.Username(); username != "" {
			proxy["username"] = username
		}
		if password, ok := value.User.Password(); ok {
			proxy["password"] = password
		}
	}
	return proxy, nil
}

func Collect(value any) []string {
	var values []string
	collectValue(value, &values)
	return values
}

func collectValue(value any, values *[]string) {
	switch typed := value.(type) {
	case string:
		appendLines(typed, values)
	case []any:
		for _, item := range typed {
			collectValue(item, values)
		}
	case map[string]any:
		if raw := proxyValueFromMap(typed); raw != "" {
			appendLines(raw, values)
			return
		}
		for _, item := range typed {
			collectValue(item, values)
		}
	}
}

func proxyValueFromMap(value map[string]any) string {
	for _, key := range []string{"proxy", "url", "server"} {
		if raw, ok := value[key].(string); ok && strings.TrimSpace(raw) != "" {
			return raw
		}
	}
	host, _ := valueString(value, "host", "hostname", "ip")
	port, _ := valueString(value, "port")
	if host == "" || port == "" {
		return ""
	}
	username, _ := valueString(value, "username", "user")
	password, _ := valueString(value, "password", "pass")
	protocol, _ := valueString(value, "protocol", "scheme", "type")
	if protocol == "" {
		protocol = "http"
	}
	if username == "" && password == "" {
		return fmt.Sprintf("%s://%s:%s", protocol, host, port)
	}
	return fmt.Sprintf("%s://%s@%s:%s", protocol, url.UserPassword(username, password), host, port)
}

func valueString(value map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		switch typed := value[key].(type) {
		case string:
			return strings.TrimSpace(typed), true
		case float64:
			return strconv.Itoa(int(typed)), true
		case int:
			return strconv.Itoa(typed), true
		}
	}
	return "", false
}

func appendLines(raw string, values *[]string) {
	for _, line := range strings.FieldsFunc(raw, func(r rune) bool { return r == '\n' || r == '\r' }) {
		line = strings.TrimSpace(line)
		if line != "" {
			*values = append(*values, line)
		}
	}
}
