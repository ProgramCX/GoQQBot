// Package apiwire 为 SDK 内部的领域客户端提供 HTTP 请求辅助方法。
package apiwire

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Path 将标识作为独立路径段编码。
func Path(format string, ids ...string) (string, error) {
	parts := make([]any, len(ids))
	for i, id := range ids {
		if strings.TrimSpace(id) == "" || id == "." || id == ".." {
			return "", fmt.Errorf("invalid path identifier at position %d", i+1)
		}
		parts[i] = url.PathEscape(id)
	}
	return fmt.Sprintf(format, parts...), nil
}

// Requester 是领域客户端所需的认证请求入口。
type Requester interface {
	Request(context.Context, string, string, any) ([]byte, error)
}

// Do 调用认证接口并解析 JSON 响应。
func Do[T any](ctx context.Context, requester Requester, method, path string, body any) (*T, error) {
	data, err := requester.Request(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// Exec 调用无需解析响应正文的认证接口。
func Exec(ctx context.Context, requester Requester, method, path string, body any) error {
	_, err := requester.Request(ctx, method, path, body)
	return err
}
