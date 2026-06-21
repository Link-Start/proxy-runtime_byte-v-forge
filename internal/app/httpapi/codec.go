package httpapi

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"
	"google.golang.org/protobuf/proto"
)

var ErrRequestBodyTooLarge = errors.New("request body too large")

func ReadRequestBody(req *http.Request, maxBytes int64) ([]byte, error) {
	defer req.Body.Close()
	data, err := io.ReadAll(io.LimitReader(req.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrRequestBodyTooLarge
	}
	return data, nil
}

func EmptyBody(data []byte) bool {
	return len(strings.TrimSpace(string(data))) == 0
}

func MarshalProto(message proto.Message) ([]byte, error) {
	return protojsoncodec.Marshal(message)
}

func UnmarshalProto(data []byte, message proto.Message) error {
	return protojsoncodec.Unmarshal(data, message)
}
