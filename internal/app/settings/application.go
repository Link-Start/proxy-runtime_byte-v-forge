package settings

import "errors"

var ErrRepositoryRequired = errors.New("runtime settings repository is not configured")

type Application struct {
	repository Repository
}

func NewApplication(repository Repository) Application {
	return Application{repository: repository}
}

func (a Application) repositoryOrError() (Repository, error) {
	if a.repository == nil {
		return nil, ErrRepositoryRequired
	}
	return a.repository, nil
}
