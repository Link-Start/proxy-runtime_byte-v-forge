package lease

import "github.com/byte-v-forge/proxy-gateway/internal/clock"

type Dependencies struct {
	Repository  Repository
	Coordinator Coordinator
	Worker      Worker
	Clock       clock.Clock
	Logger      Logger
}

type Application struct {
	repository  Repository
	coordinator Coordinator
	worker      Worker
	clock       clock.Clock
	logger      Logger
}

func NewApplication(deps Dependencies) *Application {
	clk := deps.Clock
	if clk == nil {
		clk = clock.SystemClock{}
	}
	return &Application{
		repository:  deps.Repository,
		coordinator: deps.Coordinator,
		worker:      deps.Worker,
		clock:       clk,
		logger:      deps.Logger,
	}
}
