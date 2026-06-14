package lease

type Dependencies struct {
	Repository  Repository
	Coordinator Coordinator
	Worker      Worker
	Clock       Clock
	Logger      Logger
}

type Application struct {
	repository  Repository
	coordinator Coordinator
	worker      Worker
	clock       Clock
	logger      Logger
}

func NewApplication(deps Dependencies) *Application {
	clock := deps.Clock
	if clock == nil {
		clock = SystemClock{}
	}
	return &Application{
		repository:  deps.Repository,
		coordinator: deps.Coordinator,
		worker:      deps.Worker,
		clock:       clock,
		logger:      deps.Logger,
	}
}
