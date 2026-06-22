package health

import (
	"context"
	"fmt"
	"net"
	"time"
)

// TCPHealthChecker performs TCP connectivity checks
type TCPHealthChecker struct {
	timeout time.Duration
	retries int
}

// NewTCPHealthChecker creates a TCP checker with defaults
func NewTCPHealthChecker() *TCPHealthChecker {
	return &TCPHealthChecker{
		timeout: 10 * time.Second,
		retries: 3,
	}
}

// WithTimeout sets custom timeout
func (c *TCPHealthChecker) WithTimeout(timeout time.Duration) *TCPHealthChecker {
	c.timeout = timeout
	return c
}

// WithRetries sets retry count
func (c *TCPHealthChecker) WithRetries(retries int) *TCPHealthChecker {
	c.retries = retries
	return c
}

// Check performs a TCP health check
func (c *TCPHealthChecker) Check(ctx context.Context, address string, port int) (*TCPCheckResult, error) {
	result := &TCPCheckResult{
		Address: address,
		Port:    port,
		Status:  StatusUnhealthy,
	}

	target := fmt.Sprintf("%s:%d", address, port)

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		start := time.Now()
		
		conn, err := net.DialTimeout("tcp", target, c.timeout)
		if err == nil {
			conn.Close()
			result.Status = StatusHealthy
			result.Latency = time.Since(start)
			result.Attempt = attempt + 1
			result.CheckedAt = time.Now().UTC()
			return result, nil
		}
		
		lastErr = err
		result.Attempt = attempt + 1
		
		if attempt < c.retries {
			backoff := time.Duration(attempt+1) * time.Second
			time.Sleep(backoff)
		}
	}

	result.Error = lastErr.Error()
	result.Status = StatusUnhealthy
	result.CheckedAt = time.Now().UTC()
	return result, lastErr
}

// TCPCheckResult result of a TCP health check
type TCPCheckResult struct {
	Address   string        `json:"address"`
	Port      int           `json:"port"`
	Status    ProbeStatus   `json:"status"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	Attempt   int           `json:"attempt"`
	CheckedAt time.Time     `json:"checked_at"`
}

// IsHealthy returns true if status is healthy
func (r *TCPCheckResult) IsHealthy() bool {
	return r.Status == StatusHealthy
}

// LatencyMs returns latency in milliseconds
func (r *TCPCheckResult) LatencyMs() float64 {
	return float64(r.Latency) / float64(time.Millisecond)
}
