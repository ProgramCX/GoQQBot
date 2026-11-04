# 频道管理

频道管理接口按频道和子频道区分。

- 频道使用 `guild`。
- 子频道使用 `channel`。
- 频道私信会话使用私信 `guild_id`。

SDK 可收发[频道消息](../messages/channel/README.md)，也保留旧版表态接口。频道资料、成员、身份组、权限和内容管理尚无完整封装。

| 官方模块 | 本地说明 | 当前状态 |
| --- | --- | --- |
| 频道和子频道资料 | [资料查询与子频道管理](guilds-and-channels.md) | 原始 API，自行建立响应模型 |
| 成员、身份组与权限 | [成员和权限](members-and-permissions.md) | 尚未类型化封装 |
| 公告、精华、日程、帖子、音频 | [内容管理](content.md) | 尚未类型化封装 |
| 子频道消息、频道私信 | [频道消息](../messages/channel/messages.md) | 支持发送和撤回。创建私信会话尚未封装。 |
| 消息表态 | [添加、删除和查询表态](../messages/channel/reactions.md) | 旧版专用方法，存在查询参数编码限制 |

## 接入方式

用 `api.NewClient` 创建独立客户端。处理函数也可以使用 `ctx.Client`。后台服务直接接收 `*api.Client` 和操作所需的 ID，无需构造消息 `Context`。

尚未封装的接口可通过 [Client.Request](../../advanced/raw-api.md) 调用。

## 注意事项

- QQ 群聊的 `group_openid` 无法代替频道、子频道或频道私信标识。
- `GUILDS` 等 intent 控制[事件订阅](../../getting-started/events/README.md)。平台另外检查频道管理接口权限。

## 另请参阅

[官方频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id.get.html) · [官方子频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.get.html) · [SDK 覆盖范围](../../reference/coverage.md)
