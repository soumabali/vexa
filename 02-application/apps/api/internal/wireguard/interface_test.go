package wireguard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInterfaceController_NoWG(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	assert.True(t, ic.IsMock())
}

func TestInterfaceController_MockCreate(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	err := ic.CreateInterface("wg0", "10.200.200.1/24", 51820, 1420)
	assert.NoError(t, err)
}

func TestInterfaceController_MockDelete(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	err := ic.DeleteInterface("wg0")
	assert.NoError(t, err)
}

func TestInterfaceController_MockSetPeer(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	err := ic.SetPeer("wg0", PeerConfig{
		PublicKey:           "testPubKey",
		AllowedIPs:          []string{"10.200.200.2/32"},
		PersistentKeepalive: 25,
	})
	assert.NoError(t, err)
}

func TestInterfaceController_MockSetPeerWithPSK(t *testing.T) {
	psk := "testPSK"
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	err := ic.SetPeer("wg0", PeerConfig{
		PublicKey:           "testPubKey",
		PresharedKey:        &psk,
		AllowedIPs:          []string{"0.0.0.0/0"},
		PersistentKeepalive: 25,
	})
	assert.NoError(t, err)
}

func TestInterfaceController_MockRemovePeer(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	err := ic.RemovePeer("wg0", "testPubKey")
	assert.NoError(t, err)
}

func TestInterfaceController_MockGetStats(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	stats := ic.GetTransferStats("wg0")
	assert.Equal(t, int64(0), stats.BytesSent)
	assert.Equal(t, int64(0), stats.BytesReceived)
}

func TestInterfaceController_MockUpDown(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")

	err := ic.SetInterfaceUp("wg0")
	assert.NoError(t, err)

	err = ic.SetInterfaceDown("wg0")
	assert.NoError(t, err)
}

func TestInterfaceController_MockExists(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	assert.True(t, ic.InterfaceExists("wg0"))
}

func TestInterfaceController_MockUpdatePeerKey(t *testing.T) {
	ic := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	err := ic.UpdatePeerKey("wg0", "testPubKey", "newPrivKey")
	assert.NoError(t, err)
}

func TestInterfaceController_DefaultBinPath(t *testing.T) {
	ic := NewInterfaceController("", "10.200.200.0/24")
	assert.Equal(t, "wg", ic.binPath)
}
