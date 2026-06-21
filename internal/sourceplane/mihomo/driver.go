package mihomo

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
	"github.com/byte-v-forge/proxy-gateway/internal/processruntime"
	"github.com/byte-v-forge/proxy-gateway/internal/runtimehttp"
	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

const (
	ProviderID = "mihomo"
)

type Config struct {
	Path             string
	ConfigDir        string
	APIAddr          string
	ControllerSecret string
	DashboardDir     string
	DashboardURL     string
}

type Driver struct {
	cfg    Config
	logger *slog.Logger

	mu           sync.Mutex
	process      *processruntime.Process
	processLogs  *processLogRing
	apiClient    *http.Client
	configDir    string
	configPath   string
	desiredSig   string
	signature    string
	baseSig      string
	baseCfg      dataplane.Config
	sessions     map[string]dataplane.SessionRoute
	running      bool
	lastError    string
	lastEndpoint sourceplane.Endpoint
}

func New(cfg Config, logger *slog.Logger) *Driver {
	if logger == nil {
		logger = slog.Default()
	}
	return &Driver{cfg: cfg, logger: logger, processLogs: newProcessLogRing(logger, defaultProcessLogRingLimit), apiClient: runtimehttp.New(5 * time.Second), sessions: map[string]dataplane.SessionRoute{}}
}

func (d *Driver) Name() string { return ProviderID }
