package http

import (
	"github.com/SeraphJACK/go-annil/backend"
	"github.com/SeraphJACK/go-annil/config"
	"github.com/gin-gonic/gin"
)

var r = gin.Default()

// TODO
var be backend.Backend

func Init() error {
	// TODO
	b, err := backend.NewFileBackend(config.Cfg.RepoRoot)
	if err != nil {
		return err
	}
	be = b

	regAnniEndpoints(r)
	regUserEndpoints(r)

	return r.Run(config.Cfg.Listen)
}
