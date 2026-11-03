# 入群申请与自动审批

SDK 当前没有入群申请列表、审批、策略管理的专用方法。此页用于定位官方能力和安排业务边界，不提供自动批准或移除成员的默认行为。

## 申请处理

| 操作 | 官方文档 | 接口 |
| --- | --- | --- |
| 查询申请 | [入群申请列表拉取](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_join_request_list.get.html) | `GET /v2/groups/{group_openid}/join_request_list` |
| 作出审批决定 | [入群申请审批](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_approval_join_request_member_openid.post.html) | `POST /v2/groups/{group_openid}/approval_join_request/{member_openid}` |

收到事件后，将申请标识交给业务审批流程。审批请求的字段和可选值以对应官方页面为准；SDK 的交互按钮 ACK 只确认按钮事件，不能代替入群审批请求。

## 自动审批策略

| 操作 | 官方文档 |
| --- | --- |
| 创建 | [创建入群自动审批策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.post.html) |
| 列表 | [查询策略列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.get.html) |
| 修改 | [修改策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.patch.html) |
| 删除 | [删除策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.delete.html) |
| 执行 | [执行策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_execute.post.html) |
| 修改白名单 | [修改策略白名单号码](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_whitelist_users.post.html) |

需要这些接口时，将 `api.Client` 注入业务服务，通过 `Request` 显式调用。应用负责权限校验、人工确认、持久化和操作结果的不确定状态处理；不要给审批 POST 增加无条件自动重试。

## 接收事件

当前 WebSocket 的 intent 配置尚不支持 `GROUP_MEMBER_EVENT`，详情见[群聊管理概述](README.md)。即使已有 `OnEvent` 路由，也不能代替平台订阅和相应事件 DTO。

## 另请参阅

- [事件路由](../../getting-started/events/routing.md)
- [原始 API 扩展](../../advanced/raw-api.md)
- [错误处理](../../logging/errors.md)
