package app

import (
	"errors"
	"net/http"

	authapp "github.com/byte-v-forge/proxy-gateway/internal/app/auth"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) authorize(ctx *gin.Context) bool {
	decision := api.auth.Authorize(ctx.Request, api.clock.Now(), controlPlaneHTTPPrefix)
	return authapp.WriteAuthorizationDecision(ctx.Writer, ctx.Request, decision, func(w http.ResponseWriter) {
		writeHTTPError(w, errors.New("unauthorized"), http.StatusUnauthorized)
	})
}
