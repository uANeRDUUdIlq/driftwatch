// Package main is the entry point for the driftwatch daemon.
// It wires together configuration, snapshot monitoring, alerting,
// and digest reporting into a single long-running process.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/config"
	"github.com/yourorg/driftwatch/internal/reporter"
	"github.com/yourorg/driftwatch/internal/watcher"
)

func main() {
	cfgPath := flag.String("config", "driftwatch.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration from the specified YAML file.
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("driftwatch: failed to load config: %v", err)
	}

	// Build the alert sender from configured targets.
	alerter, err := alert.New(cfg)
	if err != nil {
		log.Fatalf("driftwatch: failed to initialise alerter: %v", err)
	}

	// Build the digest reporter; it batches drift events and flushes
	// a summary on a configurable interval.
	rep := reporter.New(alerter, reporter.Options{
		FlushInterval: time.Duration(cfg.DigestIntervalSecs) * time.Second,
	})

	// Build the file watcher; it polls each watched path and emits
	// drift events when content changes.
	w, err := watcher.New(cfg, rep)
	if err != nil {
		log.Fatalf("driftwatch: failed to initialise watcher: %v", err)
	}

	// Root context — cancelled on SIGINT / SIGTERM so every component
	// can shut down gracefully.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	log.Printf("driftwatch: starting — watching %d path(s), poll interval %ds",
		len(cfg.WatchPaths), cfg.IntervalSecs)

	// Start the reporter's background flush loop.
	go rep.Run(ctx)

	// Run the watcher; blocks until ctx is cancelled.
	if err := w.Run(ctx); err != nil {
		log.Printf("driftwatch: watcher exited with error: %v", err)
	}

	// Give the reporter a moment to flush any buffered events before exit.
	log.Println("driftwatch: shutting down — flushing digest")
	rep.Flush()
	log.Println("driftwatch: bye")
}
