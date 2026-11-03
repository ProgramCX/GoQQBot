# 单聊消息

QQ 用户直接与机器人对话时，消息事件为 `C2C_MESSAGE_CREATE`。它与频道私信 `DIRECT_MESSAGE_CREATE` 是两个场景，用户标识和发送路径不能混用。

## 先做什么

1. 在平台开通所需能力，并配置事件订阅；WebSocket 的普通单聊消息使用 `GROUP_AND_C2C_EVENT`。
2. [接收单聊消息](receive.md)，从 `ctx.Message()` 读取用户、正文和附件。
3. [发送单聊消息](send.md)，直接使用 `ctx.ReplyText` 或组合消息构造器。

## 最小处理函数

以下函数可以传给 `bot.OnMessage(message.C2C_MESSAGE_CREATE, HandlePrivate)`；导入及注册方式见[接收单聊消息](receive.md)。

```go
package guide

import "github.com/ProgramCX/GoQQBot/shared/types"

func HandlePrivate(ctx *types.Context) error {
	return ctx.ReplyText("已收到：" + ctx.Text())
}
```

## 场景限制

`user_openid` 是当前机器人在单聊场景得到的标识，不是普通 QQ 号码，也不能用群聊的 `member_openid` 替代。回复来源的有效期及消息次数由 QQ 检查；主动消息和唤醒消息还受资格及配额限制。

输入状态和流式消息仅适用于单聊。普通图片、视频及语音需要先上传再发送 `file_info`；Markdown 图片素材是另一套机制。

## 另请参阅

[接收单聊消息](receive.md) · [发送单聊消息](send.md) · [单聊流式消息](../streaming.md) · [官方发送单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)
