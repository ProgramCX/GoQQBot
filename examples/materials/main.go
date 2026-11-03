// Example materials demonstrates bot-owned Markdown image management and typed rich-message receiving.
// materials 示例演示机器人 Markdown 图片素材管理和类型化富文本接收。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type materialDemo struct {
	mu    sync.RWMutex
	image *message.ImageMaterial
	file  string
}

func (d *materialDemo) current() *message.ImageMaterial {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.image == nil {
		return nil
	}
	image := *d.image
	return &image
}

type demoCase struct {
	command, label string
	handler        goqqrobot.Handler
}

func (d *materialDemo) cases() []demoCase {
	return []demoCase{
		{"material-upload", "上传并展示", d.upload},
		{"material-reuse", "复用当前素材", d.reuse},
		{"material-list", "查询素材首页", d.list},
		{"material-delete", "删除当前素材", d.delete},
		{"material-received", "接收结构说明", func(c *types.Context) error {
			return c.ReplyText("直接发送图片、表情或引用消息，示例会通过 OnMessage 读取正文、图片附件、卡片和嵌套元素；不自动下载或上传接收到的内容。")
		}},
	}
}

func materialMenu(cases []demoCase) message.Outgoing {
	var buttons []button.Button
	for _, item := range cases {
		buttons = append(buttons, button.Command(item.label, "/"+item.command).ID(item.command).AutoSend())
	}
	return message.Markdown("## Markdown 图片素材\n上传创建可复用素材；复用不会重复上传。删除只作用于本次进程记住的当前素材，其他素材由业务管理。").Columns(2).Buttons(buttons...)
}

func (d *materialDemo) upload(c *types.Context) error {
	if d.file == "" {
		return c.ReplyText("请在启动时配置 -image-file 本地图片路径。")
	}
	ctx, cancel := context.WithTimeout(c.RequestContext(), 2*time.Minute)
	defer cancel()
	image, err := c.UploadMarkdownImageFile(ctx, d.file)
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.image = image
	d.mu.Unlock()
	return c.ReplyContext(ctx, image.Markdown("已上传的素材"))
}

func (d *materialDemo) reuse(c *types.Context) error {
	image := d.current()
	if image == nil {
		return c.ReplyText("请先上传一张素材。本示例只在内存中记住当前素材；业务系统可保存 image_id 和引用 URL。")
	}
	return c.Reply(image.Markdown("再次引用同一素材"))
}

func (d *materialDemo) list(c *types.Context) error {
	ctx, cancel := context.WithTimeout(c.RequestContext(), 30*time.Second)
	defer cancel()
	page, err := c.Client.ListMarkdownImages(ctx, "")
	if err != nil {
		return err
	}
	var text strings.Builder
	fmt.Fprintf(&text, "有效素材总数：%d；本页：%d\n", page.Total, len(page.Images))
	for _, image := range page.Images {
		fmt.Fprintf(&text, "ID=%s，名称=%s，过期时间（Unix秒）=%d\n", image.ID, image.FileName, image.ExpiresAt)
	}
	if page.NextCursor != "" {
		text.WriteString("还有下一页；业务代码将 NextCursor 传给 ListMarkdownImages 继续读取。")
	}
	return c.ReplyText(text.String())
}

func (d *materialDemo) delete(c *types.Context) error {
	image := d.current()
	if image == nil {
		return c.ReplyText("本次进程没有记住可删除的当前素材。")
	}
	ctx, cancel := context.WithTimeout(c.RequestContext(), 30*time.Second)
	defer cancel()
	if err := c.Client.DeleteMarkdownImage(ctx, image.ID); err != nil {
		return err
	}
	d.mu.Lock()
	if d.image != nil && d.image.ID == image.ID {
		d.image = nil
	}
	d.mu.Unlock()
	return c.ReplyText("当前素材已删除；此前发送的消息副本不受影响。")
}

func receiveRich(c *types.Context) error {
	m, err := c.Message()
	if err != nil {
		return err
	}
	if m.Author.Bot {
		return nil
	}
	kind := "无"
	if m.ARK != nil {
		kind = m.ARK.Type
	}
	return c.ReplyText(fmt.Sprintf("正文字符数=%d，图片附件=%d，全部附件（含嵌套）=%d，卡片类型=%s，顶层消息元素=%d。\nSDK 保留原始正文；表情包仅在平台下发图片附件时可读取图片 URL。", len([]rune(m.Content)), len(m.Images()), len(m.AllAttachments()), kind, len(m.MsgElements)))
}

func main() {
	config := flag.String("config", "config.yaml", "bot configuration")
	file := flag.String("image-file", "", "local png/jpeg/gif/webp image for explicit upload tests")
	flag.Parse()
	bot, err := goqqrobot.NewBot(*config, nil, goqqrobot.WithLogger(slog.Default()))
	if err != nil {
		log.Fatal(err)
	}
	demo := &materialDemo{file: *file}
	cases := demo.cases()
	for _, event := range []message.EventId{message.C2C_MESSAGE_CREATE, message.GROUP_AT_MESSAGE_CREATE} {
		bot.OnCommand("materials", event, func(c *types.Context) error { return c.Reply(materialMenu(cases)) })
		for _, item := range cases {
			bot.OnCommand(item.command, event, item.handler)
		}
		bot.OnMessage(event, receiveRich)
	}
	if err := bot.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}
