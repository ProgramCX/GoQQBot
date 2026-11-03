// Example interactions is a manual QQ acceptance suite. It never sends messages before a user command/click.
// interactions 示例用于 QQ 人工验收；仅在用户发送命令或点击按钮后发送消息。
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
	path := flag.String("config", "config.yaml", "YAML with GROUP_AND_C2C_EVENT and INTERACTION intents")
	options := suiteOptions{}
	flag.StringVar(&options.TemplateID, "keyboard-template", "", "optional platform keyboard template ID")
	flag.StringVar(&options.MiniAppURL, "miniapp-url", "", "optional real mini-app jump URL")
	flag.StringVar(&options.RoleID, "role-id", "", "optional real guild role ID")
	flag.StringVar(&options.ChannelID, "channel-id", "", "optional channel ID within the current guild")
	flag.StringVar(&options.MenuFeature, "menu-feature", "", "optional configured native shortcut-menu feature ID")
	flag.Parse()
	bot, err := goqqrobot.NewBot(*path, nil, goqqrobot.WithLogger(slog.Default()))

	if err != nil {
		log.Fatal(err)
	}
	bot.OnMessage(message.GROUP_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText(ctx.Text())
	})
	newSuite(options).register(bot)
	log.Print("在 QQ 发送 /interact 打开验收目录；客户端效果需逐项点击核对。")
	if err := bot.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}

}
