package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func writeSettingsLoadHTTPError(ctx *gin.Context, err error) {
	writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
}

func writeSettingsUpdateHTTPError(ctx *gin.Context, err error) {
	writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
}
