package ten24

import (
	"fmt"
	"strings"
)

type Config struct {
	APIURL    string
	APIRegion string
	APIFormat string
	APITime   string
	APINum    string
	APIType   string
	Username  string
	Password  string
	Protocol  string
}

func (c Config) HasRuntimeConfig() bool {
	return strings.TrimSpace(c.APIURL) != ""
}

func (c Config) Validate() error {
	switch c.Protocol {
	case "", "http", "socks5":
		return nil
	default:
		return fmt.Errorf("unsupported 1024proxy protocol %q", c.Protocol)
	}
}
