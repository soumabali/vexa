//go:build ignore

package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"time"
)

// SSHProtocolVersion is the expected SSH-2.0 banner.
const SSHProtocolVersion = "SSH-2.0-OpenSSH_8.9"

// FuzzTarget generates random byte sequences to send to an SSH server.
type FuzzTarget struct {
	Host    string
	Port    string
	Timeout time.Duration
}

// NewFuzzTarget creates a fuzzer pointing at the given address.
func NewFuzzTarget(host, port string) *FuzzTarget {
	return &FuzzTarget{
		Host:    host,
		Port:    port,
		Timeout: 5 * time.Second,
	}
}

// Run executes n fuzzing iterations against the target.
func (ft *FuzzTarget) Run(iterations int) error {
	addr := net.JoinHostPort(ft.Host, ft.Port)
	for i := 0; i < iterations; i++ {
		conn, err := net.DialTimeout("tcp", addr, ft.Timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[fuzz] connection failed: %v\n", err)
			continue
		}
		conn.SetDeadline(time.Now().Add(ft.Timeout))

		// Read server banner.
		banner := make([]byte, 256)
		_, _ = conn.Read(banner)

		// Send random payload.
		payload := ft.randomPayload()
		_, writeErr := conn.Write(payload)
		if writeErr != nil {
			fmt.Fprintf(os.Stderr, "[fuzz] write error: %v\n", writeErr)
		}

		// Read response.
		resp := make([]byte, 1024)
		_, readErr := conn.Read(resp)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "[fuzz] read error (expected): %v\n", readErr)
		}

		conn.Close()
		fmt.Printf("[fuzz] iteration %d: sent %d bytes\n", i+1, len(payload))
	}
	return nil
}

func (ft *FuzzTarget) randomPayload() []byte {
	length := 16 + randInt(512)
	b := make([]byte, length)
	_, _ = rand.Read(b)
	// Occasionally inject SSH-like structures.
	if randInt(10) == 0 {
		prefix := []byte("SSH-2.0-Fuzzer\r\n")
		b = append(prefix, b...)
	}
	return b
}

func randInt(max int) int {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return int(bytes.NewReader(b).Read(b))
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ssh_fuzz.go <host> <port>")
		os.Exit(1)
	}
	ft := NewFuzzTarget(os.Args[1], os.Args[2])
	if err := ft.Run(100); err != nil {
		fmt.Fprintf(os.Stderr, "[fuzz] fatal: %v\n", err)
		os.Exit(1)
	}
}
