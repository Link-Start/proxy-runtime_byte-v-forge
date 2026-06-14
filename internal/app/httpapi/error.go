package httpapi

import (
	"net/http"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func WriteError(w http.ResponseWriter, httpStatus int, code codes.Code, message string) {
	data, marshalErr := MarshalProto(grpcstatus.New(code, message).Proto())
	if marshalErr != nil {
		httpStatus = http.StatusInternalServerError
		data = []byte(`{"code":13,"message":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(httpStatus)
	_, _ = w.Write(data)
}
