package app

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) authorize(ctx *gin.Context) bool {
	decision := api.auth.Authorize(ctx.Request, time.Now(), controlPlaneHTTPPrefix)
	if decision.Authorized {
		return true
	}
	if decision.RedirectURL != "" {
		ctx.Redirect(http.StatusSeeOther, decision.RedirectURL)
		return false
	}
	if decision.Challenge != "" {
		ctx.Header("WWW-Authenticate", decision.Challenge)
	}
	writeHTTPError(ctx.Writer, errors.New("unauthorized"), http.StatusUnauthorized)
	return false
}
