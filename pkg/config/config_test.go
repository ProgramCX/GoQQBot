package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadModesAndCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, extra, mode string
		workers           int
	}{
		{"legacy", "server:\n  pool:\n    workers: 3\n    queueSize: 7\n", "webhook", 3},
		{"websocket", "connection:\n  mode: websocket\n  websocket:\n    intents: [GROUP_AND_C2C_EVENT, INTERACTION]\n    dialTimeout: 2s\nruntime:\n  pool:\n    workers: 4\n", "websocket", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte("bot:\n  appId: test\n  appSecret: example-secret\n"+tc.extra), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Connection.Mode != tc.mode || cfg.Runtime.Pool.Workers != tc.workers {
				t.Fatalf("unexpected config: mode=%s workers=%d", cfg.Connection.Mode, cfg.Runtime.Pool.Workers)
			}
		})
	}
}

func TestInvalidSettings(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(c *Config) { c.Connection.Mode = "invalid" },
		func(c *Config) { c.Connection.WebSocket.Intents = []string{"UNKNOWN"} },
		func(c *Config) { c.Connection.WebSocket.Shard.ID = 1 },
		func(c *Config) { c.Runtime.Pool.QueueSize = -1 },
		func(c *Config) { c.Connection.WebSocket.Reconnect.InitialDelay = time.Minute },
		func(c *Config) { c.Bot.AppSecret = "" },
	} {
		cfg := DefaultConfig()
		cfg.Bot = Bot{AppId: "test", AppSecret: "secret"}
		cfg.Connection.Mode = "websocket"
		mutate(cfg)
		if err := cfg.Validate(); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}
func TestConfigCopyAndPoolPrecedence(t *testing.T) {
	cfg := Config{Server: Server{Pool: ServerPool{Workers: 2, QueueSize: 4}}, Runtime: Runtime{Pool: ServerPool{Workers: 7}}, Connection: Connection{WebSocket: WebSocket{Intents: []string{"INTERACTION"}}}}
	copy := cfg.WithDefaults()
	copy.Connection.WebSocket.Intents[0] = "GUILDS"
	if cfg.Connection.WebSocket.Intents[0] != "INTERACTION" {
		t.Fatal("shared intents backing array")
	}
	if copy.Runtime.Pool.Workers != 7 || copy.Runtime.Pool.QueueSize != 4 {
		t.Fatal("incorrect precedence")
	}
}
