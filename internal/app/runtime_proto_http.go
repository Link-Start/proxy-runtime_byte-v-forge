package app

import (
	"io"
	"net/http"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
	"github.com/gin-gonic/gin"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (api *runtimeHTTPAPI) readProto(ctx *gin.Context, message proto.Message) bool {
	if ctx.Request.Body == nil {
		writeHTTPError(ctx.Writer, invalidArgument("request body is required", nil), http.StatusBadRequest)
		return false
	}
	body, err := readRequestBody(ctx.Request)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return false
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return true
	}
	if err := protojsoncodec.Unmarshal(body, message); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return false
	}
	return true
}

func (api *runtimeHTTPAPI) readOptionalProto(ctx *gin.Context, message proto.Message) bool {
	if ctx.Request.Body == nil || ctx.Request.ContentLength == 0 {
		return true
	}
	body, err := readRequestBody(ctx.Request)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return false
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return true
	}
	if err := protojsoncodec.Unmarshal(body, message); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return false
	}
	return true
}

func (api *runtimeHTTPAPI) writeProto(ctx *gin.Context, message proto.Message) {
	data, err := protojsoncodec.Marshal(message)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Type", "application/json")
	_, _ = ctx.Writer.Write(data)
}

func readRequestBody(req *http.Request) ([]byte, error) {
	defer req.Body.Close()
	return io.ReadAll(io.LimitReader(req.Body, 1<<20))
}

func writeHTTPError(w http.ResponseWriter, err error, fallbackStatus int) {
	httpStatus, code, message := httpErrorDetails(err, fallbackStatus)
	data, marshalErr := protojsoncodec.Marshal(grpcstatus.New(code, message).Proto())
	if marshalErr != nil {
		httpStatus = http.StatusInternalServerError
		data = []byte(`{"code":13,"message":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(httpStatus)
	_, _ = w.Write(data)
}
