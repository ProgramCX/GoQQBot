# 群聊管理

本分类对应官方“群聊管理”，区别于[群聊消息收发](../messages/group/README.md)。当前 SDK 尚无群资料、成员管理、禁言和入群审批的类型化封装。

## 能力对应表

| 任务 | 官方接口 | 当前接入方式 |
| --- | --- | --- |
| [获取群基本信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_info.get.html) | `GET /v2/groups/{group_openid}/info` | 原始 API |
| [获取机器人群内状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_bot_state.get.html) | `GET /v2/groups/{group_openid}/bot_state` | 原始 API |
| [查询禁言状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.get.html) | `GET /v2/groups/{group_openid}/restrict_chat_setting` | 原始 API |
| [设置成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.post.html) | `POST /v2/groups/{group_openid}/restrict_chat_setting` | 原始 API，业务明确授权后执行 |
| 成员、移除和黑名单 | [群成员管理](members.md) | 原始 API |
| 入群申请和自动审批 | [入群审批](join-requests.md) | 原始 API |

是否获得管理员身份、是否在接口白名单内、操作次数和业务资格由 QQ 平台判断。存在 API 路径并不表示每个机器人都能使用。

## 获取当前群标识

在群消息处理函数中，用 `ctx.Message()` 读取 `GroupOpenID` 或 `ConversationID()`，再将该值交给应用自己的管理服务。后台服务可以通过已有业务映射获得群 OpenID，不需要伪造 `Context`。

群 OpenID、普通 QQ 群号、频道 `guild_id` 是不同标识。原始 API 路径中的标识要用 `url.PathEscape` 编码，查询参数用 `url.Values` 编码。

## 事件订阅限制

群成员及入群申请事件使用官方 `GROUP_MEMBER_EVENT` intent。当前 SDK 的 `config.IntentMask` 尚不接受该名称，不能在 WebSocket 配置中直接填写并期待订阅成功。Webhook 可以按平台开放情况配置事件，但仍需业务自行解码原始载荷。

这个限制与普通群消息的 `GROUP_AND_C2C_EVENT` 不同，也不应与频道成员的 `GUILD_MEMBERS` 混用。

## 另请参阅

- [官方入群申请事件及 intent](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_join_request.html)
- [原始 API 扩展](../../advanced/raw-api.md)
- [SDK 覆盖范围](../../reference/coverage.md)
