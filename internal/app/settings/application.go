package settings

import "errors"

var ErrRepositoryRequired = errors.New("runtime settings repository is not configured")

type ApplyScheduler func([]string)

type Dependencies struct {
	Repository       Repository
	ScheduleApply    ApplyScheduler
	Logger           Logger
	ValidateProfiles ProfileValidator
}

type Application struct {
	repository           Repository
	scheduleApply        ApplyScheduler
	logger               Logger
	validateProfilesFunc ProfileValidator
}

func NewApplication(deps Dependencies) Application {
	return Application{repository: deps.Repository, scheduleApply: deps.ScheduleApply, logger: deps.Logger, validateProfilesFunc: deps.ValidateProfiles}
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
