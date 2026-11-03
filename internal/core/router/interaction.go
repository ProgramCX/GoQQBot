// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// 本源代码的使用受 MIT 许可证约束，许可证内容见 LICENSE 文件。

package router

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

// OnInteraction registers an exact interaction-kind fallback; nil is ignored.
// OnInteraction 注册指定交互类型的兜底处理函数，忽略 nil。
func (c *CommandRouter) OnInteraction(kind interaction.InteractionType, handler Handler) {
	if handler == nil {
		return
	}
	c.mu.Lock()
	c.interactions[kind] = handler
	c.mu.Unlock()
}

// OnButton registers a message-button callback ID; empty IDs and nil handlers are ignored.
// OnButton 按消息按钮回调 ID 注册；忽略空 ID 和 nil 处理函数。
func (c *CommandRouter) OnButton(id string, handler Handler) {
	if id == "" || handler == nil {
		return
	}
	c.mu.Lock()
	c.buttons[id] = handler
	c.mu.Unlock()
}

// OnMenu registers a shortcut-menu feature ID; empty IDs and nil handlers are ignored.
// OnMenu 按快捷菜单功能 ID 注册；忽略空 ID 和 nil 处理函数。
func (c *CommandRouter) OnMenu(id string, handler Handler) {
	if id == "" || handler == nil {
		return
	}
	c.mu.Lock()
	c.menus[id] = handler
	c.mu.Unlock()
}

func (c *CommandRouter) handleEvent(task *types.Task) {
	ctx := &types.Context{Context: task.Ctx, Payload: task.Payload, Client: task.Client, Args: []string{}}
	var event *interaction.InteractionCreateEvent
	if task.Payload.T == message.INTERACTION_CREATE {
		decoded, err := ctx.Interaction()
		if err != nil {
			c.fail(ctx, "decode", err)
			return
		}
		event = decoded
	}
	c.mu.RLock()
	var handler Handler
	if event != nil {
		switch event.Kind() {
		case interaction.InteractionTypeButton:
			handler = c.buttons[event.Data.Resolved.ButtonID]
		case interaction.InteractionTypeMenu:
			handler = c.menus[event.Data.Resolved.FeatureID]
		}
		if handler == nil {
			handler = c.interactions[event.Kind()]
		}
	}
	if handler == nil {
		handler = c.handlers[RoutePair{eventId: task.Payload.T}]
	}
	c.mu.RUnlock()
	if handler != nil {
		c.invoke(ctx, handler)
	}
}
