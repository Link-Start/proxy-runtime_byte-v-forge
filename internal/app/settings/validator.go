package settings

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

type ProfileValidator func([]*proxyruntimev1.EgressProfileSettings) error

func (a Application) validateProfiles(profiles []*proxyruntimev1.EgressProfileSettings) error {
	if a.validateProfilesFunc == nil {
		return nil
	}
	return a.validateProfilesFunc(profiles)
}
