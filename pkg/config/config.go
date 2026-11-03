// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

// Package config loads and validates bot configuration. Loaded values are owned by the caller.
// config 包加载并校验机器人配置，加载结果由调用方持有。
package config

import (
	"fmt"
	"time"
)

// Config configures credentials, transport and bounded event processing.
// Config 配置凭证、传输方式及有界事件处理。
type Config struct {
	Server     Server     `yaml:"server"`
	Bot        Bot        `yaml:"bot"`
	Connection Connection `yaml:"connection"`
	Runtime    Runtime    `yaml:"runtime"`
}

// Connection selects exactly one event transport. Empty Mode defaults to webhook.
// Connection 选择唯一的事件传输方式，Mode 为空时默认使用 Webhook。
type Connection struct {
	Mode      string    `yaml:"mode"`
	WebSocket WebSocket `yaml:"websocket"`
}

// WebSocket configures a single gateway connection. Intents use official names.
// WebSocket 配置单个网关连接，Intents 使用官方名称。
type WebSocket struct {
	Intents     []string      `yaml:"intents"`
	DialTimeout time.Duration `yaml:"dialTimeout"`
	Reconnect   Reconnect     `yaml:"reconnect"`
	Shard       Shard         `yaml:"shard"`
}

// Reconnect bounds exponential reconnect delays; jitter is applied internally.
// Reconnect 限制指数重连延迟，内部加入随机抖动。
type Reconnect struct {
	InitialDelay time.Duration `yaml:"initialDelay"`
	MaxDelay     time.Duration `yaml:"maxDelay"`
}

// Shard selects one connection; ID must be in [0, Total).
// Shard 选择一个连接，ID 必须位于 [0, Total) 范围内。
type Shard struct {
	ID    int `yaml:"id"`
	Total int `yaml:"total"`
}

// Runtime controls shared event processing and the maximum shutdown wait.
// Runtime 控制共享事件处理及最大关闭等待时间。
type Runtime struct {
	Pool            ServerPool    `yaml:"pool"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout"`
}

// Server configures the webhook listener only.
// Server 仅配置 Webhook 监听器。
type Server struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
	// Pool is the legacy pool configuration. Runtime.Pool takes precedence.
	// Pool 是旧版线程池配置，Runtime.Pool 优先。
	Pool ServerPool `yaml:"pool"`
}

// ServerPool bounds worker concurrency and pending events.
// ServerPool 限制工作协程并发数及待处理事件数量。
type ServerPool struct {
	Workers   int `yaml:"workers"`
	QueueSize int `yaml:"queueSize"`
}

// Bot contains application credentials.
// Bot 包含应用凭证。
type Bot struct {
	AppId string `yaml:"appId"`
	// Token is ignored. Deprecated: use AppSecret and access-token authentication.
	// Token 被忽略。Deprecated: 请使用 AppSecret 和访问令牌认证。
	Token     string `yaml:"token"`
	AppSecret string `yaml:"appSecret"`
}

// WithDefaults returns an independent normalized copy. Zero optional values select defaults.
// WithDefaults 返回独立且规范化的副本；可选值为零时使用默认值。
func (c Config) WithDefaults() Config {
	c.Connection.WebSocket.Intents = append([]string(nil), c.Connection.WebSocket.Intents...)
	if c.Connection.Mode == "" {
		c.Connection.Mode = "webhook"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Runtime.Pool.Workers == 0 {
		c.Runtime.Pool.Workers = c.Server.Pool.Workers
	}
	if c.Runtime.Pool.QueueSize == 0 {
		c.Runtime.Pool.QueueSize = c.Server.Pool.QueueSize
	}
	if c.Runtime.Pool.Workers == 0 {
		c.Runtime.Pool.Workers = 10
	}
	if c.Runtime.Pool.QueueSize == 0 {
		c.Runtime.Pool.QueueSize = 100
	}
	if c.Runtime.ShutdownTimeout == 0 {
		c.Runtime.ShutdownTimeout = 15 * time.Second
	}
	w := &c.Connection.WebSocket
	if len(w.Intents) == 0 {
		w.Intents = []string{"GROUP_AND_C2C_EVENT"}
	}
	if w.DialTimeout == 0 {
		w.DialTimeout = 10 * time.Second
	}
	if w.Reconnect.InitialDelay == 0 {
		w.Reconnect.InitialDelay = time.Second
	}
	if w.Reconnect.MaxDelay == 0 {
		w.Reconnect.MaxDelay = 30 * time.Second
	}
	if w.Shard.Total == 0 {
		w.Shard.Total = 1
	}
	return c
}

// Validate checks the normalized configuration without modifying it.
// Validate 检查规范化后的配置，不修改配置。
func (c *Config) Validate() error {
	v := c.WithDefaults()
	if v.Bot.AppId == "" {
		return fmt.Errorf("bot.appId is required")
	}
	if v.Bot.AppSecret == "" {
		return fmt.Errorf("bot.appSecret is required")
	}
	if v.Runtime.Pool.Workers <= 0 {
		return fmt.Errorf("runtime.pool.workers must be positive")
	}
	if v.Runtime.Pool.QueueSize <= 0 {
		return fmt.Errorf("runtime.pool.queueSize must be positive")
	}
	if v.Runtime.ShutdownTimeout <= 0 {
		return fmt.Errorf("runtime.shutdownTimeout must be positive")
	}
	switch v.Connection.Mode {
	case "webhook":
		if v.Server.Port <= 0 || v.Server.Port > 65535 {
			return fmt.Errorf("server.port must be in 1..65535")
		}
	case "websocket":
		w := v.Connection.WebSocket
		if w.DialTimeout <= 0 {
			return fmt.Errorf("connection.websocket.dialTimeout must be positive")
		}
		if w.Reconnect.InitialDelay <= 0 || w.Reconnect.MaxDelay < w.Reconnect.InitialDelay {
			return fmt.Errorf("connection.websocket.reconnect requires 0 < initialDelay <= maxDelay")
		}
		if w.Shard.Total <= 0 || w.Shard.ID < 0 || w.Shard.ID >= w.Shard.Total {
			return fmt.Errorf("connection.websocket.shard requires 0 <= id < total")
		}
		if _, err := IntentMask(w.Intents); err != nil {
			return err
		}
	default:
		return fmt.Errorf("connection.mode must be webhook or websocket")
	}
	return nil
}

// IntentMask combines official intent names; unknown names return an error.
// IntentMask 合并官方 intent 名称，未知名称返回错误。
func IntentMask(names []string) (uint32, error) {
	bits := map[string]uint32{
		"GUILDS": 1 << 0, "GUILD_MEMBERS": 1 << 1, "GUILD_MESSAGES": 1 << 9,
		"GUILD_MESSAGE_REACTIONS": 1 << 10, "DIRECT_MESSAGE": 1 << 12,
		"GROUP_AND_C2C_EVENT": 1 << 25, "INTERACTION": 1 << 26, "MESSAGE_AUDIT": 1 << 27,
		"FORUMS_EVENT": 1 << 28, "AUDIO_ACTION": 1 << 29, "PUBLIC_GUILD_MESSAGES": 1 << 30,
	}
	var mask uint32
	for _, name := range names {
		bit, ok := bits[name]
		if !ok {
			return 0, fmt.Errorf("connection.websocket.intents: unknown name %q", name)
		}
		mask |= bit
	}
	return mask, nil
}

// DefaultConfig returns independent defaults; credentials still need to be supplied.
// DefaultConfig 返回独立的默认配置，仍需提供凭证。
func DefaultConfig() *Config { c := (Config{}).WithDefaults(); return &c }
