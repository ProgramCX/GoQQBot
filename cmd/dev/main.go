package main

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/pkg/api/c2c"
	"github.com/ProgramCX/GoQQBot/pkg/api/group"
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func main() {
	_, currentFile, _, _ := runtime.Caller(0)
	configPath := filepath.Join(filepath.Dir(currentFile), "config.yaml")

	bot, err := goqqrobot.NewBot(configPath)
	if err != nil {
		logger.ERROR("Error in creating new bot instance: %v", err)
	}

	bot.OnCommand("/v", message.GROUP_AT_MESSAGE_CREATE, func(ctx *types.Context) error {
		logger.INFO("收到 /v 命令")
		logger.DEBUG("ctx: %+v", ctx)
		err := group.ReplyPlainText(ctx, "你好"+ctx.Args[0], configPath)
		if err != nil {
			logger.ERROR("ReplyPlainText failed: %v", err)
		}
		return nil
	})

	bot.OnCommand("/v", message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
		logger.INFO("收到 /v 命令")
		err := c2c.ReplyPlainText(ctx, "你好"+ctx.Args[0], configPath)
		if err != nil {
			logger.ERROR("ReplyPlainText failed: %v", err)
		}
		return nil
	})

	bot.OnEvent(message.FRIEND_DEL, func(ctx *types.Context) error {
		logger.INFO("收到好友添加事件")
		return nil
	})
	if err := bot.StartWithGracefulShutdown(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
		return
	}
}
