package appcore

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

// CleanSecretRefs normalizes, de-duplicates and re-stamps secret refs with the
// given provider/purpose, dropping nil/empty ones.
func CleanSecretRefs(values []*commonv1.SecretRef, provider string, purpose string) []*commonv1.SecretRef {
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

// CloneSecretRef normalizes a single secret ref, returning nil when empty.
func CloneSecretRef(value *commonv1.SecretRef, provider string, purpose string) *commonv1.SecretRef {
	refs := CleanSecretRefs([]*commonv1.SecretRef{value}, provider, purpose)
	if len(refs) == 0 {
		return nil
	}
	return refs[0]
}

// SecretRefConfigured reports whether value is a usable secret ref.
func SecretRefConfigured(value *commonv1.SecretRef) bool {
	return secretref.Configured(value)
}

// ResolveRuntimeSecretRefs resolves the given refs to their plaintext values,
// dropping empties. It returns an error when a resolver is required but absent.
func ResolveRuntimeSecretRefs(ctx context.Context, resolver secretref.Resolver, refs []*commonv1.SecretRef, purpose string) ([]string, error) {
	refs = CleanSecretRefs(refs, "proxy-runtime", purpose)
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
