package app

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func rejectMissingProxyUserProfiles(users []config.ProxyUserRoute, profiles []*proxyruntimev1.EgressProfileSettings) error {
	referenced := map[string]struct{}{}
	for _, user := range users {
		if strings.TrimSpace(user.Route) != config.ListenerRouteProfile {
			continue
		}
		if id := runtimeSafeID(user.ProfileID); id != "" {
			referenced[id] = struct{}{}
		}
	}
	if len(referenced) == 0 {
		return nil
	}
	enabled := enabledEgressProfileIDsFromProfiles(profiles)
	for id := range referenced {
		if _, exists := enabled[id]; !exists {
			return failedPrecondition(fmt.Sprintf("proxy user profile %q is not enabled", id), nil)
		}
	}
	return nil
}
