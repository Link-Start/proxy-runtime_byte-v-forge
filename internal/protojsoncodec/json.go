package protojsoncodec

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var marshalOptions = protojson.MarshalOptions{UseProtoNames: true}
var unmarshalOptions = protojson.UnmarshalOptions{DiscardUnknown: true}

func Marshal(value proto.Message) ([]byte, error) {
	return marshalOptions.Marshal(value)
}

func Unmarshal(data []byte, value proto.Message) error {
	return unmarshalOptions.Unmarshal(data, value)
}
