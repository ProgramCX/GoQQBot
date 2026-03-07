// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package config

import "fmt"

type Config struct {
	Server Server `yaml:"server"`
	Bot    Bot    `yaml:"bot"`
}

type Server struct {
	Port int        `yaml:"port"`
	Host string     `yaml:"host"`
	Pool ServerPool `yaml:"pool"`
}

type ServerPool struct {
	Workers   int `yaml:"workers"`
	QueueSize int `yaml:"queueSize"`
}

type Bot struct {
	AppId     string `yaml:"appId"`
	Token     string `yaml:"token"`
	AppSecret string `yaml:"appSecret"`
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}
	if c.Server.Pool.Workers <= 0 {
		return fmt.Errorf("invalid workers count: %d", c.Server.Pool.Workers)
	}
	if c.Server.Pool.QueueSize <= 0 {
		return fmt.Errorf("invalid queue size: %d", c.Server.Pool.QueueSize)
	}
	if c.Bot.AppSecret == "" {
		return fmt.Errorf("appSecret is required")
	}
	if c.Bot.Token == "" {
		return fmt.Errorf("bot token is required")
	}
	if c.Bot.AppId == "" {
		return fmt.Errorf("appId is required")
	}
	return nil
}

func DefaultConfig() *Config {
	csPool := ServerPool{
		Workers:   10,
		QueueSize: 1000,
	}
	cBot := Bot{
		AppId:     "",
		Token:     "",
		AppSecret: "",
	}
	cServer := Server{
		Port: 8080,
		Host: "0.0.0.0",
		Pool: csPool,
	}
	return &Config{
		Server: cServer,
		Bot:    cBot,
	}
}
