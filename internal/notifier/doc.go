// Package notifier provides a unified notification pipeline for driftwatch.
//
// It wires together rate limiting, maintenance-window suppression, debouncing,
// alerting, and audit logging so that callers need only call Send with a drift
// Event.  Each stage is optional; pass nil to skip it.
//
// Typical usage:
//
//	n := notifier.New(notifier.Config{
//		Alerter:     myAlerter,
//		Limiter:     myLimiter,
//		Suppression: myStore,
//		Debouncer:   myDebouncer,
//		Auditor:     myAuditor,
//	})
//	n.Send(ctx, notifier.Event{Path: "/etc/nginx/nginx.conf", ...})
package notifier
