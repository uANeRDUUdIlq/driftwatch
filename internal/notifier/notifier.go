// Package notifier provides a unified notification pipeline that combines
// rate limiting, suppression, debouncing, and alerting into a single Send call.
package notifier

import (
	"context"
	"log"
	"time"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/audit"
	"github.com/example/driftwatch/internal/debounce"
	"github.com/example/driftwatch/internal/ratelimit"
	"github.com/example/driftwatch/internal/suppression"
)

// Event represents a drift detection event to be notified.
type Event struct {
	Path      string
	OldHash   string
	NewHash   string
	DetectedAt time.Time
}

// Notifier orchestrates the notification pipeline.
type Notifier struct {
	alerter    *alert.Alerter
	rl         *ratelimit.Limiter
	sup        *suppression.Store
	deb        *debounce.Debouncer
	auditor    *audit.Logger
	debounceWait time.Duration
}

// Config holds dependencies and tuning parameters for Notifier.
type Config struct {
	Alerter      *alert.Alerter
	Limiter      *ratelimit.Limiter
	Suppression  *suppression.Store
	Debouncer    *debounce.Debouncer
	Auditor      *audit.Logger
	DebounceWait time.Duration
}

// New creates a Notifier from the provided Config.
func New(cfg Config) *Notifier {
	if cfg.DebounceWait == 0 {
		cfg.DebounceWait = 2 * time.Second
	}
	return &Notifier{
		alerter:      cfg.Alerter,
		rl:           cfg.Limiter,
		sup:          cfg.Suppression,
		deb:          cfg.Debouncer,
		auditor:      cfg.Auditor,
		debounceWait: cfg.DebounceWait,
	}
}

// Send processes an Event through the full notification pipeline.
// It returns false if the event was suppressed, rate-limited, or debounced.
func (n *Notifier) Send(ctx context.Context, ev Event) bool {
	if n.sup != nil && n.sup.IsSuppressed(ev.Path) {
		log.Printf("notifier: suppressed event for %s", ev.Path)
		return false
	}

	if n.rl != nil && !n.rl.Allow(ev.Path) {
		log.Printf("notifier: rate-limited event for %s", ev.Path)
		return false
	}

	if n.deb != nil {
		n.deb.Trigger(ev.Path, n.debounceWait, func() {
			n.dispatch(ctx, ev)
		})
		return true
	}

	n.dispatch(ctx, ev)
	return true
}

func (n *Notifier) dispatch(ctx context.Context, ev Event) {
	aev := alert.Event{
		Path:    ev.Path,
		OldHash: ev.OldHash,
		NewHash: ev.NewHash,
	}
	if err := n.alerter.Send(ctx, aev); err != nil {
		log.Printf("notifier: alert send error for %s: %v", ev.Path, err)
	}
	if n.auditor != nil {
		n.auditor.Record(audit.Entry{
			Path:    ev.Path,
			OldHash: ev.OldHash,
			NewHash: ev.NewHash,
		})
	}
}
