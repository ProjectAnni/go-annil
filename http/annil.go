package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func regAnniEndpoints(r *gin.Engine) {
	r.GET("/albums", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, be.ListCatalogs())
	})
	// TODO
}
