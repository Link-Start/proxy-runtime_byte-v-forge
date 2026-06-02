package app

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *redisLeaseStore) loadIndex(ctx context.Context) []string {
	value, ok, err := s.store.Load(ctx, leaseIndexKey)
	if err != nil || !ok {
		return nil
	}
	var ids []string
	_ = json.Unmarshal([]byte(value), &ids)
	return cleanIDs(ids)
}

func (s *redisLeaseStore) saveIndex(ctx context.Context, ids []string) error {
	data, err := json.Marshal(cleanIDs(ids))
	if err != nil {
		return err
	}
	return s.store.SaveTTL(ctx, leaseIndexKey, string(data), 0)
}

func (s *redisLeaseStore) updateIndex(ctx context.Context, update func([]string) []string) error {
	lock, err := s.locker.Lock(ctx, "index")
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock(ctx) }()
	return s.saveIndex(ctx, update(s.loadIndex(ctx)))
}

func (s *redisLeaseStore) pruneStaleIndex(ctx context.Context, staleIDs []string) error {
	staleIndex := leaseIDIndex(staleIDs)
	if len(staleIndex) == 0 {
		return nil
	}
	lock, err := s.locker.Lock(ctx, "index")
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock(ctx) }()
	now := time.Now().UTC()
	out := make([]string, 0)
	for _, id := range s.loadIndex(ctx) {
		if _, stale := staleIndex[id]; !stale {
			out = append(out, id)
			continue
		}
		lease, err := s.loadLease(ctx, id)
		if err == nil && leaseActive(lease, now) {
			out = append(out, id)
			continue
		}
		if err != nil && !errors.Is(err, errLeaseNotFound) {
			out = append(out, id)
		}
	}
	return s.saveIndex(ctx, out)
}

var errLeaseNotFound = errors.New("active proxy lease not found")

func leaseKey(accountID string) string { return "lease:" + strings.TrimSpace(accountID) }

func leaseTTL(lease *proxyruntimev1.ProxyDynamicLease) time.Duration {
	if lease.GetExpiresAt() == nil {
		return defaultLeaseTTL
	}
	ttl := time.Until(lease.GetExpiresAt().AsTime())
	if ttl <= 0 {
		return time.Second
	}
	return ttl
}

func appendLeaseIndex(ids []string, id string) []string {
	ids = append(ids, id)
	return cleanIDs(ids)
}

func removeLeaseIndex(ids []string, id string) []string {
	id = strings.TrimSpace(id)
	out := make([]string, 0, len(ids))
	for _, item := range ids {
		if strings.TrimSpace(item) != "" && strings.TrimSpace(item) != id {
			out = append(out, item)
		}
	}
	return cleanIDs(out)
}

func leaseIDIndex(ids []string) map[string]struct{} {
	index := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			index[id] = struct{}{}
		}
	}
	return index
}

func cleanIDs(ids []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
