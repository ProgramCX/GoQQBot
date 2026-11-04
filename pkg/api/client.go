// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

// Package api provides concurrency-safe authenticated HTTP access to QQ Bot OpenAPI.
// api 包提供可并发使用且带认证的 QQ 机器人 OpenAPI HTTP 客户端。
// Clients own their token cache; injected HTTP clients remain caller-owned.
// 客户端持有各自的令牌缓存；注入的 HTTP 客户端仍由调用方持有。
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"

	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Current QQ Bot authentication and OpenAPI endpoints.
// 当前 QQ 机器人认证和 OpenAPI 端点。
const (
	TokenURL = "https://api.bot.qq.com/app/getAppAccessToken"
	BaseURL  = "https://api.bot.qq.com"
)

// Client shares token refresh across concurrent API calls. Construct it with NewClient.
// Client 在并发 API 调用间共享令牌刷新，通过 NewClient 构造。
type Client struct {
	appID string

	appSecret   string
	accessToken string
	expiresAt   time.Time
	mu          sync.RWMutex
	httpClient  *http.Client
	refresh     chan struct{}
}

// NewClient constructs a dormant client. Credentials are used only when a request needs a token.
// NewClient 创建尚未发起请求的客户端，仅在请求需要令牌时使用凭证。
func NewClient(appID, appSecret string, options ...ClientOption) *Client {
	c := &Client{
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// ClientOption customizes a client before its first use.
// ClientOption 在客户端首次使用前定制配置。
type ClientOption func(*Client)

// WithHTTPClient uses the caller-owned HTTP client, including for token refresh.
// WithHTTPClient 使用调用方持有的 HTTP 客户端，包括令牌刷新请求。
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// APIError preserves HTTP status and the platform's machine-readable error code.
// APIError 保留 HTTP 状态和平台机器可读错误码。
type APIError struct {
	StatusCode int
	Code       int
	TraceID    string
	// MessageAuditID 是 QQ 返回的消息审核 ID，未进入审核时为空。
	MessageAuditID string
	diagnostic     string
}

// DiagnosticMessage returns the bounded, redacted platform reason, or empty if none was supplied.
// DiagnosticMessage 返回长度受限且脱敏的平台原因；未提供原因则返回空串。
// Credentials known to this client, URLs and likely opaque tokens are removed on capture.
// 捕获时移除客户端已知凭证、URL 及疑似不透明令牌。
// This best-effort diagnostic is omitted from Error; do not treat its text as a stable error code.
// 此诊断为尽力脱敏，不包含在 Error 中；不得将其文本视为稳定错误码。
func (e *APIError) DiagnosticMessage() string {
	if e == nil {
		return ""
	}
	return e.diagnostic
}

// Error omits response content because it may contain sensitive values.
// Error 不包含响应内容，因为其中可能存在敏感值。
func (e *APIError) Error() string {
	return fmt.Sprintf("QQ API failed: HTTP %d, code %d, trace %s", e.StatusCode, e.Code, e.TraceID)
}

// EnsureToken ensures that a token is cached. Prefer AccessToken for cancellable operations.
// EnsureToken 确保令牌已缓存；可取消的操作优先使用 AccessToken。
func (c *Client) EnsureToken() error { _, err := c.AccessToken(context.Background()); return err }

// InvalidateToken expires a rejected cached token without invalidating a newer concurrent refresh.
// InvalidateToken 使被拒绝的缓存令牌过期，不会使并发刷新产生的新令牌失效。
func (c *Client) InvalidateToken(rejected string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.accessToken == rejected {
		c.expiresAt = time.Time{}
	}
}

// AccessToken returns a cached token, refreshing as necessary. Concurrent callers share a refresh.
// AccessToken 返回缓存令牌，必要时刷新；并发调用共享一次刷新。
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		c.mu.Lock()
		if c.accessToken != "" && time.Now().Before(c.expiresAt) {
			token := c.accessToken
			c.mu.Unlock()
			return token, nil
		}
		if pending := c.refresh; pending != nil {
			c.mu.Unlock()
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-pending:
				continue
			}
		}
		pending := make(chan struct{})
		c.refresh = pending
		c.mu.Unlock()
		token, expiry, err := c.fetchToken(ctx)
		c.mu.Lock()
		if err == nil {
			c.accessToken = token
			c.expiresAt = expiry
		}
		c.refresh = nil
		close(pending)
		c.mu.Unlock()
		return token, err
	}
}

func (c *Client) fetchToken(ctx context.Context) (string, time.Time, error) {
	data, err := json.Marshal(map[string]string{"appId": c.appID, "clientSecret": c.appSecret})
	if err != nil {
		return "", time.Time{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, bytes.NewReader(data))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("fetch access token: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string          `json:"access_token"`
		ExpiresIn   json.RawMessage `json:"expires_in"`
		Code        int             `json:"code"`
		Message     string          `json:"message"`
	}
	decodeErr := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result)
	if resp.StatusCode != http.StatusOK || result.Code != 0 {
		return "", time.Time{}, &APIError{StatusCode: resp.StatusCode, Code: result.Code, diagnostic: redactAPIDiagnostic(result.Message, c.appID, c.appSecret, result.AccessToken)}
	}
	if decodeErr != nil {
		return "", time.Time{}, fmt.Errorf("decode access token: %w", decodeErr)
	}
	seconds, err := strconv.ParseInt(strings.Trim(string(result.ExpiresIn), "\""), 10, 64)
	if err != nil || seconds <= 0 || seconds > 86400 || result.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("invalid access-token response")
	}
	return result.AccessToken, time.Now().Add(time.Duration(seconds)*time.Second - min(30*time.Second, time.Duration(seconds)*time.Second/10)), nil
}

// Gateway returns the dynamically discovered secure WebSocket endpoint.
// Gateway 返回动态发现的安全 WebSocket 端点。
func (c *Client) Gateway(ctx context.Context) (string, error) {
	data, err := c.Get(ctx, "/gateway")
	if err != nil {
		return "", fmt.Errorf("discover gateway: %w", err)
	}
	var gateway struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(data, &gateway); err != nil {
		return "", fmt.Errorf("decode gateway: %w", err)
	}
	u, err := url.Parse(gateway.URL)
	if err != nil || u.Scheme != "wss" || u.Host == "" {
		return "", fmt.Errorf("gateway must return a secure WebSocket URL")
	}
	return gateway.URL, nil
}

// Request performs an authenticated JSON request and returns an owned response body.
// Request 执行已认证的 JSON 请求，并返回独立持有的响应正文。
func (c *Client) Request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)

		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}
	return c.request(ctx, method, path, "application/json", bodyReader)
}

// PostMultipart 发送已构造的 multipart 表单，沿用客户端的认证和错误处理。
// 调用方负责结束表单写入，并在请求期间保持表单正文可读。
func (c *Client) PostMultipart(ctx context.Context, path string, form *multipart.Writer, body io.Reader) ([]byte, error) {
	if form == nil || body == nil {
		return nil, fmt.Errorf("multipart form and body are required")
	}
	return c.request(ctx, http.MethodPost, path, form.FormDataContentType(), body)
}

func (c *Client) request(ctx context.Context, method, path, contentType string, bodyReader io.Reader) ([]byte, error) {
	if c == nil || c.httpClient == nil || ctx == nil {
		return nil, fmt.Errorf("API client and operation context are required")
	}
	token, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}

	url := BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "QQBot "+token)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var failure struct {
		Code       int    `json:"err_code"`
		LegacyCode int    `json:"code"`
		TraceID    string `json:"trace_id"`
		Message    string `json:"message"`
		Msg        string `json:"msg"`
		ErrMessage string `json:"err_msg"`
		Data       struct {
			MessageAudit struct {
				ID string `json:"audit_id"`
			} `json:"message_audit"`
		} `json:"data"`
	}
	// Success bodies vary by endpoint, so inspect only the optional error envelope.
	// 成功正文因端点而异，因此仅检查可选的错误信封。
	if len(data) > 0 {
		_ = json.Unmarshal(data, &failure)
	}
	if failure.Code == 0 {
		failure.Code = failure.LegacyCode
	}
	if resp.StatusCode >= 400 || failure.Code != 0 {
		if failure.TraceID == "" {
			failure.TraceID = resp.Header.Get("X-Tps-trace-ID")
		}
		reason := failure.Message
		if reason == "" {
			reason = failure.Msg
		}
		if reason == "" {
			reason = failure.ErrMessage
		}
		apiErr := &APIError{StatusCode: resp.StatusCode, Code: failure.Code, TraceID: failure.TraceID, diagnostic: redactAPIDiagnostic(reason, c.appID, c.appSecret, token)}
		if failure.Code == 304023 || failure.Code == 304024 {
			apiErr.MessageAuditID = failure.Data.MessageAudit.ID
		}
		return nil, apiErr
	}

	return data, nil
}

// Post sends a JSON POST request.
// Post 发送 JSON POST 请求。
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}

// Get sends a GET request.
// Get 发送 GET 请求。
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}

// Delete sends a DELETE request.
// Delete 发送 DELETE 请求。
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodDelete, path, nil)
}

// Put sends a JSON PUT request.
// Put 发送 JSON PUT 请求。
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPut, path, body)
}
