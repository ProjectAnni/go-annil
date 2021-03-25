package main

import (
	"fmt"
	"github.com/SeraphJACK/go-annil/config"
	"github.com/SeraphJACK/go-annil/http"
	"github.com/SeraphJACK/go-annil/storage"
	"os"
)

func main() {
	err := config.Init()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read config: %v", err)
		os.Exit(1)
	}
	err = storage.Init()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize database: %v", err)
		os.Exit(1)
	}
	err = http.Init()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to start http server: %v", err)
		os.Exit(1)
	}
}
