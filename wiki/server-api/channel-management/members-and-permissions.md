# 频道成员、身份组与权限

成员和权限管理通过 `Client.Request` 调用。SDK 尚无专用请求和响应模型。业务层按官方字段表定义模型。

## 按任务查找官方接口

按操作选择对应的官方字段表。

| 任务 | 官方文档 |
| --- | --- |
| 成员列表和详情 | [成员列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_members.html)、[成员详情](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_member.html) |
| 身份组成员 | [身份组成员列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_role_members.html) |
| 移除成员 | [删除频道成员](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/delete_member.html) |
| 身份组 | [获取身份组](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/get_guild_roles.html)、[创建身份组](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/post_guild_role.html) |
| 子频道用户权限 | [获取权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/get_channel_permissions.html)、[修改权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/put_channel_permissions.html) |
| 子频道身份组权限 | [获取身份组权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/get_channel_roles_permissions.html) |
| 接口授权 | [可用接口权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/api_permissions/get_guild_api_permission.html)、[发送授权链接](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/api_permissions/post_api_permission_demand.html) |
| 禁言 | [指定成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_member_mute.html)、[全员禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_mute.html) |



## 事件与数据归属

通过 `GUILDS`、`GUILD_MEMBERS` 等 intent [订阅事件](../../getting-started/events/README.md)。用 `OnEvent` 接收原始事件数据。缺少专用事件常量时，用 `message.EventId("官方事件名")`。

成员数据和频道身份组按机器人与频道管理。

## 注意事项

- 收到频道消息不表示有管理权限。`AdminOnly()` 只描述按钮交互权限。服务端仍需检查业务权限和平台授权。
- 注册处理函数不会自动订阅事件，也不会为事件生成用于解析数据的结构（DTO）。
- 普通 QQ 号码无法替代 OpenID。群成员身份组与频道身份组的权限不可互用。

## 另请参阅

[原始 API 扩展](../../advanced/raw-api.md) · [事件路由](../../getting-started/events/routing.md) · [按钮权限](../messages/interactions/buttons.md)
