# 查询和管理群成员

通过 `ListMembers` 查询一页群成员。将结果中的 `NextCursor` 传给下一次调用。游标为空时，表示已经查到最后一页。

## 查询一页成员

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/group"
)

func ReadMembers(ctx context.Context, client *api.Client, groupOpenID, cursor string) (*group.MemberPage, error) {
	return group.NewClient(client).ListMembers(ctx, groupOpenID, cursor)
}
```

## 其他成员操作

| 官方能力 | SDK 方法 |
| --- | --- |
| [获取群成员信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members_member_openid.get.html) | `GetMember` |
| [群成员批量移除](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_batch_remove_members.post.html) | `RemoveMembers` |
| [群黑名单查询](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.get.html) | `ListBlacklist` |
| [群黑名单操作](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.post.html) | `UpdateBlacklist` |

应用负责核对成员身份、检查操作权限并记录执行结果。遇到 QQ 返回的错误时，可通过 `errors.As` 读取 `*api.APIError`。详见[错误处理](../../logging/errors.md)。

## 注意事项

::: warning 群成员与黑名单查询需要白名单权限
测试群成员列表、成员详情和黑名单查询时，QQ 均返回 `40012010`（应用无接口访问权限）。这三个接口只对加入白名单的机器人开放，需联系 QQ 平台运营申请。成功响应仍需用已获授权的机器人测试。
:::

- 官方规定成员列表每页最多 30 条。黑名单默认每页 20 条，最多 100 条。SDK 每次请求只取一页。
- 接口受内邀或白名单限制。遇到权限错误时，先确认平台资格。
- `RemoveMembers` 和 `UpdateBlacklist` 每次最多操作 20 个成员。接口成功时仍应检查结果中的失败名单。
- 仍在群内的成员无法通过 `UpdateBlacklist` 加入黑名单。需要同时移除并拉黑时，可为 `RemoveMembers` 设置 `AddToMemberBlacklist`。
- 多个协程同时读写同一份分页结果，可能出现数据竞争。共享原始事件数据时也是如此。

## 另请参阅

[官方获取群成员列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members.get.html) · [群聊管理概述](README.md) · [入群审批](join-requests.md)
