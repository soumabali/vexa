package wireguard

import (
	"context"
	"log"
	"sync"
	"time"
)

type StatsCollector struct {
	repo     *Repository
	interval time.Duration
	stopCh   chan struct{}
	running  bool
	mu       sync.Mutex
}

func NewStatsCollector(repo *Repository) *StatsCollector {
	return &StatsCollector{
		repo:     repo,
		interval: 30 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

func (sc *StatsCollector) Start() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.running {
		return
	}

	sc.running = true
	sc.stopCh = make(chan struct{})
	go sc.loop()

	log.Printf("[wg-stats] stats collector started (interval=%s)", sc.interval)
}

func (sc *StatsCollector) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.running {
		return
	}

	close(sc.stopCh)
	sc.running = false
	log.Printf("[wg-stats] stats collector stopped")
}

func (sc *StatsCollector) SetInterval(d time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.interval = d
}

func (sc *StatsCollector) IsRunning() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.running
}

func (sc *StatsCollector) loop() {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-sc.stopCh:
			return
		case <-ticker.C:
			sc.collect()
		}
	}
}

func (sc *StatsCollector) collect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tunnels, err := sc.repo.ListAll(ctx, 1000, 0)
	if err != nil {
		log.Printf("[wg-stats] failed to list tunnels: %v", err)
		return
	}

	for _, tunnel := range tunnels {
		if !tunnel.IsEnabled {
			continue
		}

		// stats are collected from wg show via interface controller
		// but we should not have a dependency on InterfaceController here
		// The main service will call this periodically to update DB
		_ = tunnel
	}
}
