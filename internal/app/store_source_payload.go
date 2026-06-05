package app

import (
	"encoding/json"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/secretbox"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type proxySourcePayload struct {
	Subscription *sourceplane.SubscriptionProvider `json:"subscription,omitempty"`
	FixedProxy   *sourceplane.FixedProxy           `json:"fixed_proxy,omitempty"`
}

type proxySourceCodec struct {
	box secretbox.Box
}

func (c proxySourceCodec) decode(secret string) (proxySourcePayload, error) {
	if strings.TrimSpace(secret) == "" {
		return proxySourcePayload{}, nil
	}
	plain, err := c.box.Open(secret)
	if err != nil {
		return proxySourcePayload{}, err
	}
	if len(plain) == 0 {
		return proxySourcePayload{}, nil
	}
	var payload proxySourcePayload
	if err := json.Unmarshal(plain, &payload); err != nil {
		return proxySourcePayload{}, err
	}
	return payload, nil
}

func (c proxySourceCodec) encode(payload proxySourcePayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return c.box.Seal(data)
}
