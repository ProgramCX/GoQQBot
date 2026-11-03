// Example menu replies to group/C2C messages using the same high-level API.
// menu 示例使用同一套高级 API 回复群聊和单聊消息。
package main

import (
	"flag"
	"log"
	"log/slog"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func main() {
	path := flag.String("config", "config.yaml", "bot YAML configuration")
	flag.Parse()
	bot, err := goqqrobot.NewBot(*path, nil, goqqrobot.WithLogger(slog.Default()))
	if err != nil {
		log.Fatal(err)
	}
	menu := func(ctx *types.Context) error {
		return ctx.ReplyMarkdown("## 快捷菜单\n选择下方按钮，或发送 `/help`。",
			button.Command("帮助", "/help").ID("help").Primary().AutoSend(),
			button.Link("开发文档", "https://bot.q.qq.com/wiki/develop/api-v2/"),
		)
	}
	for _, event := range []message.EventId{message.GROUP_AT_MESSAGE_CREATE, message.C2C_MESSAGE_CREATE} {
		bot.OnCommand("menu", event, menu)
		bot.OnCommand("help", event, func(ctx *types.Context) error { return ctx.ReplyText("发送 /menu 打开菜单。") })
	}
	if err := bot.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}
