package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Bootloader BootloaderConfig          `toml:"bootloader"`
	Runlevels  map[string]RunlevelConfig `toml:"runlevels"`
}

type BootloaderConfig struct {
	Version             string `toml:"version"`
	Mode                string `toml:"mode"`
	HeartbeatIntervalMs int    `toml:"heartbeat_interval_ms"`
	LogLevel            string `toml:"log_level"`
}

type RunlevelConfig struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Precheck    map[string]string `toml:"preconditions"`
	Activate    ActivateConfig    `toml:"activate"`
	Health      HealthConfig      `toml:"health"`
	Constraints ConstraintsConfig `toml:"constraints"`
	Advance     AdvanceConfig     `toml:"advance"`
}

type ActivateConfig struct {
	Command   string   `toml:"command"`
	Container string   `toml:"container"`
	DependsOn []string `toml:"depends_on"`
}

type HealthConfig struct {
	GrpcPort         int    `toml:"grpc_port"`
	ContainerRunning bool   `toml:"container_running"`
	CheckEndpoint    string `toml:"check_endpoint"`
	TimeoutMs        int    `toml:"timeout_ms"`
}

type ConstraintsConfig struct {
	RequireAll        bool   `toml:"require_all"`
	Retry             int    `toml:"retry"`
	RetryBackoffMs    int    `toml:"retry_backoff_ms"`
	RollbackOnFailure bool   `toml:"rollback_on_failure"`
	RollbackCommand   string `toml:"rollback_command"`
	EscalateTo        string `toml:"escalate_to"`
}

type AdvanceConfig struct {
	Next int `toml:"next"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode TOML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Runlevels) == 0 {
		return fmt.Errorf("no runlevels defined in configuration")
	}
	return nil
}

func (c *Config) SetDefaults() {
	if c.Bootloader.LogLevel == "" {
		c.Bootloader.LogLevel = "info"
	}
	if c.Bootloader.HeartbeatIntervalMs == 0 {
		c.Bootloader.HeartbeatIntervalMs = 5000
	}
}
