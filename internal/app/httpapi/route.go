package httpapi

import "github.com/gin-gonic/gin"

type Route struct {
	Methods []string
	Path    string
	Handler gin.HandlerFunc
}
