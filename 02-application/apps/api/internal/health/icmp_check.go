package health

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ICMPHealthChecker performs ICMP ping checks
type ICMPHealthChecker struct {
	timeout time.Duration
	count   int
	retries int
}

// NewICMPHealthChecker creates ICMP checker with defaults
func NewICMPHealthChecker() *ICMPHealthChecker {
	return &ICMPHealthChecker{
		timeout: 10 * time.Second,
		count:   3,
		retries: 2,
	}
}

// WithTimeout sets custom timeout
func (c *ICMPHealthChecker) WithTimeout(timeout time.Duration) *ICMPHealthChecker {
	c.timeout = timeout
	return c
}

// WithCount sets packet count
func (c *ICMPHealthChecker) WithCount(count int) *ICMPHealthChecker {
	c.count = count
	return c
}

// Check performs ICMP ping to host
func (c *ICMPHealthChecker) Check(ctx context.Context, address string) (*ICMPCheckResult, error) {
	result := &ICMPCheckResult{
		Address: address,
		Status:  StatusUnhealthy,
	}

	// Resolve IP if hostname
	ip := net.ParseIP(address)
	if ip == nil {
		ips, err := net.LookupIP(address)
		if err != nil {
			result.Error = fmt.Sprintf("DNS resolution failed: %v", err)
			return result, err
		}
		ip = ips[0]
		result.ResolvedIP = ip.String()
	} else {
		result.ResolvedIP = ip.String()
	}

	// Platform-specific ping command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "-n", strconv.Itoa(c.count), "-w", strconv.Itoa(int(c.timeout.Milliseconds())), address)
	} else {
		cmd = exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(c.count), "-W", strconv.Itoa(int(c.timeout.Seconds())), address)
	}

	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)
	result.CheckedAt = time.Now().UTC()

	if err != nil {
		result.Error = fmt.Sprintf("ping failed: %v (output: %s)", err, string(output))
		result.Status = StatusUnhealthy
		result.Latency = duration
		return result, fmt.Errorf("ping failed: %v", err)
	}

	// Parse output for latency
	outputStr := string(output)
	result.RawOutput = outputStr

	// Extract average RTT
	avgLatency := parsePingLatency(outputStr)
	if avgLatency > 0 {
		result.Latency = avgLatency
	} else {
		result.Latency = duration / time.Duration(c.count)
	}

	// Extract packet loss
	result.PacketLoss = parsePacketLoss(outputStr)

	// Determine status
	if result.PacketLoss == 100 {
		result.Status = StatusUnhealthy
		result.Error = "100% packet loss"
	} else if result.PacketLoss > 50 {
		result.Status = StatusUnhealthy
		result.Error = fmt.Sprintf("high packet loss: %.1f%%", result.PacketLoss)
	} else {
		result.Status = StatusHealthy
	}

	result.Attempt = c.count
	return result, nil
}

// ICMPCheckResult result of ICMP check
type ICMPCheckResult struct {
	Address    string        `json:"address"`
	ResolvedIP string        `json:"resolved_ip,omitempty"`
	Status     ProbeStatus   `json:"status"`
	Latency    time.Duration `json:"latency"`
	PacketLoss float64       `json:"packet_loss_percent"`
	Error      string        `json:"error,omitempty"`
	RawOutput  string        `json:"raw_output,omitempty"`
	Attempt    int           `json:"attempt"`
	CheckedAt  time.Time     `json:"checked_at"`
}

// LatencyMs returns latency in milliseconds
func (r *ICMPCheckResult) LatencyMs() float64 {
	return float64(r.Latency) / float64(time.Millisecond)
}

// IsHealthy returns true if status is healthy
func (r *ICMPCheckResult) IsHealthy() bool {
	return r.Status == StatusHealthy
}

// parsePingLatency extracts average RTT from ping output
func parsePingLatency(output string) time.Duration {
	// Linux: "rtt min/avg/max/mdev = 0.123/0.456/0.789/0.012 ms"
	// Windows: "Average = 1ms"

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Linux format
		if strings.Contains(line, "avg/") || strings.Contains(line, "rtt min/avg") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				values := strings.TrimSpace(parts[1])
				subParts := strings.Split(values, "/")
				if len(subParts) >= 2 {
					avgStr := strings.TrimSpace(subParts[1])
					avgStr = strings.TrimSuffix(avgStr, " ms")
					if ms, err := strconv.ParseFloat(avgStr, 64); err == nil {
						return time.Duration(ms * float64(time.Millisecond))
					}
				}
			}
		}

		// Windows format
		if strings.Contains(line, "Average =") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				avgStr := strings.TrimSpace(parts[1])
				avgStr = strings.TrimSuffix(avgStr, "ms")
				if ms, err := strconv.ParseFloat(avgStr, 64); err == nil {
					return time.Duration(ms * float64(time.Millisecond))
				}
			}
		}
	}

	return 0
}

// parsePacketLoss extracts packet loss percentage
func parsePacketLoss(output string) float64 {
	// Linux: "3 packets transmitted, 3 received, 0% packet loss"
	// Windows: "Lost = 0 (0% loss)"

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "packet loss") {
			// Extract percentage
			if idx := strings.Index(line, "%"); idx > 0 {
				// Find the number before %
				start := idx - 1
				for start >= 0 && (line[start] >= '0' && line[start] <= '9' || line[start] == '.') {
					start--
				}
				if start+1 < idx {
					pctStr := line[start+1 : idx]
					if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
						return pct
					}
				}
			}
		}

		if strings.Contains(line, "loss)") {
			// Windows: "Lost = 0 (0% loss)"
			if idx := strings.Index(line, "%"); idx > 0 {
				start := idx - 1
				for start >= 0 && (line[start] >= '0' && line[start] <= '9' || line[start] == '.') {
					start--
				}
				if start+1 < idx {
					pctStr := line[start+1 : idx]
					if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
						return pct
					}
				}
			}
		}
	}

	return 0
}
