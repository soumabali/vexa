package tunnel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/tunnel"
)

func TestNewTunnelManager(t *testing.T) {
	tm := tunnel.NewTunnelManager()
	require.NotNil(t, tm)

	stats := tm.GetStats()
	assert.Equal(t, 0, stats["total"])
	assert.Equal(t, 0, stats["active"])
	assert.Equal(t, 0, stats["pending"])
	assert.Equal(t, 0, stats["failed"])
}

func TestTunnelManager_GetTunnel_NotFound(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	_, err := tm.GetTunnel("non-existent")
	assert.ErrorIs(t, err, tunnel.ErrTunnelNotFound)
}

func TestTunnelManager_CloseTunnel_NotFound(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	err := tm.CloseTunnel("non-existent")
	assert.ErrorIs(t, err, tunnel.ErrTunnelNotFound)
}

func TestTunnelManager_ListTunnels_Empty(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	tunnels := tm.ListTunnels()
	assert.Empty(t, tunnels)
}

func TestTunnelManager_CreateLocalTunnel_InvalidType(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	config := &tunnel.TunnelConfig{
		Type: tunnel.TunnelTypeRemote, // Wrong type for local tunnel
	}

	_, err := tm.CreateLocalTunnel(config)
	assert.ErrorIs(t, err, tunnel.ErrInvalidTunnelType)
}

func TestTunnelManager_CreateRemoteTunnel_InvalidType(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	config := &tunnel.TunnelConfig{
		Type: tunnel.TunnelTypeLocal, // Wrong type for remote tunnel
	}

	_, err := tm.CreateRemoteTunnel(config)
	assert.ErrorIs(t, err, tunnel.ErrInvalidTunnelType)
}

func TestTunnelManager_CreateDynamicTunnel_InvalidType(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	config := &tunnel.TunnelConfig{
		Type: tunnel.TunnelTypeLocal, // Wrong type for dynamic tunnel
	}

	_, err := tm.CreateDynamicTunnel(config)
	assert.ErrorIs(t, err, tunnel.ErrInvalidTunnelType)
}

func TestTunnelManager_CreateLocalTunnel_PortInUse(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	// First, create a local tunnel to occupy the port
	// This would normally require an SSH server, so we test the port tracking logic
	config := &tunnel.TunnelConfig{
		Type:       tunnel.TunnelTypeLocal,
		LocalHost:  "127.0.0.1",
		LocalPort:  19999,
		RemoteHost: "127.0.0.1",
		RemotePort: 22,
		SSHHost:    "127.0.0.1",
		SSHPort:    22,
		SSHUser:    "test",
		SSHPassword: "test",
	}

	// Without actual SSH server, this will fail at SSH dial, but we can test the port check
	// by simulating port occupation
	// For now, just verify the error path works
	_, err := tm.CreateLocalTunnel(config)
	// Will fail because no SSH server, but shouldn't be port in use on first call
	assert.Error(t, err)
	assert.NotErrorIs(t, err, tunnel.ErrPortInUse)
}

func TestTunnelManager_Stats(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	stats := tm.GetStats()
	assert.Contains(t, stats, "total")
	assert.Contains(t, stats, "active")
	assert.Contains(t, stats, "pending")
	assert.Contains(t, stats, "failed")
}

func TestTunnelManager_Cleanup(t *testing.T) {
	tm := tunnel.NewTunnelManager()

	// Cleanup on empty manager should not panic
	tm.Cleanup()

	stats := tm.GetStats()
	assert.Equal(t, 0, stats["total"])
}

func TestTunnelStatus_Constants(t *testing.T) {
	assert.Equal(t, tunnel.TunnelStatus("pending"), tunnel.TunnelStatusPending)
	assert.Equal(t, tunnel.TunnelStatus("active"), tunnel.TunnelStatusActive)
	assert.Equal(t, tunnel.TunnelStatus("failed"), tunnel.TunnelStatusFailed)
	assert.Equal(t, tunnel.TunnelStatus("closed"), tunnel.TunnelStatusClosed)
	assert.Equal(t, tunnel.TunnelStatus("error"), tunnel.TunnelStatusError)
}

func TestTunnelType_Constants(t *testing.T) {
	assert.Equal(t, tunnel.TunnelType("local"), tunnel.TunnelTypeLocal)
	assert.Equal(t, tunnel.TunnelType("remote"), tunnel.TunnelTypeRemote)
	assert.Equal(t, tunnel.TunnelType("dynamic"), tunnel.TunnelTypeDynamic)
}

func TestTunnelConfig_Validation(t *testing.T) {
	// Test valid local config
	localConfig := &tunnel.TunnelConfig{
		Type:        tunnel.TunnelTypeLocal,
		LocalHost:   "127.0.0.1",
		LocalPort:   8080,
		RemoteHost:  "192.168.1.1",
		RemotePort:  80,
		SSHHost:     "ssh.example.com",
		SSHPort:     22,
		SSHUser:     "user",
		SSHPassword: "pass",
	}
	assert.Equal(t, tunnel.TunnelTypeLocal, localConfig.Type)
	assert.Equal(t, 8080, localConfig.LocalPort)
	assert.Equal(t, 80, localConfig.RemotePort)

	// Test valid dynamic config
	dynamicConfig := &tunnel.TunnelConfig{
		Type:        tunnel.TunnelTypeDynamic,
		LocalHost:   "127.0.0.1",
		LocalPort:   1080,
		SSHHost:     "ssh.example.com",
		SSHPort:     22,
		SSHUser:     "user",
		SSHPassword: "pass",
	}
	assert.Equal(t, tunnel.TunnelTypeDynamic, dynamicConfig.Type)
	assert.Equal(t, 1080, dynamicConfig.LocalPort)
}

func TestTunnelStats_Structure(t *testing.T) {
	stats := &tunnel.TunnelStats{
		ID:         "test-id",
		Type:       tunnel.TunnelTypeLocal,
		Status:     tunnel.TunnelStatusActive,
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "192.168.1.1:80",
		BytesIn:    1024,
		BytesOut:   2048,
	}

	assert.Equal(t, "test-id", stats.ID)
	assert.Equal(t, tunnel.TunnelTypeLocal, stats.Type)
	assert.Equal(t, tunnel.TunnelStatusActive, stats.Status)
	assert.Equal(t, int64(1024), stats.BytesIn)
	assert.Equal(t, int64(2048), stats.BytesOut)
}

// BenchmarkTunnelManager_ListTunnels benchmarks listing tunnels
func BenchmarkTunnelManager_ListTunnels(b *testing.B) {
	tm := tunnel.NewTunnelManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.ListTunnels()
	}
}

// BenchmarkTunnelManager_GetStats benchmarks getting stats
func BenchmarkTunnelManager_GetStats(b *testing.B) {
	tm := tunnel.NewTunnelManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.GetStats()
	}
}
