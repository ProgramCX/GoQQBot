// Example media demonstrates C2C/group receiving, uploads, cards, quotes and recalling bot replies.
// media 示例演示单聊与群聊接收、上传、卡片、引用及撤回机器人回复。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type fixture struct {
	name, label, url, path string
	kind                   message.FileType
}
type options struct {
	fixtures             []fixture
	cardURL, cardPicture string
}
type mediaCase struct {
	command, label string
	handle         func(*types.Context) error
}

func main() {
	config := flag.String("config", "config.yaml", "bot configuration")
	debugEvents := flag.Bool("debug-events", false, "log event routing without message content; enables /media-ping")
	opts := options{}
	for i, name := range []string{"image", "video", "voice", "file"} {
		opts.fixtures = append(opts.fixtures, fixture{name: name, label: []string{"图片", "视频", "语音", "文件"}[i], kind: message.FileType(i + 1)})
	}
	for i := range opts.fixtures {
		f := &opts.fixtures[i]
		flag.StringVar(&f.url, f.name+"-url", "", "public media URL for explicit URL-upload tests")
		flag.StringVar(&f.path, f.name+"-file", "", "local file for explicit multipart-upload tests")
	}
	flag.StringVar(&opts.cardURL, "card-url", "", "HTTPS destination for the card test")
	flag.StringVar(&opts.cardPicture, "card-image", "", "HTTPS image for the card test")
	flag.Parse()
	bot, err := goqqrobot.NewBot(*config, nil, goqqrobot.WithLogger(slog.Default()))
	if err != nil {
		log.Fatal(err)
	}
	cases := mediaCases(opts)
	bot.OnCommand("/ping", message.GROUP_AT_MESSAGE_CREATE, func(ctx *types.Context) error {
		return ctx.ReplyText(ctx.Text())
	})
	for _, event := range []message.EventId{message.C2C_MESSAGE_CREATE, message.GROUP_AT_MESSAGE_CREATE} {
		bot.OnCommand("media", event, traceMediaHandler(*debugEvents, "media", func(c *types.Context) error {
			return c.Reply(mediaMenu(cases))
		}))
		for _, item := range cases {
			bot.OnCommand(item.command, event, traceMediaHandler(*debugEvents, item.command, item.handle))
		}
		bot.OnMessage(event, traceMediaHandler(*debugEvents, "message", inspectMessage))
		if *debugEvents {
			bot.OnCommand("media-ping", event, traceMediaHandler(true, "media-ping", func(c *types.Context) error {
				return c.ReplyText("已收到消息，纯文本回复正常。")
			}))
		}
	}
	if err := bot.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}

func traceMediaHandler(enabled bool, route string, handler func(*types.Context) error) goqqrobot.Handler {
	return func(c *types.Context) error {
		if enabled {
			log.Printf("media receive: event=%s route=%s command=%t text_bytes=%d", c.Payload.T, route, c.Command != "", len(c.Text()))
		}
		err := handler(c)
		if enabled {
			log.Printf("media complete: event=%s route=%s success=%t", c.Payload.T, route, err == nil)
			var apiError *api.APIError
			if errors.As(err, &apiError) && apiError.DiagnosticMessage() != "" {
				log.Printf("media API diagnostic: code=%d message=%q", apiError.Code, apiError.DiagnosticMessage())
			}
		}
		return err
	}
}

func mediaMenu(cases []mediaCase) message.Outgoing {
	buttons := make([]button.Button, 0, len(cases))
	for _, item := range cases {
		buttons = append(buttons, button.Command(item.label, "/"+item.command).ID(item.command).AutoSend())
	}
	// Fourteen cases require three columns to stay within QQ's five-row limit.
	// 十四个用例需使用三列，才能满足 QQ 最多五行的限制。
	return message.Markdown("## 消息与富媒体验收\n点击对应按钮执行；未配置素材的项目会说明缺少哪个启动参数。撤回测试只撤回刚发送的机器人回复。").Columns(3).Buttons(buttons...)
}

func mediaCases(opts options) []mediaCase {
	var cases []mediaCase
	for _, f := range opts.fixtures {
		for _, local := range []bool{false, true} {
			f, local := f, local
			suffix := "url"
			label := " URL"
			if local {
				suffix = "local"
				label = " 分片"
			}
			cases = append(cases, mediaCase{"media-" + f.name + "-" + suffix, f.label + label, func(c *types.Context) error {
				ctx, cancel := context.WithTimeout(c.RequestContext(), 2*time.Minute)
				defer cancel()
				var uploaded *message.RichMediaResponse
				var err error
				if local {
					if f.path == "" {
						return c.ReplyText("请配置 -" + f.name + "-file 后再测试。")
					}
					file, openErr := os.Open(f.path)
					if openErr != nil {
						return openErr
					}
					defer file.Close()
					info, statErr := file.Stat()
					if statErr != nil {
						return statErr
					}
					uploaded, err = c.UploadFile(ctx, api.FileUpload{Reader: file, Size: info.Size(), Name: filepath.Base(f.path), Type: f.kind})
				} else {
					if f.url == "" {
						return c.ReplyText("请配置 -" + f.name + "-url 后再测试。")
					}
					uploaded, err = c.UploadMedia(ctx, message.RichMediaRequest{FileType: f.kind, URL: f.url})
				}
				if err != nil {
					return fmt.Errorf("upload %s media: %w", f.name, err)
				}
				if err := c.ReplyContext(ctx, message.Media(uploaded.FileInfo)); err != nil {
					return fmt.Errorf("send uploaded %s media: %w", f.name, err)
				}
				return nil
			}})
		}
	}
	cases = append(cases,
		mediaCase{"media-card", "图文卡片", func(c *types.Context) error {
			if opts.cardURL == "" || opts.cardPicture == "" {
				return c.ReplyText("请配置 -card-url 和 -card-image。")
			}
			return c.Reply(message.Card(message.ImageTextCard{Title: "SDK 图文卡片", Description: "同一个处理函数支持群聊和单聊。", URL: opts.cardURL, PictureURL: opts.cardPicture}))
		}},
		mediaCase{"media-recall", "撤回机器人回复", func(c *types.Context) error {
			result, err := c.SendReply(c.RequestContext(), message.Text("这条机器人回复将被立即撤回。"))
			if err != nil {
				return err
			}
			if err := c.Recall(result.ID); err != nil {
				return err
			}
			return c.ReplyText("撤回 API 已成功返回，请检查上一条机器人回复已消失。")
		}},
		mediaCase{"media-quote", "引用当前消息", func(c *types.Context) error {
			m, err := c.Message()
			if err != nil {
				return err
			}
			index := m.MessageScene.MessageIndex()
			if index == "" {
				return c.ReplyText("本次事件没有提供 msg_idx，无法构造显示引用。")
			}
			return c.Reply(message.Markdown("引用当前消息。").Reference(index))
		}},
		mediaCase{"media-typing", "单聊输入状态", func(c *types.Context) error {
			target, err := c.Target()
			if err != nil {
				return err
			}
			if target.Scene != message.SceneC2C {
				return c.ReplyText("QQ 的输入中状态仅支持单聊，请到单聊测试。")
			}
			return c.Reply(message.Typing(10))
		}},
		mediaCase{"media-stream", "单聊流式消息", func(c *types.Context) error {
			target, err := c.Target()
			if err != nil {
				return err
			}
			if target.Scene != message.SceneC2C {
				return c.ReplyText("QQ 的流式消息仅支持单聊，请到单聊测试。")
			}
			first, err := c.SendStreamChunk(c.RequestContext(), message.StreamChunk{InputState: 1, ContentType: "text", Content: "第一段。"})
			if err != nil {
				return err
			}
			_, err = c.SendStreamChunk(c.RequestContext(), message.StreamChunk{Index: 1, InputState: 10, ContentType: "text", Content: "第二段，生成结束。", StreamMessageID: first.ID})
			return err
		}},
		mediaCase{"media-received", "读取消息结构", inspectMessage},
	)
	return cases
}

func inspectMessage(c *types.Context) error {
	m, err := c.Message()
	if err != nil {
		return err
	}
	if m.Author.Bot {
		return nil
	}
	cardType := "无"
	if m.ARK != nil {
		cardType = m.ARK.Type
	}
	// Deliberately summarize shape only: Raw and MessageScene may contain authorization tokens.
	// 仅汇总数据结构：Raw 和 MessageScene 可能包含授权令牌。
	return c.ReplyText(fmt.Sprintf("收到消息：type=%d，附件=%d，卡片类型=%s，顶层消息元素=%d，存在引用索引=%t。发送 /media 打开测试菜单。", m.MessageType, len(m.Attachments), cardType, len(m.MsgElements), m.MessageScene.MessageIndex() != ""))
}
