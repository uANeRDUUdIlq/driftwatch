// Package debounce provides per-key event debouncing for driftwatch.
//
// When infrastructure files change rapidly (e.g. during a batch deploy),
// a debouncer ensures that drift alerts are not fired for every intermediate
// write. Instead, the alert fires once the file has been stable for the
// configured quiet period.
//
// Usage:
//
//	db := debounce.New(500*time.Millisecond, func(path string) {
//		// called once per path after writes settle
//		checkDrift(path)
//	})
//	db.Trigger("/etc/nginx/nginx.conf")
package debounce
