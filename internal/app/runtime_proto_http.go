package app

import (
	"errors"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/httpx"
	"github.com/byte-v-forge/common-lib/protojsonx"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (api *runtimeHTTPAPI) readProto(w http.ResponseWriter, req *http.Request, message proto.Message) bool {
	if req.Body == nil {
		writeHTTPError(w, invalidArgument("request body is required", nil), http.StatusBadRequest)
		return false
	}
	body, err := readRequestBody(req)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadRequest)
		return false
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return true
	}
	if err := protojsonx.Unmarshal(body, message); err != nil {
		writeHTTPError(w, err, http.StatusBadRequest)
		return false
	}
	return true
}

func (api *runtimeHTTPAPI) readOptionalProto(w http.ResponseWriter, req *http.Request, message proto.Message) bool {
	if req.Body == nil || req.ContentLength == 0 {
		return true
	}
	body, err := readRequestBody(req)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadRequest)
		return false
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return true
	}
	if err := protojsonx.Unmarshal(body, message); err != nil {
		writeHTTPError(w, err, http.StatusBadRequest)
		return false
	}
	return true
}

func (api *runtimeHTTPAPI) writeProto(w http.ResponseWriter, message proto.Message) {
	data, err := protojsonx.Marshal(message)
	if err != nil {
		writeHTTPError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func readRequestBody(req *http.Request) ([]byte, error) {
	defer req.Body.Close()
	return httpx.ReadLimited(req.Body, 1<<20)
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeHTTPError(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
}

func writeHTTPError(w http.ResponseWriter, err error, fallbackStatus int) {
	httpStatus, code, message := httpErrorDetails(err, fallbackStatus)
	data, marshalErr := protojsonx.Marshal(grpcstatus.New(code, message).Proto())
	if marshalErr != nil {
		httpStatus = http.StatusInternalServerError
		data = []byte(`{"code":13,"message":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(httpStatus)
	_, _ = w.Write(data)
}
