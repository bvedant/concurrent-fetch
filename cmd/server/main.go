package main

import (
	"log"

	"github.com/bvedant/concurrent-fetch/internal/api"
	"github.com/bvedant/concurrent-fetch/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app := api.NewApp(cfg)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
