package http

import (
	"github.com/gin-gonic/gin"
	"github.com/markbates/pkger"
	"net/http"
	"strings"
)

func serveFrontend(ctx *gin.Context) {
	f, err := pkger.Open("/front-end")
	if err != nil {
		ctx.String(http.StatusNotFound, "front-end files not found")
		return
	}
	defer f.Close()
	if strings.HasPrefix(ctx.Request.URL.Path, "/.") {
		ctx.String(http.StatusNotFound, "404 not found")
		return
	}
	http.FileServer(f).ServeHTTP(ctx.Writer, ctx.Request)
}
