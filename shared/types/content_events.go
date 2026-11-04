package types

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/channel"
)

// ForumEvent 解析当前论坛事件。开放论坛事件只提供标识，详情字段可能为空。
func (c *Context) ForumEvent() (*channel.ForumEvent, error) {
	return decodeEvent[channel.ForumEvent](c,
		message.FORUM_THREAD_CREATE, message.FORUM_THREAD_UPDATE, message.FORUM_THREAD_DELETE,
		message.FORUM_POST_CREATE, message.FORUM_POST_DELETE,
		message.FORUM_REPLY_CREATE, message.FORUM_REPLY_DELETE, message.FORUM_PUBLISH_AUDIT_RESULT,
		message.OPEN_FORUM_THREAD_CREATE, message.OPEN_FORUM_THREAD_UPDATE, message.OPEN_FORUM_THREAD_DELETE,
		message.OPEN_FORUM_POST_CREATE, message.OPEN_FORUM_POST_DELETE,
		message.OPEN_FORUM_REPLY_CREATE, message.OPEN_FORUM_REPLY_DELETE,
	)
}

// AudioAction 解析当前音频播放或机器人上下麦事件。
func (c *Context) AudioAction() (*channel.AudioAction, error) {
	return decodeEvent[channel.AudioAction](c,
		message.AUDIO_START, message.AUDIO_FINISH, message.AUDIO_ON_MIC, message.AUDIO_OFF_MIC,
	)
}

// AudioLiveMember 解析当前成员进入或退出音视频或直播子频道的事件。
func (c *Context) AudioLiveMember() (*channel.AudioLiveMember, error) {
	return decodeEvent[channel.AudioLiveMember](c,
		message.AUDIO_OR_LIVE_CHANNEL_MEMBER_ENTER, message.AUDIO_OR_LIVE_CHANNEL_MEMBER_EXIT,
	)
}
