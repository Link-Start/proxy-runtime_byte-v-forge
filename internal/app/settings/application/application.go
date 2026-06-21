package application

import (
	"errors"

	"github.com/byte-v-forge/proxy-gateway/internal/config"
)

var ErrRepositoryRequired = errors.New("runtime settings repository is not configured")

type ApplyScheduler func([]string)

type Dependencies struct {
	Repository                  Repository
	ScheduleApply               ApplyScheduler
	Logger                      Logger
	ProxyUsers                  []config.ProxyUserRoute
	ProfileValidationError      ProfileValidationErrorFunc
	IPFraudProviderViews        IPFraudProviderViews
	IPGeoProviderViews          IPGeoProviderViews
	LoadMihomoNativeSettings    MihomoNativeLoader
	UpdateMihomoNativeSettings  MihomoNativeUpdater
	DefaultMihomoNativeSettings MihomoNativeDefault
}

type Application struct {
	repository                  Repository
	scheduleApply               ApplyScheduler
	logger                      Logger
	proxyUsers                  []config.ProxyUserRoute
	profileValidationErrorFunc  ProfileValidationErrorFunc
	ipFraudProviderViews        IPFraudProviderViews
	ipGeoProviderViews          IPGeoProviderViews
	loadMihomoNativeSettings    MihomoNativeLoader
	updateMihomoNativeSettings  MihomoNativeUpdater
	defaultMihomoNativeSettings MihomoNativeDefault
}

func NewApplication(deps Dependencies) Application {
	return Application{
		repository:                  deps.Repository,
		scheduleApply:               deps.ScheduleApply,
		logger:                      deps.Logger,
		proxyUsers:                  append([]config.ProxyUserRoute(nil), deps.ProxyUsers...),
		profileValidationErrorFunc:  deps.ProfileValidationError,
		ipFraudProviderViews:        deps.IPFraudProviderViews,
		ipGeoProviderViews:          deps.IPGeoProviderViews,
		loadMihomoNativeSettings:    deps.LoadMihomoNativeSettings,
		updateMihomoNativeSettings:  deps.UpdateMihomoNativeSettings,
		defaultMihomoNativeSettings: deps.DefaultMihomoNativeSettings,
	}
}

func (a Application) repositoryOrError() (Repository, error) {
	if a.repository == nil {
		return nil, ErrRepositoryRequired
	}
	return a.repository, nil
}

func (a Application) schedule(changedUsernames []string) {
	if a.scheduleApply != nil {
		a.scheduleApply(changedUsernames)
	}
}
