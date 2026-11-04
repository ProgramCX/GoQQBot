# 频道管理

调用频道接口前，先确认操作对象是频道、子频道还是频道私信会话。

- 频道使用 `guild`。
- 子频道使用 `channel`。
- 频道私信会话使用私信 `guild_id`。

`channel.Client` 提供频道、子频道、成员、身份组、权限和内容管理方法，也可用于[频道消息](../messages/channel/README.md)。

| 官方模块 | 本地说明 | 当前状态 |
| --- | --- | --- |
| 频道和子频道资料 | [资料查询与子频道管理](guilds-and-channels.md) | 频道列表、详情和子频道增删改查 |
| 成员、身份组与权限 | [成员和权限](members-and-permissions.md) | 列表、详情、禁言及权限管理 |
| 公告、精华、日程、帖子、音频 | [内容管理](content.md) | 公告、精华、日程、帖子、音频控制及上下麦 |
| 子频道消息、频道私信 | [频道消息](../messages/channel/messages.md) | 创建会话、发送和撤回 |
| 消息表态 | [添加、删除和查询表态](../messages/channel/reactions.md) | 添加、删除和分页查询 |

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GUILDS
      - GUILD_MEMBERS
```

[全部 Intent 与下属事件](../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`GUILD_CREATE`、`GUILD_UPDATE`、`GUILD_DELETE`、`CHANNEL_CREATE`、`CHANNEL_UPDATE`、`CHANNEL_DELETE`，以及需要处理的 `GUILD_MEMBER_ADD`、`GUILD_MEMBER_UPDATE`、`GUILD_MEMBER_REMOVE`。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../getting-started/events/webhook.md)

## 接入方式

用 `channel.NewClient(client)` 创建频道客户端。`client` 可以来自 `api.NewClient` 或消息处理函数的 `ctx.Client`。管理方法接收标准库的 `context.Context`，后台任务可独立调用。

尚未封装的接口可通过 [Client.Request](../../advanced/raw-api.md) 调用。

## 注意事项

- QQ 群聊的 `group_openid` 无法代替频道、子频道或频道私信标识。
- `GUILDS` 等 intent 控制[事件订阅](../../getting-started/events/README.md)。注册后可通过 `ctx.GuildChange()`、`ctx.SubchannelChange()` 和 `ctx.GuildMemberChange()` 读取相应事件。平台另外检查频道管理接口权限。

## 另请参阅

[官方频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id.get.html) · [官方子频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.get.html) · [SDK 覆盖范围](../../reference/coverage.md)
