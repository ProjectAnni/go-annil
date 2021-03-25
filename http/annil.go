package http

import (
	"encoding/json"
	"github.com/SeraphJACK/go-annil/token"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type CreateSharePayload struct {
	Audios map[string][]uint8 `json:"audios"`
	Expire uint               `json:"expire"`
}

func regAnniEndpoints(r *gin.Engine) {
	r.GET("/albums", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, be.ListCatalogs())
	})

	r.POST("/share", func(ctx *gin.Context) {
		tok := ctx.GetHeader("Authorization")
		username, err := token.ValidateUserToken(tok)
		if err != nil {
			ctx.Status(http.StatusUnauthorized)
			return
		}
		defer ctx.Request.Body.Close()
		var payload CreateSharePayload
		err = json.NewDecoder(ctx.Request.Body).Decode(&payload)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		sTok, err := token.GenerateShareToken(username, payload.Audios, time.Hour*time.Duration(payload.Expire))
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			log.Printf("Failed to generate share token for %s: %v\n", username, err)
		} else {
			ctx.Header("Content-Type", "text/plain")
			ctx.String(http.StatusOK, sTok)
		}
	})
}
