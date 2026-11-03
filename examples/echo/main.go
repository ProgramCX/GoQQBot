// Example echo loads the selected event transport from a YAML file.
// echo 示例从 YAML 文件加载所选事件传输方式。
package main

import (
	"flag"
	"log"
	"log/slog"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func main() {
	path := flag.String("config", "config.yaml", "bot YAML configuration")
	flag.Parse()
	bot, err := goqqrobot.NewBot(*path, nil, goqqrobot.WithLogger(slog.Default()))
	if err != nil {
		log.Fatal(err)
	}
	bot.OnCommand("", message.GROUP_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText("Hello!")
	})
	bot.OnCommand("hello", message.C2C_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText("Hello!")
	})

	bot.OnEvent(message.GROUP_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText("全量消息")
	})
	if err := bot.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}
