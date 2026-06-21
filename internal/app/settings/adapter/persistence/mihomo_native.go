package persistence

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative"
)

func (s *Store) LoadMihomoNative(ctx context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadMihomoNativeLocked(ctx)
}

func (s *Store) loadMihomoNativeLocked(ctx context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	if s.store == nil {
		return mihomonative.NormalizeSettings(nil), nil
	}
	settings, err := s.store.LoadMihomoNativeSettings(ctx)
	if err != nil {
		return nil, err
	}
	return mihomonative.NormalizeSettings(settings), nil
}

func (s *Store) SaveMihomoNative(ctx context.Context, settings *proxygatewayv1.ProxyGatewayMihomoNativeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveMihomoNativeLocked(ctx, settings)
}

func (s *Store) saveMihomoNativeLocked(ctx context.Context, settings *proxygatewayv1.ProxyGatewayMihomoNativeConfig) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveMihomoNativeSettings(ctx, mihomonative.NormalizeSettings(settings))
}
