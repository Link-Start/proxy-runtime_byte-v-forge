package lease

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
}
