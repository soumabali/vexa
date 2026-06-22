package discovery

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// TCPProber handles TCP port scanning
type TCPProber struct{}

// NewTCPProber creates a new TCP prober
func NewTCPProber() *TCPProber {
	return &TCPProber{}
}

// IsOpen checks if a TCP port is open
func (p *TCPProber) IsOpen(ip string, port int, timeoutMs int) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GrabBanner attempts to grab a service banner
func (p *TCPProber) GrabBanner(ip string, port int, timeoutMs int) string {
	target := fmt.Sprintf("%s:%d", ip, port)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return ""
	}
	defer conn.Close()

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(timeout))

	// For SSH: just read the banner the server sends
	// For other protocols: might need to send a probe first
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return ""
	}

	banner := string(buf[:n])
	// Truncate and sanitize
	if len(banner) > 256 {
		banner = banner[:256] + "..."
	}
	return banner
}

// IsTLSEnabled checks if the port speaks TLS
func (p *TCPProber) IsTLSEnabled(ip string, port int, timeoutMs int) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: timeout},
		"tcp", target, tlsConfig,
	)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// ProbePort scans a single port with banner grab
func (p *TCPProber) ProbePort(ip string, port int, timeoutMs int) PortResult {
	result := PortResult{
		Port:     port,
		Protocol: "tcp",
	}

	if service, ok := KnownServices[port]; ok {
		result.Service = service
	}

	if !p.IsOpen(ip, port, timeoutMs) {
		return result
	}

	// Grab banner for known services
	if port == 22 || port == 3389 || port == 5900 || port == 80 || port == 8080 {
		banner := p.GrabBanner(ip, port, timeoutMs)
		if banner != "" {
			result.Banner = banner
		}
	}

	// Check TLS
	if port == 443 || port == 8443 {
		result.TLS = p.IsTLSEnabled(ip, port, timeoutMs)
	}

	return result
}
