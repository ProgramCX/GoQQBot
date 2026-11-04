// Package channel 提供 QQ 频道管理和消息接口。
package channel

import "github.com/ProgramCX/GoQQBot/pkg/api"

// Client 使用现有 API 客户端调用频道接口，共用其认证和 HTTP 设置。
type Client struct {
	api *api.Client
}

// NewClient 创建频道客户端，不发送网络请求。
func NewClient(client *api.Client) *Client {
	return &Client{api: client}
}

func (c *Client) requester() *api.Client {
	if c == nil {
		return nil
	}
	return c.api
}
