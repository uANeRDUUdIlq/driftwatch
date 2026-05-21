package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return p
}

func TestLoad_Valid(t *testing.T) {
	raw := `
watch_paths:
  - /etc/app/config.yaml
  - /etc/app/secrets.yaml
interval: 1m
alerts:
  webhook_url: https://hooks.example.com/abc123
`
	cfg, err := config.Load(writeTemp(t, raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.WatchPaths) != 2 {
		t.Errorf("expected 2 watch paths, got %d", len(cfg.WatchPaths))
	}
	if cfg.Interval != time.Minute {
		t.Errorf("expected interval 1m, got %v", cfg.Interval)
	}
	if cfg.Alerts.WebhookURL != "https://hooks.example.com/abc123" {
		t.Errorf("unexpected webhook URL: %s", cfg.Alerts.WebhookURL)
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	raw := `
watch_paths:
  - /etc/app/config.yaml
alerts:
  webhook_url: https://hooks.example.com/abc123
`
	cfg, err := config.Load(writeTemp(t, raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", cfg.Interval)
	}
}

func TestLoad_MissingWatchPaths(t *testing.T) {
	raw := `
watch_paths: []
alerts:
  webhook_url: https://hooks.example.com/abc123
`
	_, err := config.Load(writeTemp(t, raw))
	if err == nil {
		t.Fatal("expected error for empty watch_paths, got nil")
	}
}

func TestLoad_MissingAlerts(t *testing.T) {
	raw := `
watch_paths:
  - /etc/app/config.yaml
`
	_, err := config.Load(writeTemp(t, raw))
	if err == nil {
		t.Fatal("expected error when no alert target configured, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
