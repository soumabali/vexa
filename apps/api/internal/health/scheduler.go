package health

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Scheduler manages periodic health checks for all registered hosts
type Scheduler struct {
	mu       sync.RWMutex
	engine   *ProbeEngine
	db       *sql.DB
	interval time.Duration
	ticker   *time.Ticker
	cancel   context.CancelFunc
	running  bool
}

// NewScheduler creates a health check scheduler
func NewScheduler(db *sql.DB, engine *ProbeEngine) *Scheduler {
	return &Scheduler{
		db:       db,
		engine:   engine,
		interval: 5 * time.Minute,
	}
}

// WithInterval sets check interval
func (s *Scheduler) WithInterval(interval time.Duration) *Scheduler {
	s.interval = interval
	return s
}

// Start begins the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.ticker = time.NewTicker(s.interval)
	s.running = true

	// Initial load
	s.loadHosts(ctx)

	go s.run(ctx)
	return nil
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.running = false
}

// IsRunning returns scheduler status
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// run main scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	defer s.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.ticker.C:
			s.loadHosts(ctx)
		}
	}
}

// loadHosts fetches active hosts and registers them with engine
func (s *Scheduler) loadHosts(ctx context.Context) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, address, COALESCE(port, 22)
		FROM hosts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, address string
		var port int
		if err := rows.Scan(&id, &name, &address, &port); err != nil {
			continue
		}

		// Only register if not already registered
		if _, ok := s.engine.GetHealth(id); !ok {
			config := DefaultProbeConfig()
			config.Port = port
			config.Interval = s.interval
			s.engine.Register(id, name, address, config)
		}
	}
}

// ManualCheck triggers immediate check for a host
func (s *Scheduler) ManualCheck(hostID string) (*HostHealth, error) {
	health, ok := s.engine.GetHealth(hostID)
	if !ok {
		return nil, fmt.Errorf("host not registered")
	}

	// Force immediate check
	config := DefaultProbeConfig()
	config.Port = 22 // Default, should be from host config
	go s.engine.executeCheck(hostID, health.Address, config)

	// Return current state (will be updated asynchronously)
	return health, nil
}
