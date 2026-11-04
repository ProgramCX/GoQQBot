# 查询和管理群成员

用 `Client.Get` 查询一页群成员。SDK 尚无 `ListGroupMembers` 一类专用方法。

## 查询一页成员

下面的函数每次查一页。要继续查，把上一页返回的 `next_cursor` 传进去。`next_cursor` 为空时，表示已经查到最后一页。

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

## 其他成员操作

下列接口需要直接调用。请求用到哪些字段，可以对照官方文档定义结构体。

| 官方能力 | 请求路径 |
| --- | --- |
| [获取群成员信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members_member_openid.get.html) | `GET /v2/groups/{group_openid}/members/{member_openid}` |
| [群成员批量移除](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_batch_remove_members.post.html) | `POST /v2/groups/{group_openid}/batch_remove_members` |
| [群黑名单查询](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.get.html) | `GET /v2/groups/{group_openid}/member_blacklist` |
| [群黑名单操作](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.post.html) | `POST /v2/groups/{group_openid}/member_blacklist` |

应用负责核对成员身份和审批资格，也要记录执行过的操作。调用方式见[原始 API 扩展](../../advanced/raw-api.md)。

## 注意事项

- 官方当前规定每页最多 30 条。示例每次查询一页，不自动翻页，也不修改成员。
- 接口仍受内邀或白名单限制。循环重试无法获得权限，遇到权限错误先确认平台资格。
- 只更换查询示例的 HTTP 方法，无法构成移除成员的请求。
- 多个协程同时读写同一份分页结果，可能出现数据竞争。共享原始事件数据时也是如此。

## 另请参阅

[官方获取群成员列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members.get.html) · [群聊管理概述](README.md) · [入群审批](join-requests.md)
