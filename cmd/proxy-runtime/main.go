package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/byte-v-forge/proxy-runtime/internal/app"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	mihomosource "github.com/byte-v-forge/proxy-runtime/internal/sourceplane/mihomo"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("load config failed", "error", err)
		os.Exit(1)
	}

	ipFraudProviders, err := ipfraud.NewDefaultRegistry()
	if err != nil {
		logger.Error("create IP fraud provider registry failed", "error", err)
		os.Exit(1)
	}
	ipGeoProviders, err := ipgeo.NewDefaultRegistry()
	if err != nil {
		logger.Error("create IP geo provider registry failed", "error", err)
		os.Exit(1)
	}
	proxyProviders, err := providerregistry.NewDefaultRegistry()
	if err != nil {
		logger.Error("create provider registry failed", "error", err)
		os.Exit(1)
	}
	providerHTTPClient, err := app.NewProviderHTTPClient(cfg)
	if err != nil {
		logger.Error("create provider HTTP client failed", "error", err)
		os.Exit(1)
	}
	proxyProvider, err := proxyProviders.NewPoolProvider(cfg, providerHTTPClient)
	if err != nil {
		logger.Error("create provider failed", "error", err)
		os.Exit(1)
	}

	dataPlane := mihomosource.New(mihomosource.Config{
		Path:         cfg.Mihomo.Path,
		ConfigDir:    cfg.Mihomo.ConfigDir,
		APIAddr:      cfg.Mihomo.APIAddr,
		DashboardDir: cfg.Mihomo.DashboardDir,
		DashboardURL: cfg.Mihomo.DashboardURL,
	}, logger)
	store, err := app.NewControlStore(context.Background(), cfg, proxyProviders, logger)
	if err != nil {
		logger.Error("create store failed", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	leaseRuntimeLocks, err := app.NewLeaseRuntimeLocks(context.Background(), cfg)
	if err != nil {
		logger.Error("create lease runtime locks failed", "error", err)
		os.Exit(1)
	}
	defer leaseRuntimeLocks.Close()
	providerConcurrency, err := app.NewProviderAccountConcurrencyLimiter(context.Background(), cfg)
	if err != nil {
		logger.Error("create provider account concurrency limiter failed", "error", err)
		os.Exit(1)
	}
	defer providerConcurrency.Close()
	runtime, err := app.NewRuntime(app.RuntimeDeps{
		Config:              cfg,
		ProxyProvider:       proxyProvider,
		AccountProviders:    proxyProviders,
		IPFraudProviders:    ipFraudProviders,
		IPGeoProviders:      ipGeoProviders,
		DataPlane:           dataPlane,
		Store:               store,
		LeaseLocks:          leaseRuntimeLocks,
		ProviderConcurrency: providerConcurrency,
		ProviderHTTPClient:  providerHTTPClient,
		Logger:              logger,
	})
	if err != nil {
		logger.Error("create runtime failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runtime.Run(ctx); err != nil {
		logger.Error("proxy runtime stopped", "error", err)
		os.Exit(1)
	}
}
