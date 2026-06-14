package settings

import "errors"

var ErrRepositoryRequired = errors.New("runtime settings repository is not configured")

type ApplyScheduler func([]string)

type Dependencies struct {
	Repository           Repository
	ScheduleApply        ApplyScheduler
	Logger               Logger
	ValidateProfiles     ProfileValidator
	IPFraudProviderViews IPFraudProviderViews
	IPGeoProviderViews   IPGeoProviderViews
}

type Application struct {
	repository           Repository
	scheduleApply        ApplyScheduler
	logger               Logger
	validateProfilesFunc ProfileValidator
	ipFraudProviderViews IPFraudProviderViews
	ipGeoProviderViews   IPGeoProviderViews
}

func NewApplication(deps Dependencies) Application {
	return Application{repository: deps.Repository, scheduleApply: deps.ScheduleApply, logger: deps.Logger, validateProfilesFunc: deps.ValidateProfiles, ipFraudProviderViews: deps.IPFraudProviderViews, ipGeoProviderViews: deps.IPGeoProviderViews}
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
