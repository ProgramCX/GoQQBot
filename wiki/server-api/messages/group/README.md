# 群聊消息

群聊 @机器人消息与群全量消息使用不同事件名。选择平台已开通的事件注册处理函数，发送时则可以使用相同的上下文回复入口。

## 事件与用途

| 事件 | 用途 |
| --- | --- |
| `GROUP_AT_MESSAGE_CREATE` | 群里 @机器人触发的消息 |
| `GROUP_MESSAGE_CREATE` | 已获准接收的群全量消息；仅注册处理函数不会开通权限 |

## 共用业务函数

以下函数可以同时注册到单聊与群聊消息类型。

```go
package guide

import "github.com/ProgramCX/GoQQBot/shared/types"

func ReplyShared(ctx *types.Context) error {
	return ctx.ReplyText("当前会话的业务处理已完成。")
}
```

不必在业务函数中拼群消息路径，也不必把 `group_openid` 传给回复方法。需要在后台发消息时，使用明确的 `message.Target{Scene: message.SceneGroup, ID: groupOpenID}`。

## 权限与边界

`group_openid` 标识群会话，`member_openid` 标识该场景的用户，都不是普通群号或 QQ 号。群管理、禁言、成员移除等不是消息发送能力；当前 SDK 没有完整群管理领域封装。

## 另请参阅

[接收群聊消息](receive.md) · [发送群聊消息](send.md) · [官方群 @消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html) · [官方群全量消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_message_create.html)
