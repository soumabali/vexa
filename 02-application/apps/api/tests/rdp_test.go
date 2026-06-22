package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/rdp"
)

func TestNewGateway(t *testing.T) {
	g := rdp.NewGateway()
	assert.NotNil(t, g)
	assert.Empty(t, g.ListSessions())
}

func TestCreateSession(t *testing.T) {
	g := rdp.NewGateway()
	
	session, err := g.CreateSession(
		"host-1",
		"user-1",
		"192.168.1.100",
		3389,
		"admin",
		"password123",
		"WORKGROUP",
	)
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "host-1", session.HostID)
	assert.Equal(t, "user-1", session.UserID)
	assert.Equal(t, "192.168.1.100", session.Hostname)
	assert.Equal(t, 3389, session.Port)
	assert.Equal(t, "admin", session.Username)
	assert.Equal(t, "WORKGROUP", session.Domain)
	assert.Equal(t, rdp.StatusDisconnected, session.Status())
}

func TestGetSession(t *testing.T) {
	g := rdp.NewGateway()
	session, _ := g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")
	
	// Get existing
	found, ok := g.GetSession(session.ID)
	assert.True(t, ok)
	assert.Equal(t, session.ID, found.ID)
	
	// Get non-existing
	_, ok = g.GetSession("non-existent")
	assert.False(t, ok)
}

func TestListSessions(t *testing.T) {
	g := rdp.NewGateway()
	
	// Empty list
	assert.Empty(t, g.ListSessions())
	
	// Add sessions
	g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")
	g.CreateSession("host-2", "user-1", "192.168.1.101", 3389, "admin", "pass", "")
	
	sessions := g.ListSessions()
	assert.Len(t, sessions, 2)
}

func TestCloseSession(t *testing.T) {
	g := rdp.NewGateway()
	session, _ := g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")
	
	err := g.CloseSession(session.ID)
	assert.NoError(t, err)
	
	// Verify closed
	_, ok := g.GetSession(session.ID)
	assert.False(t, ok)
	
	// Close non-existent
	err = g.CloseSession("non-existent")
	assert.Error(t, err)
}

func TestSessionStatus(t *testing.T) {
	g := rdp.NewGateway()
	session, _ := g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")
	
	assert.Equal(t, rdp.StatusDisconnected, session.Status())
	assert.Equal(t, "disconnected", session.Status().String())
}

func TestSessionStats(t *testing.T) {
	g := rdp.NewGateway()
	session, _ := g.CreateSession("host-1", "user-1", "192.168.1.100", 3389, "admin", "pass", "")
	
	stats := session.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, session.ID, stats["id"])
	assert.Equal(t, "disconnected", stats["status"])
	assert.Equal(t, int64(0), stats["bytes_sent"])
	assert.Equal(t, int64(0), stats["bytes_recv"])
}

func TestFrameTypes(t *testing.T) {
	assert.Equal(t, byte(0x01), rdp.FrameTypeScreen)
	assert.Equal(t, byte(0x02), rdp.FrameTypeInput)
	assert.Equal(t, byte(0x03), rdp.FrameTypeAudio)
	assert.Equal(t, byte(0x04), rdp.FrameTypeResize)
	assert.Equal(t, byte(0x05), rdp.FrameTypeClipboard)
	assert.Equal(t, byte(0x06), rdp.FrameTypeCursor)
	assert.Equal(t, byte(0xFF), rdp.FrameTypeError)
}

func TestClientConfig(t *testing.T) {
	g := rdp.NewGateway()
	config := rdp.ClientConfig{
		Hostname:          "192.168.1.100",
		Port:              3389,
		Username:          "admin",
		Password:          "pass",
		Domain:            "WORKGROUP",
		Width:             1920,
		Height:            1080,
		ColorDepth:        32,
		ReconnectAttempts: 5,
		ReconnectDelay:    10 * time.Second,
	}
	
	client, err := rdp.NewClient(g, config)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 5, config.ReconnectAttempts)
	assert.Equal(t, 10*time.Second, config.ReconnectDelay)
}

func TestClientSessionID(t *testing.T) {
	g := rdp.NewGateway()
	config := rdp.ClientConfig{
		Hostname: "192.168.1.100",
		Port:     3389,
		Username: "admin",
		Password: "pass",
	}
	
	client, _ := rdp.NewClient(g, config)
	assert.NotEmpty(t, client.SessionID())
}

func TestClientStats(t *testing.T) {
	g := rdp.NewGateway()
	config := rdp.ClientConfig{
		Hostname: "192.168.1.100",
		Port:     3389,
		Username: "admin",
		Password: "pass",
	}
	
	client, _ := rdp.NewClient(g, config)
	stats := client.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, "disconnected", stats["status"])
}

func TestClientIsConnected(t *testing.T) {
	g := rdp.NewGateway()
	config := rdp.ClientConfig{
		Hostname: "192.168.1.100",
		Port:     3389,
		Username: "admin",
		Password: "pass",
	}
	
	client, _ := rdp.NewClient(g, config)
	assert.False(t, client.IsConnected())
}

func TestClientSendInput(t *testing.T) {
	g := rdp.NewGateway()
	config := rdp.ClientConfig{
		Hostname: "192.168.1.100",
		Port:     3389,
		Username: "admin",
		Password: "pass",
	}
	
	client, _ := rdp.NewClient(g, config)
	
	// Should succeed when disconnected
	err := client.SendInput([]byte("test"))
	assert.NoError(t, err)
	
	// Disconnect and try again
	client.Disconnect()
	err = client.SendInput([]byte("test"))
	assert.Error(t, err)
}
