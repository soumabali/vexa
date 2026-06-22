package health

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// ProbeType defines the type of health check
type ProbeType string

const (
	ProbeTypeTCP  ProbeType = "tcp"
	ProbeTypeICMP ProbeType = "icmp"
	ProbeTypeHTTP ProbeType = "http"
)

// ProbeStatus represents the result of a health check
type ProbeStatus string

const (
	StatusHealthy   ProbeStatus = "healthy"
	StatusUnhealthy ProbeStatus = "unhealthy"
	StatusUnknown   ProbeStatus = "unknown"
)

// ProbeConfig configuration for a single probe
type ProbeConfig struct {
	Type     ProbeType     `json:"type"`
	Host     string        `json:"host"`
	Port     int           `json:"port,omitempty"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
	Interval time.Duration `json:"interval"`
}

// ProbeResult result of a single probe execution
type ProbeResult struct {
	Config    ProbeConfig   `json:"config"`
	Status    ProbeStatus   `json:"status"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
	RetryCount int          `json:"retry_count"`
}

// HostHealth aggregated health for a host
type HostHealth struct {
	HostID      string        `json:"host_id"`
	HostName    string        `json:"host_name"`
	Address     string        `json:"address"`
	Status      ProbeStatus   `json:"status"`
	Latency     time.Duration `json:"latency"`
	LastChecked time.Time     `json:"last_checked"`
	LastSuccess time.Time     `json:"last_success,omitempty"`
	LastFailure time.Time     `json:"last_failure,omitempty"`
	ConsecutiveFails int     `json:"consecutive_fails"`
	Error       string        `json:"error,omitempty"`
	History     []ProbeResult `json:"history,omitempty"`
}

// DefaultProbeConfig returns default configuration
func DefaultProbeConfig() ProbeConfig {
	return ProbeConfig{
		Type:     ProbeTypeTCP,
		Timeout:  10 * time.Second,
		Retries:  3,
		Interval: 5 * time.Minute,
	}
}

// ProbeEngine manages health checks
type ProbeEngine struct {
	mu      sync.RWMutex
	checks  map[string]*HostHealth
	results chan ProbeResult
	cancel  context.CancelFunc
}

// NewProbeEngine creates a new health probe engine
func NewProbeEngine() *ProbeEngine {
	return &ProbeEngine{
		checks:  make(map[string]*HostHealth),
		results: make(chan ProbeResult, 100),
	}
}

// Start begins periodic health checks
func (e *ProbeEngine) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	
	go e.processResults(ctx)
}

// Stop halts all health checks
func (e *ProbeEngine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
}

// Register adds a host for periodic checking
func (e *ProbeEngine) Register(hostID, hostName, address string, config ProbeConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.checks[hostID] = &HostHealth{
		HostID:   hostID,
		HostName: hostName,
		Address:  address,
		Status:   StatusUnknown,
		History:  make([]ProbeResult, 0, 24), // Keep 24h of 5-min checks
	}
	
	go e.runPeriodicCheck(hostID, address, config)
}

// Unregister removes a host from checking
func (e *ProbeEngine) Unregister(hostID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.checks, hostID)
}

// GetHealth returns current health for a host
func (e *ProbeEngine) GetHealth(hostID string) (*HostHealth, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	h, ok := e.checks[hostID]
	if !ok {
		return nil, false
	}
	
	// Return copy
	hCopy := *h
	return &hCopy, true
}

// GetAllHealth returns all host health statuses
func (e *ProbeEngine) GetAllHealth() map[string]*HostHealth {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	result := make(map[string]*HostHealth, len(e.checks))
	for k, v := range e.checks {
		vc := *v
		result[k] = &vc
	}
	return result
}

// runPeriodicCheck runs checks at configured interval
func (e *ProbeEngine) runPeriodicCheck(hostID, address string, config ProbeConfig) {
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()
	
	// Initial check
	e.executeCheck(hostID, address, config)
	
	for {
		select {
		case <-ticker.C:
			e.executeCheck(hostID, address, config)
		}
	}
}

// executeCheck runs a single health check
func (e *ProbeEngine) executeCheck(hostID, address string, config ProbeConfig) {
	var result ProbeResult
	var err error
	
	start := time.Now()
	
	for attempt := 0; attempt <= config.Retries; attempt++ {
		result, err = e.runProbe(address, config)
		result.RetryCount = attempt
		result.CheckedAt = time.Now().UTC()
		
		if err == nil && result.Status == StatusHealthy {
			break
		}
		
		if attempt < config.Retries {
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
		}
	}
	
	result.Latency = time.Since(start)
	if err != nil {
		result.Error = err.Error()
		result.Status = StatusUnhealthy
	}
	
	e.updateHealth(hostID, result)
}

// runProbe executes a single probe
func (e *ProbeEngine) runProbe(address string, config ProbeConfig) (ProbeResult, error) {
	result := ProbeResult{
		Config: config,
		Status: StatusUnknown,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	switch config.Type {
	case ProbeTypeTCP:
		return e.tcpProbe(ctx, address, config)
	case ProbeTypeICMP:
		return e.icmpProbe(ctx, address, config)
	case ProbeTypeHTTP:
		return e.httpProbe(ctx, address, config)
	default:
		return result, fmt.Errorf("unknown probe type: %s", config.Type)
	}
}

// updateHealth updates stored health state
func (e *ProbeEngine) updateHealth(hostID string, result ProbeResult) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	h, ok := e.checks[hostID]
	if !ok {
		return
	}
	
	h.LastChecked = result.CheckedAt
	h.Latency = result.Latency
	
	if result.Status == StatusHealthy {
		h.Status = StatusHealthy
		h.LastSuccess = result.CheckedAt
		h.ConsecutiveFails = 0
		h.Error = ""
	} else {
		h.ConsecutiveFails++
		h.LastFailure = result.CheckedAt
		h.Error = result.Error
		
		// Mark unhealthy after 3 consecutive failures
		if h.ConsecutiveFails >= 3 {
			h.Status = StatusUnhealthy
		}
	}
	
	// Maintain rolling history (max 288 entries = 24h @ 5min intervals)
	h.History = append(h.History, result)
	if len(h.History) > 288 {
		h.History = h.History[len(h.History)-288:]
	}
}

// processResults handles async result processing
func (e *ProbeEngine) processResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-e.results:
			// Async processing (metrics, alerts, etc.)
			_ = result
		}
	}
}

// BulkCheck performs health checks on multiple hosts in parallel
func (e *ProbeEngine) BulkCheck(hosts []HostCheckRequest, maxParallel int) []HostHealthResult {
	if maxParallel <= 0 {
		maxParallel = 10
	}
	if maxParallel > 100 {
		maxParallel = 100
	}
	
	semaphore := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	results := make([]HostHealthResult, len(hosts))
	var mu sync.Mutex
	
	for i, host := range hosts {
		wg.Add(1)
		go func(idx int, h HostCheckRequest) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			config := DefaultProbeConfig()
			if h.Port != 0 {
				config.Port = h.Port
			}
			if h.Timeout != 0 {
				config.Timeout = h.Timeout
			}
			
			ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
			defer cancel()
			
			result, err := e.tcpProbe(ctx, h.Address, config)
			if err != nil {
				result.Error = err.Error()
				result.Status = StatusUnhealthy
			}
			
			mu.Lock()
			results[idx] = HostHealthResult{
				HostID:   h.HostID,
				Address:  h.Address,
				Status:   result.Status,
				Latency:  result.Latency,
				Error:    result.Error,
				CheckedAt: result.CheckedAt,
			}
			mu.Unlock()
		}(i, host)
	}
	
	wg.Wait()
	return results
}

// HostCheckRequest request for bulk health check
type HostCheckRequest struct {
	HostID  string        `json:"host_id"`
	Address string        `json:"address"`
	Port    int           `json:"port,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
}

// HostHealthResult result of bulk health check
type HostHealthResult struct {
	HostID    string        `json:"host_id"`
	Address   string        `json:"address"`
	Status    ProbeStatus   `json:"status"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

// tcpProbe performs a TCP connectivity check
func (e *ProbeEngine) tcpProbe(ctx context.Context, address string, config ProbeConfig) (ProbeResult, error) {
	result := ProbeResult{Config: config, Status: StatusUnhealthy}
	
	target := fmt.Sprintf("%s:%d", address, config.Port)
	if config.Port == 0 {
		target = address // Assume address includes port
	}
	
	start := time.Now()
	conn, err := net.DialTimeout("tcp", target, config.Timeout)
	if err != nil {
		return result, err
	}
	defer conn.Close()
	
	result.Latency = time.Since(start)
	result.Status = StatusHealthy
	return result, nil
}

// icmpProbe performs an ICMP latency check
func (e *ProbeEngine) icmpProbe(ctx context.Context, address string, config ProbeConfig) (ProbeResult, error) {
	result := ProbeResult{Config: config, Status: StatusUnhealthy}
	// ICMP requires raw socket privileges (not implemented in basic version)
	// Use TCP fallback or require CAP_NET_RAW
	result.Error = "ICMP requires elevated privileges (CAP_NET_RAW)"
	return result, fmt.Errorf("ICMP not implemented: requires root or CAP_NET_RAW")
}

// httpProbe performs an HTTP health check
func (e *ProbeEngine) httpProbe(ctx context.Context, address string, config ProbeConfig) (ProbeResult, error) {
	result := ProbeResult{Config: config, Status: StatusUnhealthy}
	// HTTP probe: send GET request, check status code
	result.Error = "HTTP probe not yet implemented"
	return result, fmt.Errorf("HTTP probe not implemented")
}
