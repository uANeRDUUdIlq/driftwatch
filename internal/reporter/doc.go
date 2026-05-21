// Package reporter implements a drift event aggregator that buffers
// DriftEvent values recorded by the watcher and periodically flushes
// a digest to a configurable Sink function.
//
// Typical usage:
//
//	sink := func(events []reporter.DriftEvent) error {
//		// send to Slack, write to log, etc.
//		return nil
//	}
//	r := reporter.New(5*time.Minute, sink)
//	go r.Run(ctx)
//
//	// later, when drift is detected:
//	r.Record(reporter.DriftEvent{
//		Path:       "/etc/app/config.yaml",
//		OldHash:    oldHash,
//		NewHash:    newHash,
//		DetectedAt: time.Now(),
//	})
package reporter
