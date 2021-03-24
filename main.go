package main

import (
	"github.com/SeraphJACK/go-annil/config"
	"github.com/SeraphJACK/go-annil/http"
	"os"
)

func main() {
	err := config.Init()
	if err != nil {
		os.Exit(1)
	}
	_ = http.Init()
}
