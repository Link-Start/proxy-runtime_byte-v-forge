package app

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

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
