package http

import (
	"github.com/SeraphJACK/go-annil/storage"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"regexp"
	"time"
)

type session struct {
	username string
	expire   time.Time
}

var sessions = make(map[string]session)

var usernameExp = regexp.MustCompile("^[0-9a-zA-Z_]{2,15}$")

func regUserEndpoints(r *gin.Engine) {
	// Cleanup expired sessions
	go func() {
		for {
			<-time.Tick(time.Hour)
			for k, v := range sessions {
				if time.Now().After(v.expire) {
					delete(sessions, k)
				}
			}
		}
	}()

	r.POST("/api/login", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		if storage.CheckPassword(username, password) {
			sid := uuid.NewV4().String()
			sessions[sid] = session{
				username: username,
				expire:   time.Now().Add(time.Hour),
			}
			ctx.SetCookie("sessionId", sid, 0, "/", "", true, true)
			ctx.Status(http.StatusOK)
		} else {
			ctx.Status(http.StatusUnauthorized)
		}
	})
	r.POST("/api/register", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		if storage.UserExists(username) || !usernameExp.MatchString(username) {
			ctx.Header("X-Status-Reason", "USERNAME_UNAVAILABLE")
			ctx.Status(http.StatusConflict)
		} else if len(password) < 5 {
			ctx.Header("X-Status-Reason", "PASSWORD_TOO_SHORT")
			ctx.Status(http.StatusForbidden)
		} else {
			err := storage.Register(ctx.PostForm("username"), ctx.PostForm("password"))
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
			} else {
				ctx.Status(http.StatusOK)
			}
		}
	})
	r.POST("/api/revoke", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			revoke := ctx.PostForm("username")
			if !storage.UserExists(revoke) {
				ctx.Status(http.StatusNotFound)
				return
			}
			if revoke == username || storage.IsAdmin(username) {
				err := storage.RevokeUser(revoke)
				if err != nil {
					ctx.Status(http.StatusInternalServerError)
					log.Printf("Failed to revoke %s: %v\n", username, err)
				} else {
					ctx.Status(http.StatusOK)
				}
			} else {
				ctx.Status(http.StatusForbidden)
			}
		}
	})
	r.POST("/api/grantAdmin", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			grant := ctx.PostForm("username")
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			if !storage.UserExists(grant) {
				ctx.Status(http.StatusNotFound)
				return
			}
			err := storage.SetAdmin(grant, true)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				log.Printf("Failed to grant admin for %s: %v\n", grant, err)
			} else {
				ctx.Status(http.StatusOK)
			}
		}
	})
	r.POST("/api/revokeAdmin", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			revoke := ctx.PostForm("username")
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			if !storage.UserExists(revoke) {
				ctx.Status(http.StatusNotFound)
				return
			}
			if revoke == username {
				ctx.Status(http.StatusForbidden)
				return
			}
			err := storage.SetAdmin(revoke, false)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				log.Printf("Failed to revoke admin for %s: %v\n", revoke, err)
			} else {
				ctx.Status(http.StatusOK)
			}
		}
	})
	r.POST("/api/allowShare", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			grant := ctx.PostForm("username")
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			if !storage.UserExists(grant) {
				ctx.Status(http.StatusNotFound)
				return
			}
			err := storage.SetAllowShare(grant, true)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				log.Printf("Failed to grant share for %s: %v\n", grant, err)
			} else {
				ctx.Status(http.StatusOK)
			}
		}
	})
	r.POST("/api/disallowShare", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			revoke := ctx.PostForm("username")
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			if !storage.UserExists(revoke) {
				ctx.Status(http.StatusNotFound)
				return
			}
			err := storage.SetAllowShare(revoke, false)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				log.Printf("Failed to revoke share for %s: %v\n", revoke, err)
			} else {
				ctx.Status(http.StatusOK)
			}
		}
	})
	r.POST("/api/generateToken", func(ctx *gin.Context) {
		// TODO
	})
}

func authorize(ctx *gin.Context, u *string) bool {
	s, ok := sessions[ctx.GetHeader("Authorization")]
	if ok {
		if time.Now().After(s.expire) {
			ctx.Status(http.StatusUnauthorized)
			return false
		} else {
			*u = s.username
			return true
		}
	} else {
		ctx.Status(http.StatusUnauthorized)
		return false
	}
}
