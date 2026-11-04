# 测试消息发送

向 SDK 提供测试用的 HTTP 客户端，可以在本地核对发送请求。下面的测试检查群聊接口路径、被回复的消息 ID 和回复正文。

将示例保存为业务项目的 `reply_test.go`，运行 `go test ./...`。

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

SDK 用同一个 HTTP 客户端取令牌和发消息。测试客户端要模拟这两种响应。

## 测试分层

本地测试可验证请求的构造方式。若要检查 QQ 客户端的显示效果，请使用真实机器人运行[运行示例程序](../examples/README.md)中的示例。

| 层次 | 验证内容 | 是否需要机器人 |
| --- | --- | --- |
| 消息构造 | `Build(scene)` 校验与场景适配 | 否 |
| API 请求 | 路径、正文、响应和错误传播 | 否，使用替身 |
| 事件接入 | 网关 Hello、认证、心跳、恢复或 Webhook 验签 | 否，可用本地服务器 |
| QQ 客户端测试 | 按钮行为、媒体显示、移动端兼容、平台权限 | 是 |

仓库的常规检查是 `go test ./...` 和 `go vet ./...`。

## 注意事项

- 单元测试使用测试凭证。示例中的临时令牌和返回 ID 均为测试数据。
- SDK 不会关闭注入客户端管理的资源。调用方负责资源回收。
- Windows 上运行 `go test -race ./...` 需要可用的 C 编译器和 CGO。
- 记录测试结果时，区分模拟请求和 QQ 客户端实测。

## 另请参阅

[运行示例程序](../examples/README.md) · [消息类型](../server-api/messages/types/README.md) · [业务接入](business-integration.md)
