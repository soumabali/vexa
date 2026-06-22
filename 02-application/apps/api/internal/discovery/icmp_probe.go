package discovery

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// ICMPProber handles ICMP ping sweeps
type ICMPProber struct{}

// NewICMPProber creates a new ICMP prober
func NewICMPProber() *ICMPProber {
	return &ICMPProber{}
}

// Probe sends an ICMP echo request and measures response time
func (p *ICMPProber) Probe(ip string, timeoutMs int) (bool, int64) {
	// On non-root systems, ICMP may not work; return true to always scan TCP
	// In production, use raw sockets (requires root/cap_net_raw) or use
	// alternative methods (e.g., TCP SYN scan)
	if runtime.GOOS != "linux" {
		// Non-Linux: fallback to TCP-only
		return true, 0
	}

	timeout := time.Duration(timeoutMs) * time.Millisecond
	duration, err := ping(ip, timeout)
	if err != nil {
		return false, 0
	}
	return true, duration.Milliseconds()
}

// ping sends an ICMP echo request
func ping(ip string, timeout time.Duration) (time.Duration, error) {
	targetIP := net.ParseIP(ip)
	if targetIP == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Determine if IPv4 or IPv6
	isIPv4 := targetIP.To4() != nil
	var proto string
	if isIPv4 {
		proto = "ip4:icmp"
	} else {
		proto = "ip6:ipv6-icmp"
	}

	// Create ICMP connection
	c, err := icmp.ListenPacket(proto, "")
	if err != nil {
		return 0, fmt.Errorf("failed to listen for ICMP: %w (try running as root)", err)
	}
	defer c.Close()

	// Build ICMP Echo Request
	var msg *icmp.Message
	if isIPv4 {
		msg = &icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   1,
				Seq:  1,
				Data: []byte("SSH-MANAGER-PING"),
			},
		}
	} else {
		msg = &icmp.Message{
			Type: ipv6.ICMPTypeEchoRequest,
			Code: 0,
			Body: &icmp.Echo{
				ID:   1,
				Seq:  1,
				Data: []byte("SSH-MANAGER-PING"),
			},
		}
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal ICMP message: %w", err)
	}

	// Send ICMP
	start := time.Now()
	n, err := c.WriteTo(data, &net.IPAddr{IP: targetIP})
	if err != nil {
		return 0, fmt.Errorf("failed to send ICMP: %w", err)
	}
	if n != len(data) {
		return 0, fmt.Errorf("partial ICMP write: %d/%d bytes", n, len(data))
	}

	// Set read timeout
	c.SetReadDeadline(time.Now().Add(timeout))

	// Read reply
	reply := make([]byte, 1500)
	_, peer, err := c.ReadFrom(reply)
	if err != nil {
		return 0, fmt.Errorf("ICMP timeout or error: %w", err)
	}

	rtt := time.Since(start)

	// Parse reply
	rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ICMP reply: %w", err)
	}

	// Validate reply
	if isIPv4 {
		if rm.Type != ipv4.ICMPTypeEchoReply {
			return 0, fmt.Errorf("unexpected ICMP type: %v", rm.Type)
		}
	} else {
		if rm.Type != ipv6.ICMPTypeEchoReply {
			return 0, fmt.Errorf("unexpected ICMP type: %v", rm.Type)
		}
	}

	_ = peer
	return rtt, nil
}

// PingSweep performs a ping sweep on a CIDR range
// Returns a map of IP -> response time in ms
func (p *ICMPProber) PingSweep(cidr string, timeoutMs int, concurrency int) (map[string]int64, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	var ips []string
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ips = append(ips, dupIP(ip).String())
	}

	results := make(map[string]int64)
	var mu sync.Mutex

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, ip := range ips {
		sem <- struct{}{}
		wg.Add(1)

		go func(target string) {
			defer wg.Done()
			defer func() { <-sem }()

			alive, rtt := p.Probe(target, timeoutMs)
			if alive && rtt > 0 {
				mu.Lock()
				results[target] = rtt
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	return results, nil
}
