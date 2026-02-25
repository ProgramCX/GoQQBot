// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package config

import (
	"fmt"
	"os"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFile = "config.yaml"
)

// LoadConfig load config from config.yaml file
// It returns error if config.yaml file is invalid
func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(DefaultConfigFile)
	if err != nil {
		logger.ERROR("read config file failed: %v", err)
		return nil, err
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logger.ERROR("unmarshal config file failed: %v", err)
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		logger.ERROR("validate config file failed: %v", err)
		return nil, err
	}

	return &cfg, nil
}

// DefaultConfig return default config
func DefaultConfig() *Config {
	csPool := ServerPool{
		Workers:   10,
		QueueSize: 1000,
	}
	cBot := Bot{
		AppId:     0,
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

// Validate validate config
// It returns error if config is invalid
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

	if c.Bot.AppId == 0 {
		return fmt.Errorf("appId is required")
	}

	return nil
}
