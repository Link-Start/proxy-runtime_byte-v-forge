package app

import (
	"strings"
	"time"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
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

func cloneRuntimeSettingsFile(settings *runtimeSettingsFile) *runtimeSettingsFile {
	if settings == nil {
		return nil
	}
	return proto.Clone(settings).(*runtimeSettingsFile)
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

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = strings.TrimSpace(value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func runtimeSafeID(value string) string {
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

func defaultExpectedStatus(status uint32) uint32 {
	if status == 0 {
		return 204
	}
	return status
}
