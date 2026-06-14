package app

import "google.golang.org/protobuf/proto"

func cloneRuntimeSettingsFile(settings *runtimeSettingsFile) *runtimeSettingsFile {
	if settings == nil {
		return nil
	}
	return proto.Clone(settings).(*runtimeSettingsFile)
}
