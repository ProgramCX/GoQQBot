package main

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/ProgramCX/GoQQBot/internal/adapter"
	"github.com/ProgramCX/GoQQBot/internal/foundation/config"
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
)

func main() {
	_, currentFile, _, _ := runtime.Caller(0)
	configPath := filepath.Join(filepath.Dir(currentFile), "config.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}
	va := adapter.NewValidator(cfg.Bot.AppSecret)
	bot, err := goqqrobot.New(*cfg, *va)
	if err != nil {
		return
	}
	if err := bot.StartWithGracefulShutdown(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
		return
	}
}
