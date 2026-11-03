# 查询和管理群成员

当前没有 `ListGroupMembers` 一类 SDK 专用方法。以下函数是应用层封装示例，通过 `Client.Get` 获取一页成员；它不会自动翻页或修改群成员。

## 查询一页成员

```go
package guide

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

type Member struct {
	OpenID string `json:"member_openid"`
	Name   string `json:"username"`
	Role   string `json:"member_role"`
}

type MemberPage struct {
	Members    []Member `json:"members"`
	NextCursor string   `json:"next_cursor"`
}

func ReadMembers(ctx context.Context, client *api.Client, groupOpenID, cursor string) (*MemberPage, error) {
	if groupOpenID == "" {
		return nil, fmt.Errorf("group OpenID is required")
	}
	path := "/v2/groups/" + url.PathEscape(groupOpenID) + "/members"
	if cursor != "" {
		path += "?" + url.Values{"cursor": {cursor}}.Encode()
	}
	raw, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var page MemberPage
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, fmt.Errorf("decode member page: %w", err)
	}
	return &page, nil
}
```

官方当前规定每页最多 30 条，使用上页的 `next_cursor` 获取后续页，空游标表示末页。该接口仍有内邀/白名单限制；遇到权限错误先确认平台资格，不要通过循环重试绕过限制。

## 其他成员操作

下列接口同样尚未提供类型化 SDK 方法。操作前请从官方字段表建立应用请求模型，不要把查询示例改个 HTTP 方法就用于移除成员。

| 官方能力 | 请求路径 |
| --- | --- |
| [获取群成员信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members_member_openid.get.html) | `GET /v2/groups/{group_openid}/members/{member_openid}` |
| [群成员批量移除](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_batch_remove_members.post.html) | `POST /v2/groups/{group_openid}/batch_remove_members` |
| [群黑名单查询](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.get.html) | `GET /v2/groups/{group_openid}/member_blacklist` |
| [群黑名单操作](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.post.html) | `POST /v2/groups/{group_openid}/member_blacklist` |

成员身份、审批资格和操作审计由业务负责。并发任务不要共享可修改的分页结果或原始事件载荷。

## 另请参阅

- [官方获取群成员列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members.get.html)
- [群聊管理概述](README.md)
- [入群审批](join-requests.md)
