package app

import (
	"errors"
	"net/http"

	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

const maxRuntimeHTTPRequestBodyBytes = 1 << 20

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
	if httpapi.EmptyBody(body) {
		return true
	}
	if err := httpapi.UnmarshalProto(body, message); err != nil {
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
	if httpapi.EmptyBody(body) {
		return true
	}
	if err := httpapi.UnmarshalProto(body, message); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return false
	}
	return true
}

func (api *runtimeHTTPAPI) writeProto(ctx *gin.Context, message proto.Message) {
	data, err := httpapi.MarshalProto(message)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Type", "application/json")
	_, _ = ctx.Writer.Write(data)
}

func readRequestBody(req *http.Request) ([]byte, error) {
	data, err := httpapi.ReadRequestBody(req, maxRuntimeHTTPRequestBodyBytes)
	if errors.Is(err, httpapi.ErrRequestBodyTooLarge) {
		return nil, resourceExhausted("request body exceeds 1MiB", nil)
	}
	return data, err
}

func writeHTTPError(w http.ResponseWriter, err error, fallbackStatus int) {
	httpStatus, code, message := httpErrorDetails(err, fallbackStatus)
	httpapi.WriteError(w, httpStatus, code, message)
}
