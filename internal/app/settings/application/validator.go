package application

import (
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

type ProfileValidationErrorFunc func(string) error

func (a Application) validateProfiles(profiles []*proxyruntimev1.EgressProfileSettings) error {
	referenced := referencedProfileIDs(a.proxyUsers)
	if len(referenced) == 0 {
		return nil
	}
	enabled := enabledEgressProfileIDs(profiles)
	for id := range referenced {
		if _, exists := enabled[id]; !exists {
			return a.profileValidationError(fmt.Sprintf("proxy user profile %q is not enabled", id))
		}
	}
	return nil
}

func (a Application) profileValidationError(message string) error {
	if a.profileValidationErrorFunc != nil {
		return a.profileValidationErrorFunc(message)
	}
	return errors.New(message)
}

func referencedProfileIDs(users []config.ProxyUserRoute) map[string]struct{} {
	referenced := map[string]struct{}{}
	for _, user := range users {
		if strings.TrimSpace(user.Route) != config.ListenerRouteProfile {
			continue
		}
		if id := safeID(user.ProfileID); id != "" {
			referenced[id] = struct{}{}
		}
	}
	return referenced
}

func enabledEgressProfileIDs(profiles []*proxyruntimev1.EgressProfileSettings) map[string]struct{} {
	out := map[string]struct{}{}
	for _, profile := range profiles {
		if id := safeID(profile.GetProfileId()); id != "" && profile.GetEnabled() {
			out[id] = struct{}{}
		}
	}
	return out
}

func safeID(value string) string {
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
