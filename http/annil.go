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
	"strings"
	"time"
)

type CreateSharePayload struct {
	Audios map[string][]int `json:"audios"`
	Expire uint             `json:"expire"`
}

func regAnniEndpoints(r *gin.Engine) {
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

	r.GET("/albums", func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.JSON(http.StatusOK, be.ListCatalogs())
	})

	r.GET("/:catalog/cover", func(ctx *gin.Context) {
		catalog := ctx.Param("catalog")
		tok := ctx.GetHeader("Authorization")
		check := token.CheckCoverPerms(tok, catalog)
		if check != 0 {
			if check == 1 {
				ctx.Status(http.StatusForbidden)
			} else {
				ctx.Status(http.StatusUnauthorized)
			}
			return
		}
		cov, err := be.GetCover(catalog)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}

		ctx.Header("Access-Control-Allow-Origin", "*")

		ctx.Stream(func(w io.Writer) bool {
			defer cov.Close()
			_, err := io.Copy(w, cov)
			return err != nil
		})
	})

	r.GET("/:catalog/:track", func(ctx *gin.Context) {
		catalog := ctx.Param("catalog")
		trackStr := ctx.Param("track")
		track, err := strconv.Atoi(trackStr)
		if err != nil || track < 0 || track > 255 {
			ctx.Status(http.StatusBadRequest)
			return
		}
		tok := ctx.GetHeader("Authorization")
		check := token.CheckAudioPerms(tok, catalog, track)
		if check != 0 {
			if check == 1 {
				ctx.Status(http.StatusForbidden)
			} else {
				ctx.Status(http.StatusUnauthorized)
			}
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

		ctx.Header("Access-Control-Allow-Origin", "*")

		ctx.Stream(func(w io.Writer) bool {
			defer aud.Close()
			_, err := io.Copy(w, aud)
			return err != nil
		})
	})

	r.OPTIONS("/share", func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.RequestURI, "/api") {
			ctx.Header("Access-Control-Allow-Origin", "*")
		}
		ctx.Status(http.StatusOK)
	})
	// r.GET("/:catalog/:track", )
	// r.GET("/:catalog/cover")
}
