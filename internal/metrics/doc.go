// Package metrics exposes lightweight operational counters for driftwatch.
//
// Counters are updated atomically by other packages (watcher, notifier, etc.)
// and served as a JSON payload on GET /metrics, making them easy to scrape
// with any HTTP-capable monitoring system.
//
// Usage:
//
//	m := metrics.New()
//	m.Counters().DriftsDetected.Add(1)
//	http.Handle("/metrics", m.Handler())
package metrics
