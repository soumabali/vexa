package wireguard

import (
	"fmt"
	"strings"
)

func GenerateConfigFile(export *ConfigFileParams) string {
	var b strings.Builder

	b.WriteString("[Interface]\n")
	b.WriteString(fmt.Sprintf("PrivateKey = %s\n", export.ClientPrivateKey))
	b.WriteString(fmt.Sprintf("Address = %s\n", export.Address))

	if len(export.DNS) > 0 {
		b.WriteString(fmt.Sprintf("DNS = %s\n", strings.Join(export.DNS, ", ")))
	}

	if export.MTU > 0 {
		b.WriteString(fmt.Sprintf("MTU = %d\n", export.MTU))
	}

	b.WriteString("\n[Peer]\n")
	b.WriteString(fmt.Sprintf("PublicKey = %s\n", export.ServerPublicKey))

	if export.PresharedKey != "" {
		b.WriteString(fmt.Sprintf("PresharedKey = %s\n", export.PresharedKey))
	}

	b.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(export.AllowedIPs, ", ")))
	b.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", export.Endpoint, export.Port))

	if export.PersistentKeepalive > 0 {
		b.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", export.PersistentKeepalive))
	}

	return b.String()
}

type ConfigFileParams struct {
	ClientPrivateKey    string
	Address             string
	DNS                 []string
	ServerPublicKey     string
	PresharedKey        string
	AllowedIPs          []string
	Endpoint            string
	Port                int
	PersistentKeepalive int
	MTU                 int
}
