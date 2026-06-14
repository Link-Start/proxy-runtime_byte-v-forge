package lease

type Application struct {
	repository  Repository
	coordinator Coordinator
}

func NewApplication(repository Repository, coordinator Coordinator) *Application {
	return &Application{repository: repository, coordinator: coordinator}
}
