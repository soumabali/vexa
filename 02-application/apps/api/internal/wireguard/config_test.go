package wireguard

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateConfigFile_Basic(t *testing.T) {
	conf := GenerateConfigFile(&ConfigFileParams{
		ClientPrivateKey:    "cFakePrivateKeyBase64",
		Address:             "10.200.200.2/24",
		DNS:                 []string{"1.1.1.1", "8.8.8.8"},
		ServerPublicKey:     "sFakePublicKeyBase64",
		AllowedIPs:          []string{"0.0.0.0/0", "::/0"},
		Endpoint:            "vpn.example.com",
		Port:                51820,
		PersistentKeepalive: 25,
		MTU:                 1420,
	})

	assert.Contains(t, conf, "[Interface]")
	assert.Contains(t, conf, "PrivateKey = cFakePrivateKeyBase64")
	assert.Contains(t, conf, "Address = 10.200.200.2/24")
	assert.Contains(t, conf, "DNS = 1.1.1.1, 8.8.8.8")
	assert.Contains(t, conf, "MTU = 1420")
	assert.Contains(t, conf, "[Peer]")
	assert.Contains(t, conf, "PublicKey = sFakePublicKeyBase64")
	assert.Contains(t, conf, "AllowedIPs = 0.0.0.0/0, ::/0")
	assert.Contains(t, conf, "Endpoint = vpn.example.com:51820")
	assert.Contains(t, conf, "PersistentKeepalive = 25")
}

func TestGenerateConfigFile_WithPSK(t *testing.T) {
	conf := GenerateConfigFile(&ConfigFileParams{
		ClientPrivateKey: "cPriv",
		Address:          "10.200.200.2/24",
		ServerPublicKey:  "sPub",
		PresharedKey:     "pskValue",
		AllowedIPs:       []string{"0.0.0.0/0"},
		Endpoint:         "vpn.example.com",
		Port:             51820,
	})

	assert.Contains(t, conf, "PresharedKey = pskValue")
}

func TestGenerateConfigFile_NoOptional(t *testing.T) {
	conf := GenerateConfigFile(&ConfigFileParams{
		ClientPrivateKey: "cPriv",
		Address:          "10.200.200.2/24",
		ServerPublicKey:  "sPub",
		AllowedIPs:       []string{"0.0.0.0/0"},
		Endpoint:         "vpn.example.com",
		Port:             51820,
	})

	assert.NotContains(t, conf, "PresharedKey")
	assert.NotContains(t, conf, "DNS")
	assert.NotContains(t, conf, "MTU")
	assert.NotContains(t, conf, "PersistentKeepalive")
	assert.True(t, strings.HasPrefix(conf, "[Interface]"))
	assert.Contains(t, conf, "\n\n[Peer]")
}
