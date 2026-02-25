// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

// Package config provides config.yaml file structure
package config

// Config config.yaml file structure
type Config struct {
	Server Server `yaml:"server"`
	Bot    Bot    `yaml:"bot"`
}

// Server server
type Server struct {
	Port int        `yaml:"port"`
	Host string     `yaml:"host"`
	Pool ServerPool `yaml:"pool"`
}

// ServerPool server.pool
type ServerPool struct {
	Workers   int `yaml:"workers"`
	QueueSize int `yaml:"queueSize"`
}

// Bot bot
type Bot struct {
	AppId     int32  `yaml:"appId"`
	Token     string `yaml:"token"`
	AppSecret string `yaml:"appSecret"`
}
