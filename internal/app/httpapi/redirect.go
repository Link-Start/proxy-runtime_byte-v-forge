package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectToTrailingSlash(ctx *gin.Context) {
	target := *ctx.Request.URL
	target.Path += "/"
	ctx.Redirect(http.StatusTemporaryRedirect, target.String())
}
