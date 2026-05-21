// Package ratelimit implements a per-key token-bucket rate limiter used
// by driftwatch to suppress repeated alerts for the same drifted file.
//
// When a configuration file drifts repeatedly within a short window —
// for example due to a flapping deployment — the limiter prevents alert
// fatigue by capping the number of notifications sent per key (typically
// the file path) within a configurable time window.
//
// Usage:
//
//	limiter := ratelimit.New(5, time.Minute)
//	if limiter.Allow(filePath) {
//		// send alert
//	}
package ratelimit
