# 管理群聊

通过 `group.NewClient` 查询群资料、管理成员、设置禁言和审批入群。群聊管理客户端复用已有 `api.Client` 的认证和 HTTP 配置，可用于事件处理函数和后台服务。收发消息请看[群聊消息收发](../messages/group/README.md)。

## 配置事件订阅

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - GROUP_MEMBER_EVENT
```

[全部 Intent 与下属事件](../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`GROUP_JOIN_REQUEST`，以及「群事件」里的 **群用户添加**（`GROUP_MEMBER_ADD`）和 **群用户移除**（`GROUP_MEMBER_REMOVE`）。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../getting-started/events/webhook.md)

## 能力对应表

| 任务 | SDK 方法 |
| --- | --- |
| [获取群基本信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_info.get.html) | `GetInfo` |
| [获取机器人群内状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_bot_state.get.html) | `GetBotState` |
| [查询禁言状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.get.html) | `GetMuteSettings` |
| [设置成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.post.html) | `SetMemberMutes` |
| 成员、移除和黑名单 | [群成员管理](members.md) |
| 入群申请和自动审批 | [入群审批](join-requests.md) |

## 获取当前群标识

在群消息处理函数中，可从 `ctx.Message()` 的结果读取 `GroupOpenID`，也可调用消息的 `ConversationID()`。将群 OpenID 传给管理方法即可。后台服务也可以从业务记录读取群 OpenID。

## 注意事项

- QQ 平台检查管理员身份、接口白名单、操作次数和业务资格。
- 群 OpenID、QQ 群号和频道 `guild_id` 分别标记不同对象。调用时按接口选择标识。
- 普通群消息使用 `GROUP_AND_C2C_EVENT`。频道成员事件使用 `GUILD_MEMBERS`。两项订阅都无法代替群成员事件订阅。
- `GetMuteSettings` 可读取全员禁言配置。`SetMemberMutes` 每次最多提交 20 项成员禁言操作。QQ 要求机器人具有群管理员身份，且单次禁言时长不超过 30 天。

## 另请参阅

[官方入群申请事件及 intent](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_join_request.html) · [原始 API 扩展](../../advanced/raw-api.md) · [SDK 覆盖范围](../../reference/coverage.md)
