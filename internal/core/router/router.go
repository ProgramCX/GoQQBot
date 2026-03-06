// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package router

import (
	"encoding/json"
	"strings"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/internal/foundation/message"
	typesShared "github.com/ProgramCX/GoQQBot/shared/types"
)

type Handler func(ctx *typesShared.Context) error
type CommandRouter struct {
	handlers       map[RoutePair]Handler
	defaultHandler Handler
}

type RoutePair struct {
	cmd     string
	eventId message.EventId
}

func New() *CommandRouter {
	return &CommandRouter{
		handlers: make(map[RoutePair]Handler),
	}
}

func (c *CommandRouter) OnCommand(cmd string, eventId message.EventId, handler Handler) {

	cmd = strings.TrimPrefix(cmd, "/")
	cr := RoutePair{
		cmd:     cmd,
		eventId: eventId,
	}

	if handler == nil {
		logger.ERROR("Error in register command: %s, %s", cmd, eventId)
	}

	c.handlers[cr] = handler
	logger.INFO("Registered Command: %s, %s", cmd, eventId)
}

func (c *CommandRouter) OnEvent(eventId message.EventId, handler Handler) {
	cr := RoutePair{
		cmd:     "",
		eventId: eventId,
	}
	if handler == nil {
		logger.ERROR("Error in register event: %s", eventId)
	}
	c.handlers[cr] = handler
	logger.INFO("Registered Event: %s", eventId)
}

func (c *CommandRouter) Handle(task *typesShared.Task) {
	var msg struct {
		Content string `json:"content"`
	}
	err := json.Unmarshal(task.Payload.D, &msg)
	if err != nil {
		logger.INFO("Not A Valid Json Object: %v", err)
		return
	}

	//See:https://bot.q.qq.com/wiki/develop/api-v2/server-inter/message/send-receive/event.html
	content := strings.TrimSpace(msg.Content)

	if !strings.HasPrefix(content, "/") && strings.TrimSpace(content) != "" {
		logger.DEBUG("Message does not have prefix '/', use default handler. Context: %s", task.Payload)

		if c.defaultHandler == nil {
			logger.ERROR("DefaultHandler is not set, handler discarded. %s", task.Payload)
			return
		}

		ctx := typesShared.Context{
			Payload:    task.Payload,
			Command:    "",
			RawMessage: content,
			Args:       make([]string, 0),
		}

		if err := c.defaultHandler(&ctx); err != nil {
			logger.ERROR("handle default message %s error: %v", content, err)
		}
		return
	}

	cmd, args := parseCommand(content)

	//if cmd == "" {
	//	logger.WARN("Empty Command")
	//	return
	//}

	pr := RoutePair{
		cmd:     cmd,
		eventId: task.Payload.T,
	}

	handler, ok := c.handlers[pr]

	if !ok {
		logger.WARN("Unknown command: /%s (event: %s)", cmd, pr.eventId)
		logger.DEBUG("Exisiting handlers: %+v", c.handlers)
		return
	}

	//task.Ctx.Command = cmd
	//task.Ctx.Args = args
	//task.Ctx.RawMessage = content

	ctx := typesShared.Context{
		Command:    cmd,
		Args:       args,
		RawMessage: content,
		Payload:    task.Payload,
	}

	if err := handler(&ctx); err != nil {
		logger.ERROR("Command /%s error: %v", cmd, err)
		return
	}
}

func parseCommand(content string) (string, []string) {
	content = strings.TrimPrefix(content, "/")

	var args []string
	var current strings.Builder
	inQuote := false

	for _, ch := range content {
		switch ch {
		case '"':
			inQuote = !inQuote
		case ' ', '\t':
			if !inQuote && current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			} else if inQuote {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	if len(args) == 0 {
		return "", nil
	}

	return args[0], args[1:]
}
