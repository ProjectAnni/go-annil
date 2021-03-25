package http

import (
	"encoding/json"
	"github.com/SeraphJACK/go-annil/backend"
	"github.com/SeraphJACK/go-annil/storage"
	"github.com/SeraphJACK/go-annil/token"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strconv"
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
		if !storage.AllowShare(username) {
			ctx.Status(http.StatusForbidden)
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

	r.GET("/:catalog/:track", func(ctx *gin.Context) {
		catalog := ctx.Param("catalog")
		track, err := strconv.Atoi(ctx.Param("track"))
		if err != nil || track < 0 || track > 255 {
			ctx.Status(http.StatusBadRequest)
			return
		}
		tok := ctx.GetHeader("Authorization")
		if !token.CheckAudioPerms(tok, catalog, uint8(track)) {
			ctx.Status(http.StatusForbidden)
			return
		}
		typ, aud, err := be.GetAudio(catalog, uint8(track))
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		if typ == backend.FLAC {
			ctx.Header("Content-Type", "audio/flac")
		} else if typ == backend.MP3 {
			ctx.Header("Content-Type", "audio/mp3")
		}

		ctx.Stream(func(w io.Writer) bool {
			defer aud.Close()
			_, err := io.Copy(w, aud)
			return err != nil
		})
	})

	r.GET("/:catalog/cover", func(ctx *gin.Context) {
		catalog := ctx.Param("catalog")
		tok := ctx.GetHeader("Authorization")
		if !token.CheckCoverPerms(tok, catalog) {
			ctx.Status(http.StatusForbidden)
			return
		}
		cov, err := be.GetCover(catalog)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Stream(func(w io.Writer) bool {
			defer cov.Close()
			_, err := io.Copy(w, cov)
			return err != nil
		})
	})
}
