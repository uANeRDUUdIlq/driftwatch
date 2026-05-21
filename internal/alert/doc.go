// Package alert provides notification delivery for driftwatch.
//
// It supports sending DriftEvent payloads to generic webhooks and
// Slack incoming webhook URLs. Multiple destinations can be configured
// simultaneously; errors from each are collected and returned together.
//
// Example usage:
//
//	n := alert.New(cfg.Alerts.WebhookURL, cfg.Alerts.SlackURL)
//	err := n.Send(alert.DriftEvent{
//		Path:       "/etc/myapp/config.yaml",
//		OldHash:    previousHash,
//		NewHash:    currentHash,
//		DetectedAt: time.Now(),
//	})
package alert
