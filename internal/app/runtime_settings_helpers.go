package app

import (
	"strings"

	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/common-lib/secretref"
)

func cleanList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

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

func cleanRegionCodes(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
