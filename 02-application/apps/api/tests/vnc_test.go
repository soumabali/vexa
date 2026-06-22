package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/vnc"
)

func TestNewProxy(t *testing.T) {
	p := vnc.NewProxy()
	assert.NotNil(t, p)
	assert.Empty(t, p.ListSessions())
}

func TestVNCProxyCreateSession(t *testing.T) {
	p := vnc.NewProxy()
	
	session, err := p.CreateSession(
		"host-1",
		"user-1",
		"192.168.1.100",
		5900,
		"password123",
	)
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "host-1", session.HostID)
	assert.Equal(t, "user-1", session.UserID)
	assert.Equal(t, "192.168.1.100", session.Hostname)
	assert.Equal(t, 5900, session.Port)
	assert.Equal(t, "password123", session.Password)
	assert.Equal(t, vnc.StatusDisconnected, session.Status())
}

func TestVNCProxyGetSession(t *testing.T) {
	p := vnc.NewProxy()
	session, _ := p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")
	
	// Get existing
	found, ok := p.GetSession(session.ID)
	assert.True(t, ok)
	assert.Equal(t, session.ID, found.ID)
	
	// Get non-existing
	_, ok = p.GetSession("non-existent")
	assert.False(t, ok)
}

func TestVNCListSessions(t *testing.T) {
	p := vnc.NewProxy()
	
	// Empty list
	assert.Empty(t, p.ListSessions())
	
	// Add sessions
	p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")
	p.CreateSession("host-2", "user-1", "192.168.1.101", 5901, "pass")
	
	sessions := p.ListSessions()
	assert.Len(t, sessions, 2)
}

func TestVNCProxyCloseSession(t *testing.T) {
	p := vnc.NewProxy()
	session, _ := p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")
	
	err := p.CloseSession(session.ID)
	assert.NoError(t, err)
	
	// Verify closed
	_, ok := p.GetSession(session.ID)
	assert.False(t, ok)
	
	// Close non-existent
	err = p.CloseSession("non-existent")
	assert.Error(t, err)
}

func TestVNCStatusString(t *testing.T) {
	assert.Equal(t, "disconnected", vnc.StatusDisconnected.String())
	assert.Equal(t, "connecting", vnc.StatusConnecting.String())
	assert.Equal(t, "handshaking", vnc.StatusHandshaking.String())
	assert.Equal(t, "authenticating", vnc.StatusAuthenticating.String())
	assert.Equal(t, "connected", vnc.StatusConnected.String())
	assert.Equal(t, "error", vnc.StatusError.String())
	assert.Equal(t, "closed", vnc.StatusClosed.String())
}

func TestVNCFrameTypes(t *testing.T) {
	assert.Equal(t, byte(0x01), vnc.FrameTypeScreen)
	assert.Equal(t, byte(0x02), vnc.FrameTypeInput)
	assert.Equal(t, byte(0x03), vnc.FrameTypeAudio)
	assert.Equal(t, byte(0x04), vnc.FrameTypeResize)
	assert.Equal(t, byte(0x05), vnc.FrameTypeClipboard)
	assert.Equal(t, byte(0x06), vnc.FrameTypeCursor)
	assert.Equal(t, byte(0xFF), vnc.FrameTypeError)
}

func TestVNCClient(t *testing.T) {
	p := vnc.NewProxy()
	config := vnc.ClientConfig{
		Hostname:          "192.168.1.100",
		Port:              5900,
		Password:          "pass",
		ReconnectAttempts: 3,
		ReconnectDelay:    5 * time.Second,
	}
	
	client, err := vnc.NewClient(p, config)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 3, config.ReconnectAttempts)
	assert.Equal(t, 5*time.Second, config.ReconnectDelay)
	assert.NotEmpty(t, client.SessionID())
}

func TestVNCClientStats(t *testing.T) {
	p := vnc.NewProxy()
	config := vnc.ClientConfig{
		Hostname: "192.168.1.100",
		Port:     5900,
		Password: "pass",
	}
	
	client, _ := vnc.NewClient(p, config)
	stats := client.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, "disconnected", stats["status"])
}

func TestVNCSessionStats(t *testing.T) {
	p := vnc.NewProxy()
	session, _ := p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")
	
	stats := session.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, session.ID, stats["id"])
	assert.Equal(t, "disconnected", stats["status"])
	assert.Equal(t, int64(0), stats["bytes_sent"])
	assert.Equal(t, int64(0), stats["bytes_recv"])
}

func TestVNCClientDisconnect(t *testing.T) {
	p := vnc.NewProxy()
	config := vnc.ClientConfig{
		Hostname: "192.168.1.100",
		Port:     5900,
		Password: "pass",
	}
	
	client, _ := vnc.NewClient(p, config)
	assert.False(t, client.IsConnected())
	
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestVNCClientSendInput(t *testing.T) {
	p := vnc.NewProxy()
	config := vnc.ClientConfig{
		Hostname: "192.168.1.100",
		Port:     5900,
		Password: "pass",
	}
	
	client, _ := vnc.NewClient(p, config)
	
	// Should succeed when disconnected
	err := client.SendInput([]byte("test"))
	assert.NoError(t, err)
	
	// Disconnect and try again
	client.Disconnect()
	err = client.SendInput([]byte("test"))
	assert.Error(t, err)
}

func TestVNCSessionClose(t *testing.T) {
	p := vnc.NewProxy()
	session, _ := p.CreateSession("host-1", "user-1", "192.168.1.100", 5900, "pass")
	
	err := session.Close()
	assert.NoError(t, err)
	
	// Double close should not error
	err = session.Close()
	assert.NoError(t, err)
}

func TestVNCConstants(t *testing.T) {
	assert.Equal(t, "RFB 003.008\n", vnc.VNCVersion)
	assert.Equal(t, 1, vnc.SecurityNone)
	assert.Equal(t, 2, vnc.SecurityVNCAuth)
	assert.Equal(t, 0, vnc.MsgFramebufferUpdate)
	assert.Equal(t, 0, vnc.EncRaw)
	assert.Equal(t, 5, vnc.EncHextile)
	assert.Equal(t, 7, vnc.EncTight)
	assert.Equal(t, 16, vnc.EncZRLE)
}