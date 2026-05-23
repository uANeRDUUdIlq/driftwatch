// Package circuitbreaker provides a lightweight circuit breaker for
// protecting outbound alert deliveries in driftwatch.
//
// When a webhook or Slack endpoint fails repeatedly, the breaker opens
// and subsequent attempts are rejected immediately until a configurable
// cooldown period elapses. After cooldown, a single probe request is
// allowed (half-open state); success closes the circuit, failure
// re-opens it.
//
// Usage:
//
//	br := circuitbreaker.New(5, 30*time.Second)
//	if err := br.Allow(); err != nil {
//	    // skip delivery
//	}
//	// ... attempt delivery ...
//	br.RecordSuccess() // or br.RecordFailure()
package circuitbreaker
