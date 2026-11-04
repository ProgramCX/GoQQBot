# 收发群聊消息

群聊中 @机器人的消息和群内全量消息分别使用不同的事件。按平台已开放的权限订阅，收到消息后可用同一个回复方法。

## 事件与用途

| 事件 | 用途 |
| --- | --- |
| `GROUP_AT_MESSAGE_CREATE` | 群里 @机器人触发的消息 |
| `GROUP_MESSAGE_CREATE` | 已获准接收的群全量消息 |

## 共用业务函数

将下面的函数同时注册到单聊与群聊消息类型。

```go
package guide

import "github.com/ProgramCX/GoQQBot/shared/types"

func ReplyShared(ctx *types.Context) error {
	return ctx.ReplyText("当前会话的业务处理已完成。")
}
```

上下文会选择群消息路径和 `group_openid`。后台发消息时，使用明确的 `message.Target{Scene: message.SceneGroup, ID: groupOpenID}`。

## 权限与边界

区分两个 OpenID。`group_openid` 标识群会话，`member_openid` 标识当前场景的用户。

## 注意事项

- 注册处理函数不会开通群全量消息权限。订阅前需确认平台已开放。
- 群会话和成员的 OpenID 分别不同于普通群号和 QQ 号。
- [群管理](../../group-management/README.md)与消息发送是两种能力。SDK 尚未完整封装群管理、禁言和成员移除。

## 另请参阅

[接收群聊消息](receive.md) · [发送群聊消息](send.md) · [官方群 @消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html) · [官方群全量消息事件](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_message_create.html)
