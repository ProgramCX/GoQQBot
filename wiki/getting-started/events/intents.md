# 查找事件对应的 Intent

在 `connection.websocket.intents` 填订阅组名，在处理函数注册时填事件名。下表覆盖 SDK 当前接受的 14 个订阅组。QQ 在外层事件的 `t` 字段提供具体事件名。

| Intent 配置值 | 事件名 | 详细说明 |
| --- | --- | --- |
| `GUILDS` | `GUILD_CREATE`、`GUILD_UPDATE`、`GUILD_DELETE`、`CHANNEL_CREATE`、`CHANNEL_UPDATE`、`CHANNEL_DELETE` | [频道与子频道](../../server-api/channel-management/guilds-and-channels.md) |
| `GUILD_MEMBERS` | `GUILD_MEMBER_ADD`、`GUILD_MEMBER_UPDATE`、`GUILD_MEMBER_REMOVE` | [频道成员](../../server-api/channel-management/members-and-permissions.md) |
| `GUILD_MESSAGES` | `MESSAGE_CREATE`、`MESSAGE_DELETE` | [频道消息](../../server-api/messages/channel/README.md) |
| `GUILD_MESSAGE_REACTIONS` | `MESSAGE_REACTION_ADD`、`MESSAGE_REACTION_REMOVE` | [表态事件](../../server-api/messages/channel/reactions.md) |
| `DIRECT_MESSAGE` | `DIRECT_MESSAGE_CREATE`、`DIRECT_MESSAGE_DELETE` | [频道私信](../../server-api/messages/channel/messages.md) |
| `OPEN_FORUM_EVENT` | `OPEN_FORUM_THREAD_CREATE`、`OPEN_FORUM_THREAD_UPDATE`、`OPEN_FORUM_THREAD_DELETE`、`OPEN_FORUM_POST_CREATE`、`OPEN_FORUM_POST_DELETE`、`OPEN_FORUM_REPLY_CREATE`、`OPEN_FORUM_REPLY_DELETE` | [频道论坛事件](../../server-api/channel-management/content.md) |
| `AUDIO_OR_LIVE_CHANNEL_MEMBER` | `AUDIO_OR_LIVE_CHANNEL_MEMBER_ENTER`、`AUDIO_OR_LIVE_CHANNEL_MEMBER_EXIT` | [音视频成员事件](../../server-api/channel-management/content.md) |
| `GROUP_MEMBER_EVENT` | `GROUP_JOIN_REQUEST`、`GROUP_MEMBER_ADD`、`GROUP_MEMBER_REMOVE` | [官方入群申请](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_join_request.html) · [成员加入](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_member_add.html) · [成员退出](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_member_remove.html) |
| `GROUP_AND_C2C_EVENT` | `C2C_MESSAGE_CREATE`、`FRIEND_ADD`、`FRIEND_DEL`、`C2C_MSG_REJECT`、`C2C_MSG_RECEIVE`、`GROUP_AT_MESSAGE_CREATE`、`GROUP_ADD_ROBOT`、`GROUP_DEL_ROBOT`、`GROUP_MSG_REJECT`、`GROUP_MSG_RECEIVE` | [QQ 单聊](../../server-api/messages/c2c/receive.md) · [QQ群聊](../../server-api/messages/group/receive.md) |
| `INTERACTION` | `INTERACTION_CREATE` | [按钮与菜单回调](../../server-api/messages/interactions/button-callbacks.md) |
| `MESSAGE_AUDIT` | `MESSAGE_AUDIT_PASS`、`MESSAGE_AUDIT_REJECT` | [频道发送与审核](../../server-api/messages/channel/messages.md) |
| `FORUMS_EVENT` | `FORUM_THREAD_CREATE`、`FORUM_THREAD_UPDATE`、`FORUM_THREAD_DELETE`、`FORUM_POST_CREATE`、`FORUM_POST_DELETE`、`FORUM_REPLY_CREATE`、`FORUM_REPLY_DELETE`、`FORUM_PUBLISH_AUDIT_RESULT` | [论坛事件](../../server-api/channel-management/content.md) |
| `AUDIO_ACTION` | `AUDIO_START`、`AUDIO_FINISH`、`AUDIO_ON_MIC`、`AUDIO_OFF_MIC` | [音频事件](../../server-api/channel-management/content.md) |
| `PUBLIC_GUILD_MESSAGES` | `AT_MESSAGE_CREATE`、`PUBLIC_MESSAGE_DELETE` | [频道 @ 消息](../../server-api/messages/channel/README.md) |

例如，接收普通 QQ 单聊、群 @消息及入群审批时，可配置：

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_AND_C2C_EVENT
      - GROUP_MEMBER_EVENT
```

## 注意事项

- 平台控制事件订阅资格。未授权的订阅可能让 WebSocket 鉴权失败。修改后重启连接，优先添加当前要用且已获准的订阅。
- 上表按 WebSocket 协议的 Intent 总表区分频道全量与 @消息。QQ群全量事件 `GROUP_MESSAGE_CREATE` 也使用 `GROUP_AND_C2C_EVENT`，需另行取得平台开放资格。[官方QQ群全量事件说明](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_message_create.html)
- 配置 Intent 后还需注册处理函数。只有原始事件结构的场景可通过 `OnEvent` 读取 `ctx.Payload.D`。Webhook 在 QQ 开放平台选择事件。

## 另请参阅

[官方 Intent 与下属事件完整表](https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/event-emit.html#事件订阅intents) · [选择 WebSocket 或 Webhook](connection.md) · [注册处理函数](routing.md)
