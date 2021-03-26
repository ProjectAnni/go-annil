package http

import (
	"github.com/SeraphJACK/go-annil/storage"
	"github.com/SeraphJACK/go-annil/token"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"regexp"
	"strconv"
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
			ctx.SetCookie("sessionId", sid, 0, "/", "", false, true)
			ctx.Status(http.StatusOK)
		} else {
			ctx.Status(http.StatusUnauthorized)
		}
	})
	r.POST("/api/register", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		code := ctx.PostForm("inviteCode")
		if storage.UserExists(username) || !usernameExp.MatchString(username) {
			ctx.Header("X-Status-Reason", "USERNAME_UNAVAILABLE")
			ctx.Status(http.StatusConflict)
		} else if len(password) < 5 {
			ctx.Header("X-Status-Reason", "PASSWORD_TOO_SHORT")
			ctx.Status(http.StatusForbidden)
		} else {
			if !storage.ShrinkInviteCode(code) {
				ctx.Header("X-Status-Reason", "INVALID_INVITE_CODE")
				ctx.Status(http.StatusForbidden)
				return
			}
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
	r.POST("/api/changePassword", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			oldPass := ctx.PostForm("oldPassword")
			newPass := ctx.PostForm("newPassword")
			if !storage.CheckPassword(username, oldPass) {
				ctx.Header("X-Status-Reason", "WRONG_OLD_PASSWORD")
				ctx.Status(http.StatusForbidden)
				return
			}
			if len(newPass) < 5 {
				ctx.Header("X-Status-Reason", "PASSWORD_TOO_SHORT")
				ctx.Status(http.StatusForbidden)
				return
			}
			err := storage.ChangePassword(username, newPass)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				log.Printf("Failed to change password for %s: %v\n", username, err)
			} else {
				ctx.Status(http.StatusOK)
			}
		}
	})
	r.POST("/api/generateToken", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			tok, err := token.GenerateUserToken(username)
			if err != nil {
				log.Printf("Failed to generate user token for %s: %v\n", username, err)
				ctx.Status(http.StatusInternalServerError)
			} else {
				ctx.String(http.StatusOK, tok)
			}
		}
	})
	r.POST("/api/listUsers", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			ctx.JSON(http.StatusOK, storage.ListUsers())
		}
	})
	r.POST("/api/current", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			ctx.String(http.StatusOK, username)
		}
	})

	r.POST("/api/createInviteCode", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			limit, err := strconv.Atoi(ctx.PostForm("limit"))
			if err != nil || limit < -1 || limit == 0 {
				ctx.Status(http.StatusBadRequest)
				return
			}
			code, err := storage.NewInviteCode(limit)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				return
			}
			ctx.String(http.StatusOK, code)
		}
	})
	r.POST("/api/listInviteCodes", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			ctx.JSON(http.StatusOK, storage.ListInviteCodes())
		}
	})
	r.POST("/api/revokeInviteCode", func(ctx *gin.Context) {
		username := ""
		if authorize(ctx, &username) {
			if !storage.IsAdmin(username) {
				ctx.Status(http.StatusForbidden)
				return
			}
			code := ctx.PostForm("code")
			err := storage.RevokeInviteCode(code)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
			} else {
				ctx.Status(http.StatusOK)
			}
		}
	})
}

func authorize(ctx *gin.Context, u *string) bool {
	sid, err := ctx.Cookie("sessionId")
	if err != nil {
		ctx.Status(http.StatusUnauthorized)
		return false
	}
	s, ok := sessions[sid]
	if ok {
		if time.Now().After(s.expire) {
			ctx.Status(http.StatusUnauthorized)
			return false
		} else {
			if !storage.UserExists(s.username) {
				ctx.Status(http.StatusUnauthorized)
				return false
			}
			*u = s.username
			return true
		}
	} else {
		ctx.Status(http.StatusUnauthorized)
		return false
	}
}
