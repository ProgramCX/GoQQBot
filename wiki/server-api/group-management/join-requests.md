# 处理入群申请与自动审批

通过 `group.Client` 查询或审批入群申请，也可以管理自动审批策略。

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

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选`GROUP_JOIN_REQUEST`（入群申请）。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../getting-started/events/webhook.md)

## 申请处理

收到入群申请事件或查询结果后，可将申请交给业务审批流程。提交决定时，从申请中取出成员 OpenID 和申请 ID。

| 操作 | SDK 方法 | 官方文档 |
| --- | --- | --- |
| 查询申请 | `ListJoinRequests` | [入群申请列表拉取](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_join_request_list.get.html) |
| 提交审批决定 | `DecideJoin` | [入群申请审批](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_approval_join_request_member_openid.post.html) |

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api/group"
)

func ApproveReviewedRequest(ctx context.Context, client *group.Client, groupOpenID string, request group.JoinRequest) error {
	return client.DecideJoin(ctx, groupOpenID, request.MemberOpenID, group.DecideJoinRequest{
		Decision:      group.JoinApprove,
		JoinRequestID: request.JoinRequestID,
	})
}
```

## 自动审批策略

| 操作 | SDK 方法 | 官方文档 |
| --- | --- | --- |
| 创建 | `CreateStrategy` | [创建入群自动审批策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.post.html) |
| 列表 | `ListStrategies` | [查询策略列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.get.html) |
| 修改 | `UpdateStrategy` | [修改策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.patch.html) |
| 删除 | `DeleteStrategy` | [删除策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.delete.html) |
| 执行 | `ExecuteStrategy` | [执行策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_execute.post.html) |
| 修改白名单 | `UpdateStrategyWhitelist` | [修改策略白名单号码](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_whitelist_users.post.html) |

创建策略时，群 OpenID 和 QQ 群号二选一。调整关联群时，沿用创建时的标识类型。白名单使用字符串形式的 QQ 号。

## 接收事件

按[群聊管理概述](README.md)配置订阅，再用 `OnEvent` 注册 `message.GROUP_JOIN_REQUEST`。处理函数通过 `ctx.GroupJoinRequest()` 读取申请。QQ 已自动批准的申请带有 `AutoApproved`，业务可用其中的策略 ID 记录处理结果。

## 注意事项

- 业务负责核对权限、执行审批流程并保存审批状态。交互按钮 ACK 只确认按钮事件。
- 如果审批请求超时，先核对申请状态。直接重试可能再次提交决定。
- `ExecuteStrategy` 返回成功表示 QQ 已受理扫描请求。平台在后台执行，官方文档给出的完成时间约为 10 分钟。
- 查询申请和策略时，单页默认 20 条，最多 50 条。`NextCursor` 为空时表示已到末页。

## 另请参阅

[事件路由](../../getting-started/events/routing.md) · [原始 API 扩展](../../advanced/raw-api.md) · [错误处理](../../logging/errors.md)
