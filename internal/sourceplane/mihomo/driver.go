package mihomo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/processruntime"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

const (
	ProviderID = "mihomo"
	groupName  = "byte-v-forge-source"
)

type Config struct {
	Path         string
	ConfigDir    string
	APIAddr      string
	DashboardDir string
	DashboardURL string
}

type Driver struct {
	cfg    Config
	logger *slog.Logger

	mu           sync.Mutex
	process      *processruntime.Process
	apiClient    *http.Client
	configDir    string
	configPath   string
	signature    string
	sourceSig    string
	sources      sourceFile
	baseCfg      dataplane.Config
	sessions     map[string]dataplane.SessionRoute
	running      bool
	lastError    string
	lastEndpoint sourceplane.Endpoint
	lastGoodData []byte
}

func New(cfg Config, logger *slog.Logger) *Driver {
	if logger == nil {
		logger = slog.Default()
	}
	return &Driver{cfg: cfg, logger: logger, apiClient: runtimehttp.New(5 * time.Second), sessions: map[string]dataplane.SessionRoute{}}
}

func (d *Driver) Name() string { return ProviderID }

func (d *Driver) Reconcile(ctx context.Context, cfg sourceplane.Config) ([]provider.Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.reconcileLocked(ctx, cfg)
}

func (d *Driver) reconcileLocked(ctx context.Context, cfg sourceplane.Config) ([]provider.Node, error) {
	file := cleanSourceFile(sourceFile{Subscriptions: cfg.Providers, FixedProxies: cfg.FixedProxies})
	d.sources = cloneSourceFile(file)
	providers := enabledProviders(file.Subscriptions)
	fixedProxies := enabledFixedProxies(file.FixedProxies)
	endpoint, err := normalizeEndpoint(cfg.Endpoint)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	dir, err := d.ensureConfigDir()
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	baseOptions := renderOptions{
		Providers:           providers,
		FixedProxies:        fixedProxies,
		EgressProfiles:      cfg.EgressProfiles,
		Endpoint:            endpoint,
		ConfigDir:           dir,
		APIAddr:             d.cfg.APIAddr,
		DashboardDir:        firstNonEmpty(d.cfg.DashboardDir, d.baseCfg.DashboardDir),
		DashboardURL:        firstNonEmpty(d.cfg.DashboardURL, d.baseCfg.DashboardURL),
		GroupStrategy:       cfg.GroupStrategy,
		HealthCheckURL:      cfg.HealthCheckURL,
		HealthCheckInterval: cfg.HealthCheckInterval,
		HealthCheckTimeout:  cfg.HealthCheckTimeout,
		BasePool:            d.baseCfg.Pool,
		ProxyUsers:          d.baseCfg.ProxyUsers,
		SessionRoutes:       d.sessionRoutesLocked(),
	}
	configFile, err := renderConfig(baseOptions)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	baseSig := signature(data)
	if err := os.MkdirAll(filepath.Join(dir, "providers"), 0o700); err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	configPath := filepath.Join(dir, "config.json")

	restartRequired := !d.running || d.lastEndpoint != endpoint
	sourceChanged := d.sourceSig != baseSig
	baseReloaded := false
	if restartRequired {
		if err := writeConfigData(configPath, data); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
		d.stopLocked()
		if err := d.startLocked(ctx, dir, configPath); err != nil {
			err = d.withRollback(ctx, configPath, err)
			d.lastError = err.Error()
			return nil, err
		}
		baseReloaded = true
	} else if sourceChanged {
		if err := writeConfigData(configPath, data); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
		if err := d.reloadLocked(ctx, configPath); err != nil {
			err = d.withRollback(ctx, configPath, err)
			d.lastError = err.Error()
			return nil, err
		}
		baseReloaded = true
	}
	if err := waitForEndpoint(ctx, endpoint.Addr, 3*time.Second); err != nil {
		if baseReloaded {
			err = d.withRollback(ctx, configPath, err)
		}
		d.lastError = err.Error()
		return nil, err
	}

	finalOptions := baseOptions
	finalConfig, err := renderConfig(finalOptions)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	finalData, err := json.MarshalIndent(finalConfig, "", "  ")
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	finalSig := signature(finalData)
	if baseReloaded || finalSig != d.signature {
		if err := writeConfigData(configPath, finalData); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
		if err := d.reloadLocked(ctx, configPath); err != nil {
			err = d.withRollback(ctx, configPath, err)
			d.lastError = err.Error()
			return nil, err
		}
		if err := waitForEndpoint(ctx, endpoint.Addr, 3*time.Second); err != nil {
			err = d.withRollback(ctx, configPath, err)
			d.lastError = err.Error()
			return nil, err
		}
	}
	d.signature = finalSig
	d.sourceSig = baseSig
	d.configPath = configPath
	d.lastEndpoint = endpoint
	d.lastGoodData = append(d.lastGoodData[:0], finalData...)
	d.lastError = ""
	return nil, nil
}

func (d *Driver) withRollback(ctx context.Context, configPath string, applyErr error) error {
	if rollbackErr := d.rollbackConfigLocked(ctx, configPath); rollbackErr != nil {
		return fmt.Errorf("%w; rollback failed: %v", applyErr, rollbackErr)
	}
	return applyErr
}

func (d *Driver) rollbackConfigLocked(ctx context.Context, configPath string) error {
	if len(d.lastGoodData) == 0 {
		return nil
	}
	if err := writeConfigData(configPath, d.lastGoodData); err != nil {
		return err
	}
	if d.running {
		if err := d.reloadLocked(ctx, configPath); err == nil {
			if strings.TrimSpace(d.lastEndpoint.Addr) == "" {
				return nil
			}
			return waitForEndpoint(ctx, d.lastEndpoint.Addr, 3*time.Second)
		}
		d.stopLocked()
	}
	if strings.TrimSpace(d.lastEndpoint.Addr) == "" {
		return nil
	}
	dir, err := d.ensureConfigDir()
	if err != nil {
		return err
	}
	if err := d.startLocked(ctx, dir, configPath); err != nil {
		return err
	}
	return waitForEndpoint(ctx, d.lastEndpoint.Addr, 3*time.Second)
}
