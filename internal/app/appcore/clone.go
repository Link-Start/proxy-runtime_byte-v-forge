package appcore

import "google.golang.org/protobuf/types/known/durationpb"

// CloneStringMap returns a shallow copy of values, or nil when values is empty.
func CloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

// CloneDuration returns a copy of value, or nil when value is nil.
func CloneDuration(value *durationpb.Duration) *durationpb.Duration {
	if value == nil {
		return nil
	}
	return durationpb.New(value.AsDuration())
}
