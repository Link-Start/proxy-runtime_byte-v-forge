package lease

type Application struct {
	repository  Repository
	coordinator Coordinator
	worker      Worker
}

func NewApplication(repository Repository, coordinator Coordinator, worker Worker) *Application {
	return &Application{repository: repository, coordinator: coordinator, worker: worker}
}
