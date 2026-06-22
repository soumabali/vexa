package wireguard

import (
	"encoding/base64"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair(t *testing.T) {
	priv, pub, err := generateKeyPair()
	require.NoError(t, err)
	require.NotEmpty(t, priv)
	require.NotEmpty(t, pub)

	privBytes, err := base64.StdEncoding.DecodeString(priv)
	require.NoError(t, err)
	assert.Len(t, privBytes, 32)

	pubBytes, err := base64.StdEncoding.DecodeString(pub)
	require.NoError(t, err)
	assert.Len(t, pubBytes, 32)
}

func TestGenerateKeyPair_Unique(t *testing.T) {
	_, pub1, _ := generateKeyPair()
	_, pub2, _ := generateKeyPair()
	assert.NotEqual(t, pub1, pub2)
}

func TestGenerateClientPublicKey(t *testing.T) {
	pub, err := generateClientPublicKey()
	require.NoError(t, err)
	require.NotEmpty(t, pub)

	pubBytes, err := base64.StdEncoding.DecodeString(pub)
	require.NoError(t, err)
	assert.Len(t, pubBytes, 32)
}

func TestGeneratePSK(t *testing.T) {
	psk := generatePSK()
	require.NotEmpty(t, psk)

	pskBytes, err := base64.StdEncoding.DecodeString(psk)
	require.NoError(t, err)
	assert.Len(t, pskBytes, 32)
}

func TestGeneratePSK_Unique(t *testing.T) {
	psk1 := generatePSK()
	psk2 := generatePSK()
	assert.NotEqual(t, psk1, psk2)
}

func TestIncrementIP_IPv4(t *testing.T) {
	ip := incrementIP(netParse("10.0.0.0"), 1)
	assert.Equal(t, "10.0.0.1", ip.String())

	ip = incrementIP(netParse("10.0.0.0"), 2)
	assert.Equal(t, "10.0.0.2", ip.String())
}

func TestIncrementIP_CarryOver(t *testing.T) {
	ip := incrementIP(netParse("10.0.0.255"), 1)
	assert.Equal(t, "10.0.1.0", ip.String())

	ip = incrementIP(netParse("10.0.255.255"), 1)
	assert.Equal(t, "10.1.0.0", ip.String())
}

func TestIncrementIP_ZeroIncrement(t *testing.T) {
	ip := incrementIP(netParse("10.0.0.1"), 0)
	assert.Equal(t, "10.0.0.1", ip.String())
}

func TestNewWireGuardService_Disabled(t *testing.T) {
	svc := NewWireGuardService(nil, nil, ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		RotationDays:      90,
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           false,
		BinPath:           "/nonexistent/wg",
	})
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.Repository())
}

func TestJoinStrings(t *testing.T) {
	assert.Equal(t, "", joinStrings(nil, ","))
	assert.Equal(t, "", joinStrings([]string{}, ","))
	assert.Equal(t, "a", joinStrings([]string{"a"}, ","))
	assert.Equal(t, "a,b", joinStrings([]string{"a", "b"}, ","))
	assert.Equal(t, "a, b", joinStrings([]string{"a", " b"}, ","))
}

func netParse(s string) net.IP {
	return net.ParseIP(s).To4()
}
