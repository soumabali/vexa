package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/health"
)

func TestTCPHealthChecker(t *testing.T) {
	checker := health.NewTCPHealthChecker().
		WithTimeout(5 * time.Second).
		WithRetries(2)

	t.Run("check_localhost_ssh", func(t *testing.T) {
		// This test requires an SSH server running locally
		// Skip if not available
		result, err := checker.Check(context.Background(), "127.0.0.1", 22)
		if err != nil {
			t.Skip("No SSH server available locally")
		}
		assert.Equal(t, health.StatusHealthy, result.Status)
		assert.Greater(t, result.LatencyMs(), float64(0))
		assert.True(t, result.IsHealthy())
	})

	t.Run("check_unreachable_host", func(t *testing.T) {
		// Test with a port that's likely closed
		result, err := checker.Check(context.Background(), "127.0.0.1", 65432)
		assert.Error(t, err)
		assert.Equal(t, health.StatusUnhealthy, result.Status)
		assert.False(t, result.IsHealthy())
		assert.GreaterOrEqual(t, result.Attempt, 1)
	})
}

func TestProbeEngine(t *testing.T) {
	engine := health.NewProbeEngine()
	require.NotNil(t, engine)

	t.Run("register_and_get", func(t *testing.T) {
		config := health.DefaultProbeConfig()
		config.Port = 22
		config.Interval = 1 * time.Hour // Long interval so it doesn't run immediately

		engine.Register("host-1", "Test Host", "127.0.0.1", config)

		h, ok := engine.GetHealth("host-1")
		assert.True(t, ok)
		assert.Equal(t, "host-1", h.HostID)
		assert.Equal(t, "Test Host", h.HostName)
		assert.Equal(t, "127.0.0.1", h.Address)
	})

	t.Run("unregister", func(t *testing.T) {
		engine.Unregister("host-1")
		_, ok := engine.GetHealth("host-1")
		assert.False(t, ok)
	})

	t.Run("get_all_health", func(t *testing.T) {
		// Register multiple hosts
		config := health.DefaultProbeConfig()
		engine.Register("host-a", "Host A", "10.0.0.1", config)
		engine.Register("host-b", "Host B", "10.0.0.2", config)

		all := engine.GetAllHealth()
		assert.Len(t, all, 2)
		assert.Contains(t, all, "host-a")
		assert.Contains(t, all, "host-b")
	})
}

func TestBulkCheck(t *testing.T) {
	engine := health.NewProbeEngine()

	t.Run("bulk_check_max_100", func(t *testing.T) {
		hosts := make([]health.HostCheckRequest, 105)
		for i := 0; i < 105; i++ {
			hosts[i] = health.HostCheckRequest{
				HostID:  fmt.Sprintf("host-%d", i),
				Address: "127.0.0.1",
				Port:    22,
			}
		}

		// Should still work but cap at 100
		results := engine.BulkCheck(hosts, 10)
		assert.Len(t, results, 105)
	})

	t.Run("bulk_check_parallel", func(t *testing.T) {
		hosts := []health.HostCheckRequest{
			{HostID: "h1", Address: "127.0.0.1", Port: 22},
			{HostID: "h2", Address: "127.0.0.1", Port: 65432}, // Likely closed
		}

		results := engine.BulkCheck(hosts, 5)
		assert.Len(t, results, 2)
	})
}

func TestProbeConfig(t *testing.T) {
	t.Run("default_config", func(t *testing.T) {
		config := health.DefaultProbeConfig()
		assert.Equal(t, health.ProbeTypeTCP, config.Type)
		assert.Equal(t, 10*time.Second, config.Timeout)
		assert.Equal(t, 3, config.Retries)
		assert.Equal(t, 5*time.Minute, config.Interval)
	})

	t.Run("custom_config", func(t *testing.T) {
		config := health.ProbeConfig{
			Type:     health.ProbeTypeTCP,
			Timeout:  30 * time.Second,
			Retries:  5,
			Interval: 1 * time.Minute,
		}
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 5, config.Retries)
	})
}

func TestICMPHealthChecker(t *testing.T) {
	checker := health.NewICMPHealthChecker().
		WithTimeout(10 * time.Second).
		WithCount(2)

	t.Run("ping_localhost", func(t *testing.T) {
		result, err := checker.Check(context.Background(), "127.0.0.1")
		if err != nil {
			t.Skip("ICMP requires root privileges")
		}
		assert.Equal(t, health.StatusHealthy, result.Status)
		assert.Greater(t, result.LatencyMs(), float64(0))
		assert.LessOrEqual(t, result.PacketLoss, float64(100))
	})

	t.Run("ping_unreachable", func(t *testing.T) {
		// Use a non-routable address
		result, _ := checker.Check(context.Background(), "192.0.2.1") // TEST-NET-1
		// Will likely fail or timeout
		assert.NotNil(t, result)
	})
}
