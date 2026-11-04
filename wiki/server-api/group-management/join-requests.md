# 处理入群申请与自动审批

查询和审批入群申请，或管理自动审批策略，都可以直接调用 QQ 官方接口。

## 申请处理

收到入群申请事件后，可以把申请标识传给自己的审批流程。提交结果时，按对应接口的官方文档填写字段。

| 操作 | 官方文档 | 接口 |
| --- | --- | --- |
| 查询申请 | [入群申请列表拉取](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_join_request_list.get.html) | `GET /v2/groups/{group_openid}/join_request_list` |
| 作出审批决定 | [入群申请审批](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_approval_join_request_member_openid.post.html) | `POST /v2/groups/{group_openid}/approval_join_request/{member_openid}` |

## 自动审批策略

将 `api.Client` 注入业务服务，再用 [`Request`](../../advanced/raw-api.md) 显式调用下列接口。

| 操作 | 官方文档 |
| --- | --- |
| 创建 | [创建入群自动审批策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.post.html) |
| 列表 | [查询策略列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.get.html) |
| 修改 | [修改策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.patch.html) |
| 删除 | [删除策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.delete.html) |
| 执行 | [执行策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_execute.post.html) |
| 修改白名单 | [修改策略白名单号码](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_whitelist_users.post.html) |

应用管理审批操作。

- 校验权限。
- 按业务流程人工确认。
- 保存审批状态。
- 处理结果不确定的请求。

## 接收事件

先按[群聊管理概述](README.md)选择平台支持的接收方式。再用 `OnEvent` 注册事件处理函数。

## 注意事项

- SDK 尚无入群申请列表、审批和策略管理的专用方法。SDK 不默认批准申请或移除成员。
- 交互按钮 ACK 只确认按钮事件，无法替代入群审批请求。
- 审批 POST 的结果可能不确定。无条件重试可能再次提交审批。
- 当前 WebSocket intent 配置不支持 `GROUP_MEMBER_EVENT`。`OnEvent` 只负责把收到的事件交给处理函数。事件仍需订阅，数据仍需用对应结构体解码。

## 另请参阅

[事件路由](../../getting-started/events/routing.md) · [原始 API 扩展](../../advanced/raw-api.md) · [错误处理](../../logging/errors.md)
