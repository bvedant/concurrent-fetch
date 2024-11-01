package api

import (
	"github.com/bvedant/concurrent-fetch/internal/config"
)

type App struct {
	config *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Start() error {
	// Add your server startup logic here
	return nil
}
