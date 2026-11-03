# 频道管理

“频道”是 `guild`，“子频道”是 `channel`，频道私信会话则使用私信 `guild_id`。不要把这些标识与 QQ 群聊的 `group_openid` 混用。

当前 SDK 可收发频道消息，并有旧版表态接口；频道资料、成员、身份组、权限和内容管理尚无完整的领域封装。选择任务时先区分[频道消息](../messages/channel/README.md)和本目录的管理接口。

| 官方模块 | 本地说明 | 当前状态 |
| --- | --- | --- |
| 频道和子频道资料 | [资料查询与子频道管理](guilds-and-channels.md) | 原始 API，自行建立响应模型 |
| 成员、身份组与权限 | [成员和权限](members-and-permissions.md) | 尚未类型化封装 |
| 公告、精华、日程、帖子、音频 | [内容管理](content.md) | 尚未类型化封装 |
| 子频道消息、频道私信 | [频道消息](../messages/channel/messages.md) | 发送和撤回已有；创建私信会话尚未封装 |
| 消息表态 | [添加、删除和查询表态](../messages/channel/reactions.md) | 旧版专用方法，存在查询参数编码限制 |

## 接入方式

通过 `api.NewClient` 创建独立客户端，或在处理函数中使用 `ctx.Client`。管理接口不需要构造一条虚假的消息 `Context`；后台业务服务直接接收 `*api.Client` 和操作所需 ID 即可。

如果所需方法尚不存在，使用 [Client.Request](../../advanced/raw-api.md) 的显式扩展方式。不要把开放了 `GUILDS` 等事件 intent 理解为获得所有频道管理接口权限，两者是不同的检查。

## 另请参阅

- [官方频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id.get.html)
- [官方子频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.get.html)
- [SDK 覆盖范围](../../reference/coverage.md)
