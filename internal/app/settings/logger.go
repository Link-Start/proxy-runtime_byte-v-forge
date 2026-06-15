package settings

import "reflect"

type Logger interface {
	Warn(string, ...any)
}

func errorType(err error) string {
	if err == nil {
		return ""
	}
	return reflect.TypeOf(err).String()
}
