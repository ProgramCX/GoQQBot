# 频道资料与子频道管理

用 `channel.NewClient(client)` 创建频道客户端，可查询机器人所在的频道，并管理频道下的子频道。

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GUILDS
```

[全部 Intent 与下属事件](../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`GUILD_CREATE`、`GUILD_UPDATE`、`GUILD_DELETE`、`CHANNEL_CREATE`、`CHANNEL_UPDATE`、`CHANNEL_DELETE`，按需勾选。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../getting-started/events/webhook.md)

## 查询频道资料

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
)

func ReadGuild(ctx context.Context, client *api.Client, guildID string) (*channel.Guild, error) {
	return channel.NewClient(client).GetGuild(ctx, guildID)
}
```

用 `ListGuilds(ctx, channel.GuildListOptions{Limit: 20})` 查询第一页。查询后一页时，将当前页最后一项的 ID 填入 `After`。

## 子频道操作

按操作选择客户端方法。

| 操作 | 官方文档 | SDK 方法 |
| --- | --- | --- |
| 列表 | [获取子频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.get.html) | `ListChannels` |
| 创建 | [创建子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.post.html) | `CreateChannel` |
| 详情 | [获取子频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.get.html) | `GetChannel` |
| 修改 | [修改子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.patch.html) | `UpdateChannel` |
| 删除 | [删除子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.delete.html) | `DeleteChannel` |

`UpdateChannel` 接收 `ChannelPatch`。只设置需要修改的字段。指针为 `nil` 时省略该字段，指向 `0` 时会明确提交零值。

## 注意事项

- 做时间运算前，按平台格式解析时间戳。RFC3339 或 ISO8601 字符串无法直接解码成整数。
- QQ 检查管理权限。删除前的业务确认与审计由应用负责。
- 公开的端点变量会影响共用地址。指定自定义服务器时，应使用独立的客户端配置。

## 另请参阅

[官方获取频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id.get.html) · [自定义 HTTP 客户端](../../getting-started/api-calls.md) · [成员和权限](members-and-permissions.md)
