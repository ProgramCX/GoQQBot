// Package group 提供 QQ 群聊消息与管理接口。
package group

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	"github.com/ProgramCX/GoQQBot/pkg/api"
)

// Client 使用已有 API 客户端调用群聊管理接口，共用其认证和 HTTP 设置。
type Client struct {
	api *api.Client
}

// NewClient 创建群聊管理客户端，不发送网络请求。
func NewClient(client *api.Client) *Client {
	return &Client{api: client}
}

func (c *Client) requester() *api.Client {
	if c == nil {
		return nil
	}
	return c.api
}

const groupPath = "/v2/groups/%s"
const strategyPath = "/v2/groups/join_approval_strategy"

// PageOptions 设置单次分页请求。零值查询首页并使用平台默认的条数。
type PageOptions struct {
	// Cursor 是上一页返回的 NextCursor。
	Cursor string `json:"cursor,omitempty"`
	// Limit 是单页条数。0 表示使用平台默认值。
	Limit int `json:"limit,omitempty"`
}

func (page PageOptions) validate(max int) error {
	if page.Limit < 0 || page.Limit > max {
		return fmt.Errorf("page limit must be between 0 and %d", max)
	}
	return nil
}

func getPage[T any](ctx context.Context, c *Client, path string, page PageOptions, max int) (*T, error) {
	if err := page.validate(max); err != nil {
		return nil, err
	}
	query := url.Values{}
	if page.Cursor != "" {
		query.Set("cursor", page.Cursor)
	}
	if page.Limit > 0 {
		query.Set("limit", strconv.Itoa(page.Limit))
	}
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return apiwire.Do[T](ctx, c.requester(), http.MethodGet, path, nil)
}

func validateBatch(ids []string, max int, field string) error {
	if len(ids) == 0 || len(ids) > max {
		return fmt.Errorf("%s must contain between 1 and %d entries", field, max)
	}
	for _, id := range ids {
		if id == "" {
			return fmt.Errorf("%s contains an empty identifier", field)
		}
	}
	return nil
}

// ListAction 表示向名单或策略关联群列表中增加或删除条目。
type ListAction string

const (
	// ListAdd 向列表增加条目。
	ListAdd ListAction = "add"
	// ListDelete 从列表删除条目。
	ListDelete ListAction = "del"
)
