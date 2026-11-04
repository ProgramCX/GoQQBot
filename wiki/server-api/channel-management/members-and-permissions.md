# 频道成员、身份组与权限

用 `channel.NewClient` 包装现有的 `*api.Client`，即可查询频道成员、管理身份组和设置发言权限。频道客户端共用原客户端的认证和 HTTP 设置。

```go
package guide

import (
    "context"

    "github.com/ProgramCX/GoQQBot/pkg/api"
    "github.com/ProgramCX/GoQQBot/pkg/api/channel"
)

func members(ctx context.Context, client *api.Client, guildID string) ([]channel.Member, error) {
    return channel.NewClient(client).ListMembers(ctx, guildID, channel.MemberListOptions{Limit: 100})
}
```

频道成员列表使用本页末位成员的 `User.ID` 作为下一页的 `After`，空页表示结束。身份组成员列表则将 `Next` 传给 `StartIndex`，`Next` 为 `"0"` 时表示已到末页。

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GUILD_MEMBERS
```

[全部 Intent 与下属事件](../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`GUILD_MEMBER_ADD`、`GUILD_MEMBER_UPDATE` 和 `GUILD_MEMBER_REMOVE`。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../getting-started/events/webhook.md)

## 按任务查找官方接口

按操作选择对应的官方字段表。

| 任务 | SDK 方法 | 官方文档 |
| --- | --- | --- |
| 成员列表和详情 | `ListMembers`、`GetMember` | [成员列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_members.html)、[成员详情](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_member.html) |
| 身份组成员 | `ListRoleMembers`、`AddMemberRole`、`RemoveMemberRole` | [身份组成员列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_role_members.html)、[添加成员](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/put_guild_member_role.html) |
| 移除成员 | `RemoveMember` | [删除频道成员](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/delete_member.html) |
| 身份组 | `ListRoles`、`CreateRole`、`UpdateRole`、`DeleteRole` | [获取身份组](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/get_guild_roles.html)、[创建身份组](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/post_guild_role.html) |
| 子频道用户权限 | `GetMemberPermissions`、`UpdateMemberPermissions` | [获取权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/get_channel_permissions.html)、[修改权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/put_channel_permissions.html) |
| 子频道身份组权限 | `GetRolePermissions`、`UpdateRolePermissions` | [获取身份组权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/get_channel_roles_permissions.html)、[修改身份组权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/put_channel_roles_permissions.html) |
| 接口授权 | `ListAPIPermissions`、`RequestAPIPermission` | [可用接口权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/api_permissions/get_guild_api_permission.html)、[发送授权链接](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/api_permissions/post_api_permission_demand.html) |
| 禁言 | `SetMemberMute`、`SetMembersMute`、`SetGuildMute` | [指定成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_member_mute.html)、[批量禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_mute_multi_member.html)、[全员禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_mute.html) |
| 在线人数与主动消息设置 | `GetOnlineCount`、`GetMessageSettings` | [在线人数](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/get_online_nums.html)、[消息设置](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/setting/message_setting.html) |



## 事件与数据归属

通过 `GUILD_MEMBERS` [订阅事件](../../getting-started/events/README.md)。用 `OnEvent` 注册 `message.GUILD_MEMBER_ADD`、`message.GUILD_MEMBER_UPDATE` 或 `message.GUILD_MEMBER_REMOVE`，然后调用 `ctx.GuildMemberChange()` 读取用户、身份组和操作者。音视频或直播子频道的成员进出事件使用 `ctx.AudioLiveMember()`，订阅方式见[接收论坛和音频事件](content.md#接收论坛和音频事件)。

## 注意事项

::: warning 频道成员列表需要私域权限
测试全频道成员列表时，QQ 返回 `40012010`（应用无接口访问权限）。当前测试应用无法验证该接口的成功响应。QQ 仅向私域机器人开放此接口。权限开通后，请将机器人从频道移除，再重新添加。频道成员详情和身份组成员列表使用独立接口，两项查询在 QQ 测试中均返回成功。
:::

- 收到频道消息不表示有管理权限。`AdminOnly()` 只描述按钮交互权限。服务端仍需检查业务权限和平台授权。
- 注册处理函数后，还需订阅对应事件。旧版成员事件的 `joined_at` 和 `roles` 可能与官方示例的格式不同。若解析失败，可按实际字段读取 `ctx.Payload.D`。[官方频道成员事件](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/guild_member.html)
- 普通 QQ 号码无法替代 OpenID。群成员身份组与频道身份组的权限不可互用。
- 成员列表与身份组成员列表面向私域机器人。频道成员翻页时可能出现重复项，业务可按用户 ID 去重。
- 添加或移除子频道管理员身份组（ID 为 `"5"`）的成员时，填写 `RoleMembershipOptions.ChannelID`。创建或修改身份组时，`RoleOptions` 中非 nil 的字段才会发送。
- 批量禁言仅返回操作成功的成员 ID。将 `MuteOptions.Seconds` 设为 `"0"` 可解除禁言。若同时填写截止时间，则 QQ 以截止时间为准。
- 查询在线人数适用于音视频或直播子频道。

## 另请参阅

[原始 API 扩展](../../advanced/raw-api.md) · [事件路由](../../getting-started/events/routing.md) · [按钮权限](../messages/interactions/buttons.md)
