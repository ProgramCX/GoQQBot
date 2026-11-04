# 管理群聊

查询群资料、管理成员、禁言和审批入群，可以直接调用 QQ 官方接口。SDK 还没有为这些操作提供专用方法。收发消息请看[群聊消息收发](../messages/group/README.md)。

## 能力对应表

按任务选择官方接口，再通过[原始 API](../../advanced/raw-api.md)调用。

| 任务 | 官方接口 | 当前接入方式 |
| --- | --- | --- |
| [获取群基本信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_info.get.html) | `GET /v2/groups/{group_openid}/info` | 原始 API |
| [获取机器人群内状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_bot_state.get.html) | `GET /v2/groups/{group_openid}/bot_state` | 原始 API |
| [查询禁言状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.get.html) | `GET /v2/groups/{group_openid}/restrict_chat_setting` | 原始 API |
| [设置成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.post.html) | `POST /v2/groups/{group_openid}/restrict_chat_setting` | 原始 API，业务明确授权后执行 |
| 成员、移除和黑名单 | [群成员管理](members.md) | 原始 API |
| 入群申请和自动审批 | [入群审批](join-requests.md) | 原始 API |

## 获取当前群标识

在群消息处理函数中，可从 `ctx.Message()` 的结果读取 `GroupOpenID`，也可调用消息的 `ConversationID()`。将群 OpenID 传给管理服务即可。后台服务可以直接从业务记录读取群 OpenID，无需构造 `Context`。

路径中的标识用 `url.PathEscape` 编码。查询参数用 `url.Values` 编码。

## 事件订阅限制

按接入方式选择订阅配置。官方群成员及入群申请事件使用 `GROUP_MEMBER_EVENT`。

[Webhook](../../getting-started/events/webhook.md) 可按平台开放范围配置事件。收到事件后，由业务读取原始事件数据。

## 注意事项

- QQ 平台检查管理员身份、接口白名单、操作次数和业务资格。存在接口路径不表示每个机器人都能使用。
- 群 OpenID、QQ 群号和频道 `guild_id` 分别标记不同对象。调用时按接口选择标识。
- 当前 SDK 的 `config.IntentMask` 不接受 `GROUP_MEMBER_EVENT`。直接填入 WebSocket 配置会导致订阅失败。
- 普通群消息使用 `GROUP_AND_C2C_EVENT`。频道成员事件使用 `GUILD_MEMBERS`。两项订阅都无法代替群成员事件订阅。

## 另请参阅

[官方入群申请事件及 intent](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_join_request.html) · [原始 API 扩展](../../advanced/raw-api.md) · [SDK 覆盖范围](../../reference/coverage.md)
