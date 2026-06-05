package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	proxySourceKindSubscription = "subscription"
	proxySourceKindFixed        = "fixed"
	proxySourceProviderID       = "mihomo"
)

type proxySourceRecord struct {
	SourceID     string
	SourceKind   string
	DisplayName  string
	Enabled      bool
	SourceSecret string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *PostgresStore) ListSources(ctx context.Context, gateways map[string][]accountproxy.Gateway) ([]*proxyruntimev1.ProxySourceDescriptor, error) {
	out := []*proxyruntimev1.ProxySourceDescriptor{}
	accounts, err := s.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.GetStatus() != proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED || len(gateways[account.GetProviderId()]) == 0 {
			continue
		}
		source, err := s.dynamicSource(account, gateways[account.GetProviderId()])
		if err != nil {
			return nil, err
		}
		out = append(out, source)
	}
	managed, err := s.ListManagedSources(ctx)
	if err != nil {
		return nil, err
	}
	return append(out, managed...), nil
}

func (s *PostgresStore) ListManagedSources(ctx context.Context) ([]*proxyruntimev1.ProxySourceDescriptor, error) {
	records, err := s.proxySourceRecords(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*proxyruntimev1.ProxySourceDescriptor, 0, len(records))
	codec := proxySourceCodec{box: s.box}
	for _, record := range records {
		payload, err := codec.decode(record.SourceSecret)
		if err != nil {
			return nil, err
		}
		out = append(out, record.toDescriptor(payload))
	}
	return out, nil
}

func (s *PostgresStore) ListSourcePlaneConfig(ctx context.Context) ([]sourceplane.SubscriptionProvider, []sourceplane.FixedProxy, error) {
	records, err := s.proxySourceRecords(ctx)
	if err != nil {
		return nil, nil, err
	}
	providers := make([]sourceplane.SubscriptionProvider, 0, len(records))
	fixedProxies := make([]sourceplane.FixedProxy, 0, len(records))
	codec := proxySourceCodec{box: s.box}
	for _, record := range records {
		if !record.Enabled {
			continue
		}
		payload, err := codec.decode(record.SourceSecret)
		if err != nil {
			return nil, nil, err
		}
		if record.SourceKind == proxySourceKindSubscription && payload.Subscription != nil {
			providers = append(providers, *payload.Subscription)
		}
		if record.SourceKind == proxySourceKindFixed && payload.FixedProxy != nil {
			fixedProxies = append(fixedProxies, *payload.FixedProxy)
		}
	}
	return providers, fixedProxies, nil
}

func (s *PostgresStore) UpsertSubscriptionSource(ctx context.Context, req *proxyruntimev1.UpsertProxySubscriptionSourceRequest) (*proxyruntimev1.ProxySourceDescriptor, error) {
	sourceID, err := sourceRequestID(req.GetSourceId(), "sub")
	if err != nil {
		return nil, err
	}
	existing, err := s.proxySourceRecord(ctx, sourceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if existing != nil && existing.SourceKind != proxySourceKindSubscription {
		return nil, fmt.Errorf("source_id %q is already used by %s source", sourceID, existing.SourceKind)
	}
	payload := proxySourcePayload{}
	if existing != nil {
		current, err := proxySourceCodec{box: s.box}.decode(existing.SourceSecret)
		if err != nil {
			return nil, err
		}
		payload = current
	}
	current := sourceplane.SubscriptionProvider{ID: sourceID}
	if payload.Subscription != nil {
		current = *payload.Subscription
	}
	current.ID = sourceID
	current.DisplayName = firstNonEmpty(req.GetDisplayName(), current.DisplayName, "Subscription")
	if req.GetClearUrl() {
		current.URL = ""
	}
	if value := strings.TrimSpace(req.GetUrl()); value != "" {
		current.URL = value
	}
	current.Interval = protoDuration(req.GetInterval(), defaultDuration(current.Interval, time.Hour))
	current.Filter = strings.TrimSpace(req.GetFilter())
	current.ExcludeFilter = strings.TrimSpace(req.GetExcludeFilter())
	current.HealthCheckURL = strings.TrimSpace(req.GetHealthCheckUrl())
	current.HealthInterval = protoDuration(req.GetHealthInterval(), defaultDuration(current.HealthInterval, 300*time.Second))
	current.HealthTimeout = protoDuration(req.GetHealthTimeout(), defaultDuration(current.HealthTimeout, 5*time.Second))
	current.HealthLazy = req.GetHealthLazy()
	current.ExpectedStatus = req.GetExpectedStatus()
	if current.ExpectedStatus == 0 {
		current.ExpectedStatus = 204
	}
	current.RegionCodes = cleanRegionCodes(req.GetRegionCodes())
	if req.GetEnabled() && strings.TrimSpace(current.URL) == "" {
		return nil, errors.New("enabled subscription requires url")
	}
	payload.Subscription = &current
	secret, err := proxySourceCodec{box: s.box}.encode(payload)
	if err != nil {
		return nil, err
	}
	record, err := s.upsertProxySourceRecord(ctx, proxySourceRecord{SourceID: sourceID, SourceKind: proxySourceKindSubscription, DisplayName: current.DisplayName, Enabled: req.GetEnabled(), SourceSecret: secret})
	if err != nil {
		return nil, err
	}
	return record.toDescriptor(payload), nil
}

func (s *PostgresStore) UpsertFixedSource(ctx context.Context, req *proxyruntimev1.UpsertProxyFixedSourceRequest) (*proxyruntimev1.ProxySourceDescriptor, error) {
	sourceID, err := sourceRequestID(req.GetSourceId(), "fixed")
	if err != nil {
		return nil, err
	}
	existing, err := s.proxySourceRecord(ctx, sourceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if existing != nil && existing.SourceKind != proxySourceKindFixed {
		return nil, fmt.Errorf("source_id %q is already used by %s source", sourceID, existing.SourceKind)
	}
	payload := proxySourcePayload{}
	if existing != nil {
		current, err := proxySourceCodec{box: s.box}.decode(existing.SourceSecret)
		if err != nil {
			return nil, err
		}
		payload = current
	}
	current := sourceplane.FixedProxy{ID: sourceID}
	if payload.FixedProxy != nil {
		current = *payload.FixedProxy
	}
	current.ID = sourceID
	current.DisplayName = firstNonEmpty(req.GetDisplayName(), current.DisplayName, fixedSourceName(req.GetUri()), "Fixed proxy")
	if req.GetClearUri() {
		current.URI = ""
	}
	if value := strings.TrimSpace(req.GetUri()); value != "" {
		current.URI = value
	}
	current.RegionCodes = cleanRegionCodes(req.GetRegionCodes())
	if req.GetEnabled() {
		if strings.TrimSpace(current.URI) == "" {
			return nil, errors.New("enabled fixed proxy requires uri")
		}
		if err := validateFixedProxyURI(current.URI); err != nil {
			return nil, err
		}
	}
	payload.FixedProxy = &current
	secret, err := proxySourceCodec{box: s.box}.encode(payload)
	if err != nil {
		return nil, err
	}
	record, err := s.upsertProxySourceRecord(ctx, proxySourceRecord{SourceID: sourceID, SourceKind: proxySourceKindFixed, DisplayName: current.DisplayName, Enabled: req.GetEnabled(), SourceSecret: secret})
	if err != nil {
		return nil, err
	}
	return record.toDescriptor(payload), nil
}

func (s *PostgresStore) DeleteSource(ctx context.Context, sourceID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM proxy_runtime_sources WHERE source_id=$1`, sourceSafeID(sourceID))
	return err
}

func (s *PostgresStore) proxySourceRecords(ctx context.Context) ([]proxySourceRecord, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+proxySourceColumns()+` FROM proxy_runtime_sources ORDER BY updated_at DESC, source_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []proxySourceRecord{}
	for rows.Next() {
		record, err := scanProxySource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

func (s *PostgresStore) proxySourceRecord(ctx context.Context, sourceID string) (*proxySourceRecord, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+proxySourceColumns()+` FROM proxy_runtime_sources WHERE source_id=$1`, sourceSafeID(sourceID))
	record, err := scanProxySource(row)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *PostgresStore) upsertProxySourceRecord(ctx context.Context, record proxySourceRecord) (*proxySourceRecord, error) {
	row := s.pool.QueryRow(ctx, `
INSERT INTO proxy_runtime_sources (source_id, source_kind, display_name, enabled, source_secret)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (source_id) DO UPDATE SET source_kind=EXCLUDED.source_kind, display_name=EXCLUDED.display_name, enabled=EXCLUDED.enabled, source_secret=EXCLUDED.source_secret, updated_at=now()
RETURNING `+proxySourceColumns(), record.SourceID, record.SourceKind, record.DisplayName, record.Enabled, record.SourceSecret)
	out, err := scanProxySource(row)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func proxySourceColumns() string {
	return `source_id, source_kind, display_name, enabled, source_secret, created_at, updated_at`
}

func scanProxySource(row pgx.Row) (proxySourceRecord, error) {
	var record proxySourceRecord
	err := row.Scan(&record.SourceID, &record.SourceKind, &record.DisplayName, &record.Enabled, &record.SourceSecret, &record.CreatedAt, &record.UpdatedAt)
	return record, err
}

func sourceRequestID(input string, prefix string) (string, error) {
	id := sourceSafeID(input)
	if id != "" {
		return id, nil
	}
	return generatedID(prefix)
}

func sourceSafeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var out strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			out.WriteRune(r)
			continue
		}
		out.WriteByte('-')
	}
	return strings.Trim(out.String(), "-")
}

func protoDuration(value *durationpb.Duration, fallback time.Duration) time.Duration {
	if value == nil || value.AsDuration() <= 0 {
		return fallback
	}
	return value.AsDuration()
}

func defaultDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}
