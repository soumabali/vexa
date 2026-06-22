package wireguard

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type PeerConfig struct {
	PublicKey           string
	PresharedKey        *string
	Endpoint            string
	AllowedIPs          []string
	PersistentKeepalive int
}

type InterfaceStats struct {
	BytesSent     int64
	BytesReceived int64
}

type InterfaceController struct {
	binPath string
	subnet  string
	mu      sync.Mutex
	mock    bool
}

func NewInterfaceController(binPath, subnet string) *InterfaceController {
	if binPath == "" {
		binPath = "wg"
	}

	_, err := exec.LookPath(binPath)
	mock := err != nil
	if mock {
		log.Printf("warning: wg binary not found at %q, running in mock mode", binPath)
	}

	return &InterfaceController{
		binPath: binPath,
		subnet:  subnet,
		mock:    mock,
	}
}

func (ic *InterfaceController) CreateInterface(name, ipAddr string, port, mtu int) error {
	if ic.mock {
		log.Printf("[mock] create interface %s ip=%s port=%d mtu=%d", name, ipAddr, port, mtu)
		return nil
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	cmds := [][]string{
		{"ip", "link", "add", "dev", name, "type", "wireguard"},
		{"ip", "address", "add", "dev", name, ipAddr},
		{"ip", "link", "set", "dev", name, "mtu", fmt.Sprintf("%d", mtu)},
		{"ip", "link", "set", "dev", name, "up"},
	}

	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("failed to execute %s: %w", strings.Join(cmd, " "), err)
		}
	}

	return nil
}

func (ic *InterfaceController) DeleteInterface(name string) error {
	if ic.mock {
		log.Printf("[mock] delete interface %s", name)
		return nil
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	return exec.Command("ip", "link", "delete", "dev", name).Run()
}

func (ic *InterfaceController) SetInterfaceUp(name string) error {
	if ic.mock {
		log.Printf("[mock] set interface %s up", name)
		return nil
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	if err := exec.Command("ip", "link", "set", "dev", name, "up").Run(); err != nil {
		return fmt.Errorf("failed to set interface up: %w", err)
	}
	return nil
}

func (ic *InterfaceController) SetInterfaceDown(name string) error {
	if ic.mock {
		log.Printf("[mock] set interface %s down", name)
		return nil
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	if err := exec.Command("ip", "link", "set", "dev", name, "down").Run(); err != nil {
		return fmt.Errorf("failed to set interface down: %w", err)
	}
	return nil
}

func (ic *InterfaceController) SetPeer(iface string, peer PeerConfig) error {
	if ic.mock {
		log.Printf("[mock] set peer on %s: pubkey=%s", iface, peer.PublicKey)
		return nil
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	args := []string{"set", iface, "peer", peer.PublicKey,
		"allowed-ips", strings.Join(peer.AllowedIPs, ","),
		"persistent-keepalive", fmt.Sprintf("%d", peer.PersistentKeepalive),
	}

	if peer.PresharedKey != nil {
		args = append(args, "preshared-key", *peer.PresharedKey)
	}

	if peer.Endpoint != "" {
		args = append(args, "endpoint", peer.Endpoint)
	}

	return exec.Command(ic.binPath, args...).Run()
}

func (ic *InterfaceController) RemovePeer(iface, publicKey string) error {
	if ic.mock {
		log.Printf("[mock] remove peer %s from %s", publicKey, iface)
		return nil
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	return exec.Command(ic.binPath, "set", iface, "peer", publicKey, "remove").Run()
}

func (ic *InterfaceController) UpdatePeerKey(iface, peerPubKey, newPrivateKey string) error {
	if ic.mock {
		log.Printf("[mock] update peer key on %s", iface)
		return nil
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	cmds := [][]string{
		{ic.binPath, "set", iface, "private-key", "/dev/stdin"},
	}

	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("failed to update key: %w", err)
		}
	}

	return nil
}

func (ic *InterfaceController) GetTransferStats(iface string) InterfaceStats {
	if ic.mock {
		return InterfaceStats{}
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	out, err := exec.Command(ic.binPath, "show", iface, "transfer").Output()
	if err != nil {
		return InterfaceStats{}
	}

	var sent, received int64
	parsed, _ := fmt.Sscanf(string(out), "%d %d", &sent, &received)
	if parsed != 2 {
		return InterfaceStats{}
	}

	return InterfaceStats{BytesSent: sent, BytesReceived: received}
}

func (ic *InterfaceController) InterfaceExists(name string) bool {
	if ic.mock {
		return true
	}

	return exec.Command("ip", "link", "show", "dev", name).Run() == nil
}

func (ic *InterfaceController) IsMock() bool {
	return ic.mock
}
