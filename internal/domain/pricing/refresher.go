// Package pricing provides background pricing refresh functionality
package pricing

import (
	"context"
	"path/filepath"
	"time"

	"github.com/royisme/bobamixer/internal/logger"
	"github.com/royisme/bobamixer/internal/store/config"
)

// Refresher handles background pricing updates
type Refresher struct {
	home     string
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewRefresher creates a new pricing refresher
func NewRefresher(home string, intervalHours int) *Refresher {
	if intervalHours <= 0 {
		intervalHours = 24 // Default to 24 hours
	}

	return &Refresher{
		home:     home,
		interval: time.Duration(intervalHours) * time.Hour,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start starts the background refresh goroutine
func (r *Refresher) Start(ctx context.Context) {
	go r.run(ctx)
}

// Stop stops the background refresh
func (r *Refresher) Stop() {
	close(r.stopCh)
	<-r.doneCh
}

// run is the main refresh loop
func (r *Refresher) run(ctx context.Context) {
	defer close(r.doneCh)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	logger.Info("Pricing refresher started", logger.String("interval", r.interval.String()))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Pricing refresher stopped due to context cancellation")
			return
		case <-r.stopCh:
			logger.Info("Pricing refresher stopped")
			return
		case <-ticker.C:
			if err := r.refresh(); err != nil {
				logger.Error("Failed to refresh pricing", logger.Err(err))
			} else {
				logger.Info("Pricing refreshed successfully")
			}
		}
	}
}

// refresh performs a single pricing refresh
func (r *Refresher) refresh() error {
	// Load pricing config to get sources
	pricingCfg, err := config.LoadPricing(r.home)
	if err != nil {
		return err
	}

	// Fetch from remote sources
	table, err := fetchRemote(pricingCfg.Sources, r.home)
	if err != nil {
		return err
	}

	// Save to cache
	cachePath := filepath.Join(r.home, "pricing.cache.json")
	return saveCache(cachePath, table)
}

// RefreshNow forces an immediate refresh
func (r *Refresher) RefreshNow() error {
	return r.refresh()
}
