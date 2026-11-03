// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

// Package types holds handler contexts and transport envelopes shared by adapters and applications.
// types 包保存适配器与业务应用共用的处理上下文及传输信封。
package types

import (
	"context"
	"sync/atomic"

	"github.com/ProgramCX/GoQQBot/pkg/api"
)

// Context describes one incoming event. Pass its pointer; do not copy it after first use.
// Context 描述一个接收事件，请传递指针，首次使用后不要复制。
// Reply methods support concurrent calls; the payload and public fields must remain read-only during handling.
// 回复方法可并发调用；处理期间载荷及公开字段必须保持只读。
type Context struct {
	replySequence                atomic.Int64
	interactionResponseAttempted atomic.Bool
	// Context is canceled when the bot stops. It may be nil in caller-created contexts.
	// 机器人停止时取消 Context，调用方自行构造的上下文中可能为 nil。
	Context    context.Context
	Payload    *Payload
	Command    string
	Args       []string
	RawMessage string
	Client     *api.Client
}

// RequestContext returns the event context, or Background for legacy caller-created values.
// RequestContext 返回事件上下文，旧版自行构造值则回退到 Background。
func (c *Context) RequestContext() context.Context {
	if c.Context != nil {
		return c.Context
	}
	return context.Background()
}
