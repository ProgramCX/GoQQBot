package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ProgramCX/GoQQBot/pkg/api"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/pkg/message/text"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

type registrar interface {
	OnCommand(string, message.EventId, goqqrobot.Handler)
	OnAnyMessage(goqqrobot.Handler)
	OnButton(string, goqqrobot.Handler)
	OnInteraction(interaction.InteractionType, goqqrobot.Handler)
	OnMenu(string, goqqrobot.Handler)
}
type suiteOptions struct{ TemplateID, MiniAppURL, RoleID, ChannelID, MenuFeature string }
type testCase struct {
	id, label, page string
	run             goqqrobot.Handler
}
type suite struct {
	options suiteOptions
	cases   []testCase
	mu      sync.Mutex
	counts  map[interaction.InteractionType]int
}

func newSuite(options suiteOptions) *suite {
	s := &suite{options: options, counts: make(map[interaction.InteractionType]int)}
	add := func(id, label, page string, run goqqrobot.Handler) {
		s.cases = append(s.cases, testCase{id, label, page, run})
	}
	command := func(id string, b button.Button) goqqrobot.Handler {
		return func(c *types.Context) error {
			return c.ReplyMarkdown("## "+id+"\n点击下方按钮，按提示发送后将收到 `/itest` 的处理结果。", b, s.back())
		}
	}
	add("callback", "普通回调", "basic", func(c *types.Context) error { return s.result(c, "普通回调已收到并 ACK") })
	add("prefill", "命令预填", "basic", command("命令预填", button.Command("预填命令", "/itest prefill 参数").ID("cmd-prefill")))
	add("auto", "自动发送", "basic", command("自动发送：单聊自动；群聊/频道预填", button.Command("发送命令", "/itest auto").AutoSend().ID("cmd-auto")))
	add("quote", "引用命令", "basic", command("引用命令", button.Command("引用发送", "/itest quote").Reply().AutoSend().ID("cmd-quote")))
	add("link", "网页跳转", "basic", func(c *types.Context) error {
		return c.ReplyMarkdown("网页跳转是客户端动作，不会产生机器人回调。打开网页后返回，点击核对按钮。", button.Link("官方文档", "https://bot.q.qq.com/wiki/develop/api-v2/").ID("docs"), s.confirmUI(), s.back())
	})
	add("visited", "点击后文字", "basic", func(c *types.Context) error {
		return c.ReplyMarkdown("核对点击后按钮文字，再核对处理结果。", button.Callback("点击测试", "visited").ID("visited-result").Visited("已点击"), s.back())
	})
	add("styles", "四种样式", "basic", func(c *types.Context) error {
		return c.ReplyMarkdown("分别点击四种样式按钮，每个都有处理结果。", button.Callback("灰色线框", "0").ID("style-0").Style(0), button.Callback("蓝色线框", "1").ID("style-1").Style(1), button.Callback("红色文字", "3").ID("style-3").Style(3), button.Callback("蓝底白字", "4").ID("style-4").Style(4), s.back())
	})
	add("old-client", "低版本提示", "basic", command("低版本客户端核对提示，高版本正常发送命令", button.Command("客户端提示", "/itest oldclient").UnsupportedTips("请手动输入 /itest oldclient")))
	add("text-buttons", "纯文本按钮", "basic", func(c *types.Context) error {
		return c.Reply(message.Text("纯文本消息携带键盘；核对当前客户端是否展示。").Buttons(button.Callback("文本回调", "text").ID("text-result"), s.back()))
	})
	add("strict", "严格模式", "advanced", func(c *types.Context) error {
		draft := message.Markdown("严格模式：只允许支持的场景保留自动发送。").Strict().Buttons(button.Command("严格发送", "/itest strict").AutoSend(), s.back())
		scene := message.SceneGroup
		if c2cScene(c) {
			scene = message.SceneC2C
		} else if channelScene(c) {
			scene = message.SceneChannel
		}
		if _, err := draft.Build(scene); err != nil {
			return s.result(c, "本地严格校验按预期拒绝："+err.Error())
		}
		return c.Reply(draft)
	})
	add("confirm", "二次确认", "advanced", func(c *types.Context) error {
		return c.ReplyMarkdown("先取消：不应收到回调；再次点击并确认：应收到处理结果。", button.Callback("确认操作", "confirm").ID("confirm-result").Confirmation(interaction.ButtonModal{Content: "是否提交这次测试？", ConfirmText: "提交", CancelText: "取消"}), s.back())
	})
	add("exclusive", "互斥分组", "advanced", func(c *types.Context) error {
		return c.ReplyMarkdown("选择 A 或 B 后，核对同组另一个按钮是否变灰。", button.Callback("选择 A", "A").ID("choice-a").Group("choice"), button.Callback("选择 B", "B").ID("choice-b").Group("choice"), s.back())
	})
	add("picker", "单聊选图", "advanced", command("选图：支持的单聊客户端打开图片选择器；其他场景预填", button.Command("选择图片", "/itest picker").PickImage().ID("cmd-picker")))
	add("chooser", "选择子频道", "advanced", command("频道选择器：频道保留，其他场景回退为预填", button.Command("选择频道", "/itest chooser").ChooseChannel().ID("cmd-chooser")))
	add("legacy", "旧点击次数", "advanced", func(c *types.Context) error {
		return c.ReplyMarkdown("click_limit 已被平台弃用；每次实际回调都有处理，不将未限次视为 SDK 成功保证。", button.Callback("旧字段测试", "legacy").ID("legacy-result").ClickLimit(1), s.back())
	})
	add("template", "模板键盘", "advanced", func(c *types.Context) error {
		if s.options.TemplateID == "" {
			return s.result(c, "未配置 -keyboard-template，不能伪造平台模板测试")
		}
		return c.Reply(message.Markdown("模板按钮由平台配置；回调由通用互动处理器接收。指令请使用 /itest template。另发 /interact 可返回。").KeyboardTemplate(s.options.TemplateID))
	})
	add("miniapp", "小程序跳转", "advanced", func(c *types.Context) error {
		if s.options.MiniAppURL == "" {
			return s.result(c, "未配置 -miniapp-url；需真实小程序链接")
		}
		return c.ReplyMarkdown("小程序跳转不会产生机器人回调，返回后手动确认。", button.Link("打开小程序", s.options.MiniAppURL), s.confirmUI(), s.back())
	})
	add("all", "所有人权限", "permissions", func(c *types.Context) error {
		return c.ReplyMarkdown("任意测试成员点击应收到处理结果。", button.Callback("所有人点击", "all").ID("permission-all").All(), s.back())
	})
	add("self", "指定用户", "permissions", func(c *types.Context) error {
		id := actor(c)
		if id == "" {
			return s.result(c, "事件缺少当前用户 ID，无法构造指定用户测试")
		}
		return c.ReplyMarkdown("当前测试者可点；请另一位成员核对拒绝操作时不产生业务回调。", button.Callback("仅测试者", "self").ID("permission-self").Users(id), s.back())
	})
	add("admin", "管理员权限", "permissions", func(c *types.Context) error {
		return c.ReplyMarkdown("需管理员和普通成员分别核对允许/拒绝；拒绝操作不会产生业务回调。", button.Callback("管理员点击", "admin").ID("permission-admin").AdminOnly(), s.back())
	})
	add("role", "频道角色权限", "permissions", func(c *types.Context) error {
		if !channelScene(c) || s.options.RoleID == "" {
			return s.result(c, "仅频道可测；需配置当前频道真实 -role-id")
		}
		return c.ReplyMarkdown("请具备该角色及不具备该角色的成员分别核对。", button.Callback("指定角色", "role").ID("permission-role").Roles(s.options.RoleID), s.back())
	})
	add("mention", "提及自己", "text", func(c *types.Context) error {
		if c2cScene(c) {
			return s.result(c, "单聊没有 @成员 能力，请在群聊或频道测试")
		}
		id := actor(c)
		if id == "" {
			return s.result(c, "缺少当前用户标识")
		}
		return c.Reply(message.Markdown("## 提及测试\n").Append(text.AtUser(id)).Buttons(s.confirmUI(), s.back()))
	})
	for _, md := range []bool{false, true} {
		id, label := "everyone-text", "全体文本"
		if md {
			id, label = "everyone-md", "全体MD"
		}
		add(id, label, "text", func(c *types.Context) error {
			if c2cScene(c) {
				return s.result(c, "单聊不支持 @全体")
			}
			fragments := []text.Fragment{text.AtEveryone(), text.Plain(" SDK 交互验收，请核对显示。")}
			if md {
				return c.Reply(message.MarkdownParts(fragments...).Buttons(s.confirmUI(), s.back()))
			}
			return c.Reply(message.TextParts(fragments...))
		})
	}
	add("input", "文本预填", "text", func(c *types.Context) error {
		return c.Reply(message.Markdown("核对预填内容中的中文、空格、+、& 和引用。\n").Append(text.Input("/itest input 中文 + &").Label("填写测试命令").Reference()).Buttons(s.back()))
	})
	add("enter", "文本发送", "text", func(c *types.Context) error {
		return c.Reply(message.Markdown("单聊直接发送；群聊/频道预填，之后手动发送。\n").Append(text.Enter("/itest enter")).Buttons(s.back()))
	})
	add("channel", "频道跳转", "text", func(c *types.Context) error {
		if !channelScene(c) || s.options.ChannelID == "" {
			return s.result(c, "仅频道可测，需配置同一频道内真实 -channel-id")
		}
		return c.Reply(message.Markdown("点击跳转：").Append(text.Channel(s.options.ChannelID)).Buttons(s.confirmUI(), s.back()))
	})
	add("emoji", "频道系统表情", "text", func(c *types.Context) error {
		if !channelScene(c) {
			return s.result(c, "系统表情标签仅频道可测")
		}
		return c.Reply(message.Markdown("系统表情：").Append(text.Emoji("307"), text.Plain("；普通 Unicode：😀")).Buttons(s.confirmUI(), s.back()))
	})
	add("message-text", "普通文字", "messages", func(c *types.Context) error {
		return c.ReplyMarkdown("点击预填普通文字并发送，将进入 OnMessage。", button.Command("普通文字", "你好，普通消息").ID("cmd-message-text").AutoSend(), s.back())
	})
	add("message-file", "附件消息", "messages", func(c *types.Context) error {
		return c.ReplyMarkdown("可直接发送图片、语音或文件；也可在支持的单聊客户端点下方选图。", button.Command("发送附件", "附件测试 ").ID("cmd-message-file").PickImage(), s.back())
	})
	add("message-unknown", "未知命令", "messages", func(c *types.Context) error {
		return c.ReplyMarkdown("未注册的命令也会进入新的 OnMessage 兜底。", button.Command("未知命令", "/not_registered demo").ID("cmd-message-unknown").AutoSend(), s.back())
	})
	return s
}

func (s *suite) back() button.Button { return button.Callback("返回目录", "home").ID("page-home") }
func (s *suite) confirmUI() button.Button {
	return button.Callback("我已核对效果", "ui-confirm").ID("ui-confirm")
}
func (s *suite) result(c *types.Context, value string) error { return c.ReplyMarkdown(value, s.back()) }
func (s *suite) wrap(code interaction.InteractionCallbackCode, handler goqqrobot.Handler) goqqrobot.Handler {
	return func(c *types.Context) error {
		if err := c.AckInteraction(code); err != nil {
			return err
		}
		err := handler(c)
		var platform *api.APIError
		if errors.As(err, &platform) {
			return s.result(c, fmt.Sprintf("平台拒绝本项测试：HTTP %d，错误码 %d。请检查权限、配额或真实测试资源；本项未通过。", platform.StatusCode, platform.Code))
		}
		return err
	}
}
func (s *suite) register(r registrar) {
	r.OnAnyMessage(func(c *types.Context) error {
		received, err := c.Message()
		if err != nil {
			return err
		}
		if received.Author.Bot {
			return nil
		}
		if c.Command != "" {
			return c.ReplyText("未注册的命令已由 OnMessage 接收：/" + c.Command + "；发送 /interact 返回目录。")
		}
		if len(received.Attachments) > 0 {
			return c.ReplyText(fmt.Sprintf("OnMessage 已收到附件 %d 个，正文：%s", len(received.Attachments), c.Text()))
		}
		return c.ReplyText("OnMessage 已收到普通消息：" + c.Text())
	})
	for _, event := range []message.EventId{message.GROUP_AT_MESSAGE_CREATE, message.GROUP_MESSAGE_CREATE, message.C2C_MESSAGE_CREATE, message.AT_MESSAGE_CREATE, message.MESSAGE_CREATE, message.DIRECT_MESSAGE_CREATE} {
		r.OnCommand("interact", event, func(c *types.Context) error { return s.page(c, "home") })
		r.OnCommand("itest", event, func(c *types.Context) error {
			args := strings.Join(c.Args, " ")
			if len(c.Args) > 0 && c.Args[0] == "picker" {
				var data struct {
					Attachments []json.RawMessage `json:"attachments"`
				}
				if err := json.Unmarshal(c.Payload.D, &data); err != nil {
					return err
				}
				return s.result(c, fmt.Sprintf("已进入选图命令处理器；收到附件 %d 个。0 个时不算图片接收通过。", len(data.Attachments)))
			}
			return s.result(c, "已收到测试命令参数："+strconv.Quote(args))
		})
	}
	for _, page := range []string{"home", "basic", "advanced", "permissions", "acks", "text", "messages", "platform"} {
		r.OnButton("page-"+page, s.wrap(0, func(c *types.Context) error { return s.page(c, page) }))
	}
	for _, test := range s.cases {
		r.OnButton("case-"+test.id, s.wrap(0, test.run))
	}
	for _, id := range []string{"text-result", "visited-result", "confirm-result", "choice-a", "choice-b", "legacy-result", "permission-all", "permission-self", "permission-admin", "permission-role", "style-0", "style-1", "style-3", "style-4"} {
		r.OnButton(id, s.wrap(0, func(c *types.Context) error {
			return s.result(c, "按钮处理器已执行："+id+"；请同时核对客户端显示效果。 ")
		}))
	}
	r.OnButton("ui-confirm", s.wrap(0, func(c *types.Context) error {
		return s.result(c, "已记录你对客户端效果的手动确认；这不是后台自动检测结果。 ")
	}))
	for code := 0; code <= 5; code++ {
		r.OnButton("ack-"+strconv.Itoa(code), s.wrap(interaction.InteractionCallbackCode(code), func(c *types.Context) error {
			return s.result(c, fmt.Sprintf("已发送互动响应 code=%d。请核对客户端反馈；非零码是刻意的错误场景测试。", code))
		}))
	}
	native := func(c *types.Context) error {
		event, err := c.Interaction()
		if err != nil {
			return err
		}
		if event.NeedsAcknowledgement() {
			if err := c.AckInteraction(0); err != nil {
				return err
			}
		}
		s.mu.Lock()
		s.counts[event.Kind()]++
		s.mu.Unlock()
		if event.NeedsAcknowledgement() {
			return s.result(c, fmt.Sprintf("已收到平台互动 type=%d，并发送 ACK。", event.Kind()))
		}
		return nil
	}
	for _, kind := range []interaction.InteractionType{11, 12, 13, 14, 15, 16, 18, 19, 20} {
		r.OnInteraction(kind, native)
	}
	if s.options.MenuFeature != "" {
		r.OnMenu(s.options.MenuFeature, native)
	}
}

func (s *suite) page(c *types.Context, page string) error {
	title := "## " + page + " 验收\n"
	var buttons []button.Button
	if page == "home" {
		title = "## SDK 交互验收目录\n每项都有演示和对应处理。客户端跳转、权限拒绝、确认取消需人工核对，不能以后台测试替代。"
		for _, p := range []struct{ id, label string }{{"basic", "基础按钮"}, {"advanced", "高级按钮"}, {"permissions", "按钮权限"}, {"acks", "六种响应码"}, {"text", "文本交互"}, {"platform", "平台原生事件"}, {"messages", "普通消息"}} {
			buttons = append(buttons, button.Callback(p.label, p.id).ID("page-"+p.id))
		}
	} else if page == "acks" {
		for code := 0; code <= 5; code++ {
			buttons = append(buttons, button.Callback("响应码 "+strconv.Itoa(code), strconv.Itoa(code)).ID("ack-"+strconv.Itoa(code)))
		}
	} else if page == "platform" {
		title += "快捷菜单、反馈、清空会话、故事、模型、授权由 QQ 原生 UI 触发，普通消息按钮无法伪造这些事件。\n本页只统计通用类型处理器收到的原生/模板事件，专用目录回调不计入。\n请在客户端触发，再点本页刷新查看计数；未开放的能力不能算已测通过。\n"
		s.mu.Lock()
		var kinds []int
		for kind := range s.counts {
			kinds = append(kinds, int(kind))
		}
		sort.Ints(kinds)
		for _, kind := range kinds {
			title += fmt.Sprintf("type=%d：%d 次\n", kind, s.counts[interaction.InteractionType(kind)])
		}
		s.mu.Unlock()
		buttons = append(buttons, button.Callback("刷新事件记录", "platform").ID("page-platform"))
	} else {
		for _, test := range s.cases {
			if test.page != page {
				continue
			}
			b := button.Callback(test.label, test.id).ID("case-" + test.id)
			if strings.HasPrefix(test.id, "everyone-") {
				b = b.AdminOnly().Confirm("每群每天仅一次，确定发送 @全体？")
			}
			buttons = append(buttons, b)
		}
	}
	if page != "home" {
		buttons = append(buttons, s.back())
	}
	return c.Reply(message.Markdown(title).Columns(2).Buttons(buttons...))
}

func c2cScene(c *types.Context) bool {
	return c.Payload.T == message.C2C_MESSAGE_CREATE || sceneName(c) == "c2c"
}
func channelScene(c *types.Context) bool {
	return c.Payload.T == message.AT_MESSAGE_CREATE || c.Payload.T == message.MESSAGE_CREATE || c.Payload.T == message.DIRECT_MESSAGE_CREATE || sceneName(c) == "guild"
}
func sceneName(c *types.Context) string {
	var d struct {
		Scene string `json:"scene"`
	}
	_ = json.Unmarshal(c.Payload.D, &d)
	return d.Scene
}
func actor(c *types.Context) string {
	var d struct {
		User   string `json:"user_openid"`
		Member string `json:"group_member_openid"`
		Author struct {
			ID     string `json:"id"`
			User   string `json:"user_openid"`
			Member string `json:"member_openid"`
		} `json:"author"`
		Data struct {
			Resolved struct {
				User string `json:"user_id"`
			} `json:"resolved"`
		} `json:"data"`
	}
	if json.Unmarshal(c.Payload.D, &d) != nil {
		return ""
	}
	for _, id := range []string{d.Member, d.User, d.Data.Resolved.User, d.Author.Member, d.Author.User, d.Author.ID} {
		if id != "" {
			return id
		}
	}
	return ""
}
