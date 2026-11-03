# 测试消息发送

单元测试应注入 HTTP 替身，不使用真实凭证。下面的完整示例验证回复是否发往群消息路径，并检查被动回复来源及正文。

保存为业务项目的 `reply_test.go`，运行 `go test ./...`。临时令牌和返回 ID 仅为测试数据。

```go
package guide

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGroupReply(t *testing.T) {
	sent := false
	client := api.NewClient("test-app", "test-secret", api.WithHTTPClient(&http.Client{
		Transport: transport(func(r *http.Request) (*http.Response, error) {
			body := `{"access_token":"test-token","expires_in":7200}`
			if r.URL.Path != "/app/getAppAccessToken" {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/groups/group-test/messages" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var payload struct {
					Content string `json:"content"`
					MsgID   string `json:"msg_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				if payload.Content != "你好" || payload.MsgID != "incoming-test" {
					t.Fatal("reply content or source did not match")
				}
				sent = true
				body = `{"id":"reply-test","timestamp":"2026-11-04T12:00:00+08:00"}`
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}))
	result, err := client.SendReply(context.Background(), message.ReplyTarget{
		Scene:     message.SceneGroup,
		ID:        "group-test",
		MessageID: "incoming-test",
	}, message.Text("你好"))
	if err != nil {
		t.Fatal(err)
	}
	if !sent || result.ID != "reply-test" {
		t.Fatal("reply result did not match")
	}
}
```

注入的 HTTP 客户端同时用于令牌刷新和 API 调用，因此替身需要处理两种请求。SDK 不替调用方关闭注入客户端管理的资源。

## 测试分层

| 层次 | 验证内容 | 是否需要机器人 |
| --- | --- | --- |
| 消息构造 | `Build(scene)` 校验与场景适配 | 否 |
| API 请求 | 路径、正文、响应和错误传播 | 否，使用替身 |
| 事件接入 | 网关 Hello、认证、心跳、恢复或 Webhook 验签 | 否，可用本地服务器 |
| 客户端验收 | 按钮行为、媒体显示、移动端兼容、平台权限 | 是 |

本仓库常规检查为 `go test ./...`、`go vet ./...`。Windows 运行 `go test -race ./...` 需要可用的 C 编译器，并启用 CGO；检查未执行时不要在验收记录中标为通过。

## 另请参阅

- [示例程序与验收](../examples/README.md)
- [消息类型](../server-api/messages/types/README.md)
- [业务接入](business-integration.md)
