package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSourceApplication struct {
	runtime *Runtime
}

func newRuntimeSourceApplication(runtime *Runtime) runtimeSourceApplication {
	return runtimeSourceApplication{runtime: runtime}
}

func (s *RuntimeService) ListProxySources(ctx context.Context, _ *proxyruntimev1.ListProxySourcesRequest) (*proxyruntimev1.ListProxySourcesResponse, error) {
	return s.sources.ListProxySources(ctx)
}

func (s *RuntimeService) UpsertProxySubscriptionSource(ctx context.Context, req *proxyruntimev1.UpsertProxySubscriptionSourceRequest) (*proxyruntimev1.UpsertProxySubscriptionSourceResponse, error) {
	return s.sources.UpsertProxySubscriptionSource(ctx, req)
}

func (s *RuntimeService) UpsertProxyFixedSource(ctx context.Context, req *proxyruntimev1.UpsertProxyFixedSourceRequest) (*proxyruntimev1.UpsertProxyFixedSourceResponse, error) {
	return s.sources.UpsertProxyFixedSource(ctx, req)
}

func (s *RuntimeService) DeleteProxySource(ctx context.Context, req *proxyruntimev1.DeleteProxySourceRequest) (*proxyruntimev1.DeleteProxySourceResponse, error) {
	return s.sources.DeleteProxySource(ctx, req)
}

func (s *RuntimeService) ListProxySourceNodes(ctx context.Context, req *proxyruntimev1.ListProxySourceNodesRequest) (*proxyruntimev1.ListProxySourceNodesResponse, error) {
	return s.sources.ListProxySourceNodes(ctx, req)
}

func (a runtimeSourceApplication) ListProxySources(ctx context.Context) (*proxyruntimev1.ListProxySourcesResponse, error) {
	sources, err := a.runtime.listSources(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxySourcesResponse{Sources: sources}, nil
}

func (a runtimeSourceApplication) UpsertProxySubscriptionSource(ctx context.Context, req *proxyruntimev1.UpsertProxySubscriptionSourceRequest) (*proxyruntimev1.UpsertProxySubscriptionSourceResponse, error) {
	if err := a.rejectBlockingSourceMutation(ctx, req.GetSourceId()); err != nil {
		return nil, err
	}
	source, err := a.runtime.store.UpsertSubscriptionSource(ctx, req)
	if err != nil {
		return nil, invalidArgument("", err)
	}
	a.runtime.requestReconcile()
	return &proxyruntimev1.UpsertProxySubscriptionSourceResponse{Source: source}, nil
}

func (a runtimeSourceApplication) UpsertProxyFixedSource(ctx context.Context, req *proxyruntimev1.UpsertProxyFixedSourceRequest) (*proxyruntimev1.UpsertProxyFixedSourceResponse, error) {
	if err := a.rejectBlockingSourceMutation(ctx, req.GetSourceId()); err != nil {
		return nil, err
	}
	source, err := a.runtime.store.UpsertFixedSource(ctx, req)
	if err != nil {
		return nil, invalidArgument("", err)
	}
	a.runtime.requestReconcile()
	return &proxyruntimev1.UpsertProxyFixedSourceResponse{Source: source}, nil
}

func (a runtimeSourceApplication) DeleteProxySource(ctx context.Context, req *proxyruntimev1.DeleteProxySourceRequest) (*proxyruntimev1.DeleteProxySourceResponse, error) {
	if err := a.rejectBlockingSourceMutation(ctx, req.GetSourceId()); err != nil {
		return nil, err
	}
	if err := a.rejectProfileSourceDelete(ctx, req.GetSourceId()); err != nil {
		return nil, err
	}
	if err := a.runtime.store.DeleteSource(ctx, req.GetSourceId()); err != nil {
		return nil, invalidArgument("", err)
	}
	a.runtime.requestReconcile()
	return &proxyruntimev1.DeleteProxySourceResponse{}, nil
}

func (a runtimeSourceApplication) ListProxySourceNodes(ctx context.Context, req *proxyruntimev1.ListProxySourceNodesRequest) (*proxyruntimev1.ListProxySourceNodesResponse, error) {
	nodes, err := a.runtime.sourceNodeSnapshot(ctx, req.GetSourceId())
	if err != nil {
		return nil, unavailable("source runtime is unavailable", err)
	}
	return &proxyruntimev1.ListProxySourceNodesResponse{Nodes: nodes}, nil
}

func (a runtimeSourceApplication) rejectBlockingSourceMutation(ctx context.Context, sourceID string) error {
	sourceID = sourceSafeID(sourceID)
	if sourceID == "" {
		return nil
	}
	leases, err := a.runtime.store.BlockingLeaseFactsBySource(ctx, sourceID)
	if err != nil {
		return err
	}
	if len(leases) > 0 {
		return failedPrecondition("proxy source has active leases", nil)
	}
	return nil
}

func (a runtimeSourceApplication) rejectProfileSourceDelete(ctx context.Context, sourceID string) error {
	sourceID = sourceSafeID(sourceID)
	if sourceID == "" {
		return nil
	}
	settings, err := a.runtime.settings.load(ctx)
	if err != nil {
		return err
	}
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() {
			continue
		}
		if profile.GetLine().GetSource().GetSourceId() == sourceID {
			return failedPrecondition("proxy source is used by an enabled egress profile line", nil)
		}
		if profile.GetExit().GetSource().GetSourceId() == sourceID {
			return failedPrecondition("proxy source is used by an enabled egress profile exit", nil)
		}
	}
	return nil
}
