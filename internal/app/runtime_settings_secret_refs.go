package app

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

func cleanSecretRefs(values []*commonv1.SecretRef, provider string, purpose string) []*commonv1.SecretRef {
	out := make([]*commonv1.SecretRef, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		ref := secretref.Clone(value, provider, purpose)
		if ref == nil {
			continue
		}
		secretID := ref.GetSecretId()
		if _, exists := seen[secretID]; exists {
			continue
		}
		seen[secretID] = struct{}{}
		out = append(out, ref)
	}
	return out
}

func cloneSecretRef(value *commonv1.SecretRef, provider string, purpose string) *commonv1.SecretRef {
	refs := cleanSecretRefs([]*commonv1.SecretRef{value}, provider, purpose)
	if len(refs) == 0 {
		return nil
	}
	return refs[0]
}

func secretRefConfigured(value *commonv1.SecretRef) bool {
	return secretref.Configured(value)
}

func resolveRuntimeSecretRefs(ctx context.Context, resolver secretref.Resolver, refs []*commonv1.SecretRef, purpose string) ([]string, error) {
	refs = cleanSecretRefs(refs, "proxy-runtime", purpose)
	if len(refs) == 0 {
		return nil, nil
	}
	if resolver == nil {
		return nil, fmt.Errorf("secret resolver is required")
	}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		value, err := resolver.ResolveSecret(ctx, ref)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out, nil
}

func defaultProviderWeight(index int) uint32 {
	if index < 0 {
		return 100
	}
	if index > 9 {
		return 10
	}
	return uint32(100 - index*10)
}
