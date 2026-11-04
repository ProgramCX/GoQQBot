# 收发单聊消息

QQ 单聊消息通过 `C2C_MESSAGE_CREATE` 事件送达。

## 先做什么

1. 在平台开通所需能力并配置[事件订阅](../../../getting-started/events/README.md)。WebSocket 的普通单聊消息使用 `GROUP_AND_C2C_EVENT`。
2. [接收单聊消息](receive.md)，从 `ctx.Message()` 读取用户、正文和附件。
3. [发送单聊消息](send.md)，直接使用 `ctx.ReplyText` 或组合消息构造器。

## 最小处理函数

将 `HandlePrivate` 传给 `bot.OnMessage(message.C2C_MESSAGE_CREATE, HandlePrivate)`。导入及注册方式见[接收单聊消息](receive.md)。

```go
package guide

import "github.com/ProgramCX/GoQQBot/shared/types"

func HandlePrivate(ctx *types.Context) error {
	return ctx.ReplyText("已收到：" + ctx.Text())
}
```

## 场景限制

按发送场景选择接口。发送普通图片、视频或语音前，先[上传富媒体](../rich-media/c2c-upload.md)，再用 `file_info` 发送。[Markdown 图片素材](../material-management.md)使用独立的上传流程。

## 注意事项

- 频道私信使用 `DIRECT_MESSAGE_CREATE`。频道私信与 QQ 单聊的用户标识、发送路径各不相同。
- `user_openid` 是当前机器人取得的单聊标识。普通 QQ 号码和群聊的 `member_openid` 无法替代它。
- QQ 检查回复来源的有效期和消息次数。主动消息（proactive message）和唤醒消息还受资格及配额限制。
- 输入状态和流式消息仅适用于单聊。

## 另请参阅

[接收单聊消息](receive.md) · [发送单聊消息](send.md) · [单聊流式消息](../streaming.md) · [官方发送单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)
