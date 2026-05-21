package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level driftwatch daemon configuration.
type Config struct {
	WatchPaths []string      `yaml:"watch_paths"`
	Interval   time.Duration `yaml:"interval"`
	Alerts     AlertConfig   `yaml:"alerts"`
}

// AlertConfig holds notification target settings.
type AlertConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	SlackToken string `yaml:"slack_token"`
	SlackChannel string `yaml:"slack_channel"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.WatchPaths) == 0 {
		return fmt.Errorf("watch_paths must contain at least one path")
	}
	if c.Interval <= 0 {
		c.Interval = 30 * time.Second
	}
	hasWebhook := c.Alerts.WebhookURL != ""
	hasSlack := c.Alerts.SlackToken != "" && c.Alerts.SlackChannel != ""
	if !hasWebhook && !hasSlack {
		return fmt.Errorf("at least one alert target (webhook_url or slack) must be configured")
	}
	return nil
}
