package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DriftEvent represents a detected configuration drift.
type DriftEvent struct {
	Path      string    `json:"path"`
	OldHash   string    `json:"old_hash"`
	NewHash   string    `json:"new_hash"`
	DetectedAt time.Time `json:"detected_at"`
}

// Notifier sends drift alerts to configured destinations.
type Notifier struct {
	webhookURL string
	slackURL   string
	client     *http.Client
}

// New creates a new Notifier with the given webhook and Slack URLs.
func New(webhookURL, slackURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		slackURL:   slackURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches a DriftEvent to all configured alert destinations.
func (n *Notifier) Send(event DriftEvent) error {
	var errs []error

	if n.webhookURL != "" {
		if err := n.postJSON(n.webhookURL, event); err != nil {
			errs = append(errs, fmt.Errorf("webhook: %w", err))
		}
	}

	if n.slackURL != "" {
		payload := map[string]string{
			"text": fmt.Sprintf(":warning: Drift detected in `%s`\nOld: `%s`\nNew: `%s`\nAt: %s",
				event.Path, event.OldHash, event.NewHash, event.DetectedAt.Format(time.RFC3339)),
		}
		if err := n.postJSON(n.slackURL, payload); err != nil {
			errs = append(errs, fmt.Errorf("slack: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("alert errors: %v", errs)
	}
	return nil
}

func (n *Notifier) postJSON(url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	resp, err := n.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
