# 频道成员、身份组与权限

当前 SDK 没有本页管理能力的专用请求和响应模型。需要时使用 `Client.Request` 显式调用，并在业务层建立与官方字段表对应的模型。

## 按任务查找官方接口

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

SDK 不会因为调用者能收到某个频道的消息，就推断其有权执行管理操作。按钮设置 `AdminOnly()` 也是交互权限描述，不能代替服务端业务权限检查和平台接口授权。

## 事件与数据归属

通过 `GUILDS`、`GUILD_MEMBERS` 等官方 intent 订阅相关事件，再用 `OnEvent` 接收原始载荷。没有专用事件常量时，可以显式使用 `message.EventId("官方事件名")`，但注册路由不会自动订阅该事件，也不会自动生成对应 DTO。

成员数据与频道身份组应按当前机器人和频道管理；不要用普通 QQ 号码替换 OpenID，或把群成员身份组与频道身份组权限相互套用。

## 另请参阅

- [原始 API 扩展](../../advanced/raw-api.md)
- [事件路由](../../getting-started/events/routing.md)
- [按钮权限](../messages/interactions/buttons.md)
