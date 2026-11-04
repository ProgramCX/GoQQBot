package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/config"
	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func readImage(path string) (*channel.LocalImage, error) {
	if path == "" {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// 这是验收工具自身的内存上限，平台另有上传要求。
	const maxImage = 10 << 20
	data, err := io.ReadAll(io.LimitReader(file, maxImage+1))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > maxImage {
		return nil, errors.New("测试图片应大于 0 且不超过 10 MiB")
	}
	return &channel.LocalImage{Filename: filepath.Base(path), Data: data}, nil
}

func listen(cfg config.Config, imagePath string, replies bool) error {
	image, err := readImage(imagePath)
	if err != nil {
		return err
	}
	bot, err := goqqrobot.NewWithConfig(cfg, nil, goqqrobot.WithErrorHandler(logEventError))
	if err != nil {
		return err
	}

	for _, event := range []message.EventId{message.GROUP_AT_MESSAGE_CREATE, message.GROUP_MESSAGE_CREATE} {
		bot.OnMessage(event, func(c *types.Context) error {
			m, err := c.Message()
			if err != nil {
				return err
			}
			log.Printf("%s group_openid=%q member_openid=%q", c.Payload.T, m.GroupOpenID, m.Author.MemberOpenID)
			return nil
		})
	}
	commands := map[string]goqqrobot.Handler{
		"sdk-ping":   func(c *types.Context) error { return c.ReplyText("频道消息收发正常。") },
		"sdk-recall": recallReply,
		"sdk-image":  func(c *types.Context) error { return replyImage(c, image) },
		"sdk-legacy": legacyReply,
	}
	if replies {
		for name := range commands {
			bot.OnCommand(name, message.C2C_MESSAGE_CREATE, func(c *types.Context) error {
				log.Print("收到 C2C_MESSAGE_CREATE，当前会话是 QQ 单聊。")
				return c.ReplyText("这里是普通 QQ 单聊。请从 QQ 频道进入机器人私信，再发送测试命令。")
			})
		}
	}
	for _, event := range []message.EventId{message.AT_MESSAGE_CREATE, message.MESSAGE_CREATE, message.DIRECT_MESSAGE_CREATE} {
		bot.OnMessage(event, func(c *types.Context) error {
			m, err := c.ChannelMessage()
			if err != nil {
				return err
			}
			log.Printf("%s guild_id=%q channel_id=%q user_id=%q message_id=%q attachments=%d",
				c.Payload.T, m.GuildID, m.ChannelID, m.Author.ID, m.ID, len(m.Attachments))
			// 某些频道事件的正文保留开头的提及标签。
			if replies && !m.Author.Bot && strings.HasPrefix(c.Text(), "<@") {
				if _, rest, found := strings.Cut(c.Text(), ">"); found {
					if command, found := strings.CutPrefix(strings.TrimSpace(rest), "/"); found {
						if handler := commands[command]; handler != nil {
							return handler(c)
						}
					}
				}
			}
			return nil
		})
		if replies {
			for name, handler := range commands {
				bot.OnCommand(name, event, handler)
			}
		}
	}

	observe(bot, (*types.Context).GroupMemberChange, message.GROUP_MEMBER_ADD, message.GROUP_MEMBER_REMOVE)
	observe(bot, (*types.Context).GuildChange, message.GUILD_CREATE, message.GUILD_UPDATE, message.GUILD_DELETE)
	observe(bot, (*types.Context).SubchannelChange, message.CHANNEL_CREATE, message.CHANNEL_UPDATE, message.CHANNEL_DELETE)
	observe(bot, (*types.Context).GuildMemberChange, message.GUILD_MEMBER_ADD, message.GUILD_MEMBER_UPDATE, message.GUILD_MEMBER_REMOVE)
	observe(bot, (*types.Context).ChannelMessageAudit, message.MESSAGE_AUDIT_PASS, message.MESSAGE_AUDIT_REJECT)
	observe(bot, (*types.Context).ChannelReaction, message.MESSAGE_REACTION_ADD, message.MESSAGE_REACTION_REMOVE)
	observe(bot, (*types.Context).AudioLiveMember, message.AUDIO_OR_LIVE_CHANNEL_MEMBER_ENTER, message.AUDIO_OR_LIVE_CHANNEL_MEMBER_EXIT)
	bot.OnEvent(message.GROUP_JOIN_REQUEST, func(c *types.Context) error {
		m, err := c.GroupJoinRequest()
		if err != nil {
			return err
		}
		log.Printf("%s group_openid=%q member_openid=%q join_request_id=%q auto_approved=%t",
			c.Payload.T, m.GroupOpenID, m.MemberOpenID, m.JoinRequestID, m.AutoApproved != nil)
		return nil
	})
	for _, event := range []message.EventId{message.FORUM_THREAD_CREATE, message.FORUM_THREAD_UPDATE, message.FORUM_THREAD_DELETE, message.FORUM_POST_CREATE, message.FORUM_POST_DELETE, message.FORUM_REPLY_CREATE, message.FORUM_REPLY_DELETE, message.FORUM_PUBLISH_AUDIT_RESULT, message.OPEN_FORUM_THREAD_CREATE, message.OPEN_FORUM_THREAD_UPDATE, message.OPEN_FORUM_THREAD_DELETE, message.OPEN_FORUM_POST_CREATE, message.OPEN_FORUM_POST_DELETE, message.OPEN_FORUM_REPLY_CREATE, message.OPEN_FORUM_REPLY_DELETE} {
		bot.OnEvent(event, func(c *types.Context) error {
			m, err := c.ForumEvent()
			if err != nil {
				return err
			}
			thread := m.ThreadID
			if m.Thread != nil {
				thread = m.Thread.ThreadID
			}
			if m.Post != nil {
				thread = m.Post.ThreadID
			}
			if m.Reply != nil {
				thread = m.Reply.ThreadID
			}
			log.Printf("%s guild_id=%q channel_id=%q thread_id=%q", c.Payload.T, m.GuildID, m.ChannelID, thread)
			return nil
		})
	}
	for _, event := range []message.EventId{message.AUDIO_START, message.AUDIO_FINISH, message.AUDIO_ON_MIC, message.AUDIO_OFF_MIC} {
		bot.OnEvent(event, func(c *types.Context) error {
			m, err := c.AudioAction()
			if err != nil {
				return err
			}
			log.Printf("%s guild_id=%q channel_id=%q", c.Payload.T, m.GuildID, m.ChannelID)
			return nil
		})
	}
	log.Printf("启动验收监听，模式=%s，QQ 测试命令启用=%t。按 Ctrl+C 停止。", cfg.Connection.Mode, replies)
	return bot.StartWithGracefulShutdown()
}

func logEventError(_ *types.Context, err error) {
	log.Printf("事件处理失败: %v", err)
	var panicErr *types.PanicError
	var apiErr *api.APIError
	var requestErr *url.Error
	var handlerErr *types.HandlerError
	switch {
	case errors.As(err, &panicErr):
		log.Printf("处理函数异常: %v", panicErr)
	case errors.As(err, &apiErr):
		logAPIError(apiErr)
	case errors.As(err, &requestErr):
		log.Printf("HTTP 请求失败: operation=%s timeout=%t cause_type=%T", requestErr.Op, requestErr.Timeout(), requestErr.Err)
	case errors.As(err, &handlerErr):
		log.Printf("处理函数原因: %v", handlerErr.Err)
	}
}

func logAPIError(err *api.APIError) {
	log.Printf("QQ API HTTP=%d code=%d trace=%q message=%q audit_id=%q",
		err.StatusCode, err.Code, err.TraceID, err.DiagnosticMessage(), err.MessageAuditID)
}

func observe[T any](bot *goqqrobot.Bot, decode func(*types.Context) (*T, error), events ...message.EventId) {
	for _, event := range events {
		bot.OnEvent(event, func(c *types.Context) error {
			value, err := decode(c)
			if err != nil {
				return err
			}
			data, err := json.Marshal(value)
			if err != nil {
				return err
			}
			log.Printf("%s %s", c.Payload.T, data)
			return nil
		})
	}
}

func recallReply(c *types.Context) error {
	result, err := c.SendReply(c.RequestContext(), message.Text("撤回测试，3 秒后尝试撤回这条回复。"))
	if err != nil {
		return err
	}
	if result.ID == "" {
		return errors.New("平台响应缺少消息 ID")
	}
	log.Printf("等待撤回 message_id=%q", result.ID)
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-c.RequestContext().Done():
		return c.RequestContext().Err()
	case <-timer.C:
	}
	err = c.Recall(result.ID)
	if err == nil {
		log.Printf("撤回成功 message_id=%q", result.ID)
		return c.ReplyText("撤回成功，请检查上一条测试回复是否已消失。")
	}
	var hint string
	switch {
	case errors.Is(err, api.ErrAPIAccessDenied):
		hint = "撤回未执行：当前机器人应用没有撤回接口权限。QQ 仅向私域机器人开放频道消息撤回。"
	case errors.Is(err, api.ErrOperationDenied):
		hint = "撤回未执行：机器人没有撤回这条消息的权限。"
	case errors.Is(err, api.ErrRecallWindowExpired):
		hint = "撤回未执行：已超过平台允许的撤回时限。"
	case errors.Is(err, api.ErrInvalidAPIParameters):
		hint = "撤回失败：QQ 拒绝了撤回参数，请核对会话和消息 ID。"
	case errors.Is(err, api.ErrInvalidMessageID):
		hint = "撤回失败：消息 ID 无效，请使用发送接口返回的消息 ID。"
	case errors.Is(err, api.ErrRecallMessageLookup):
		hint = "撤回失败：QQ 未能获取这条消息。请检查消息是否仍在，再决定是否重试。"
	case errors.Is(err, api.ErrRecallChannelLookup):
		hint = "撤回失败：QQ 未能获取子频道。请确认频道仍可访问，再决定是否重试。"
	case errors.Is(err, api.ErrRecallFailed):
		hint = "QQ 报告撤回失败。请先确认消息是否仍在，再决定是否重试。"
	default:
		var apiErr *api.APIError
		if !errors.As(err, &apiErr) {
			return fmt.Errorf("撤回频道测试消息: %w", err)
		}
		hint = fmt.Sprintf("QQ 拒绝了撤回请求（错误码 %d）。请查看终端详情。", apiErr.Code)
	}
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		logAPIError(apiErr)
	}
	if replyErr := c.ReplyText(hint); replyErr != nil {
		return fmt.Errorf("发送撤回结果提示: %w", errors.Join(replyErr, err))
	}
	return nil
}

func replyImage(c *types.Context, image *channel.LocalImage) error {
	if image == nil {
		return c.ReplyText("请用 -ImageFile 配置本地测试图片。")
	}
	m, err := c.ChannelMessage()
	if err != nil {
		return err
	}
	if m.ID == "" {
		return errors.New("频道消息缺少回复来源 ID")
	}
	client := channel.NewClient(c.Client)
	body := channelmsg.ChannelSendMessage{Content: "频道图片测试", MsgId: m.ID}
	var result *message.ReplyResult
	if c.Payload.T == message.DIRECT_MESSAGE_CREATE {
		result, err = client.SendDirectMessage(c.RequestContext(), m.GuildID, body, image)
	} else {
		result, err = client.SendChannelMessage(c.RequestContext(), m.ChannelID, body, image)
	}
	if err != nil {
		return fmt.Errorf("发送本地图片 (%s): %w", c.Payload.T, err)
	}
	log.Printf("图片回复 message_id=%q", result.ID)
	return nil
}

func legacyReply(c *types.Context) error {
	m, err := c.ChannelMessage()
	if err != nil {
		return err
	}
	if m.ID == "" {
		return errors.New("频道消息缺少回复来源 ID")
	}
	var result *message.SendMessageResponse
	if c.Payload.T == message.DIRECT_MESSAGE_CREATE {
		result, err = channel.SendDMPlainText(c, m.GuildID, "频道私信兼容入口测试。", m.ID)
	} else {
		result, err = channel.SendPlainText(c, m.ChannelID, "子频道兼容入口测试。", m.ID)
	}
	if err == nil {
		log.Printf("兼容入口 message_id=%q unix_timestamp=%d", result.ID, result.Timestamp)
	}
	return err
}
