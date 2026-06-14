package settings

type Logger interface {
	Warn(string, ...any)
}
