package http

import (
	"fmt"
	"github.com/SeraphJACK/go-annil/backend"
	"github.com/SeraphJACK/go-annil/config"
	"github.com/gin-gonic/gin"
)

var r = gin.Default()

var be backend.Backend

func Init() error {
	backends := make([]backend.Backend, 0)
	for _, entry := range config.Cfg.Backends {
		switch entry.Type {
		case "file":
			{
				be, err := backend.NewFileBackend(entry.Path)
				if err != nil {
					return fmt.Errorf("failed to initialize backend: %v", err)
				}
				backends = append(backends, be)
			}
		case "relay":
			{
				be := backend.NewRelay(entry.Path, entry.Auth)
				backends = append(backends, be)
			}
		default:
			return fmt.Errorf("unknwon backend type: %s", entry.Type)
		}
	}

	be = backend.NewMultiplexer(backends)

	regAnniEndpoints(r)
	regUserEndpoints(r)

	// Static files
	r.NoRoute(serveFrontend)

	return r.Run(config.Cfg.Listen)
}
