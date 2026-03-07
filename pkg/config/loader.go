// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package config

import (
	"os"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFile = "config.yaml"
)

func LoadConfig(path ...string) (*Config, error) {
	configPath := DefaultConfigFile
	if len(path) > 0 && path[0] != "" {
		configPath = path[0]
	}
	data, err := os.ReadFile(configPath)
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
