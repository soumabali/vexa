package ssh

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/soumabali/vexa/config"
)

var (
	ErrHostKeyMismatch = errors.New("host key mismatch: possible man-in-the-middle attack")
	ErrHostKeyNotFound = errors.New("host key not found for host")

	globalStore     *KnownHostsStore
	globalStoreErr  error
	globalStoreOnce sync.Once
)

// HostKeyCallback returns a callback using the global known_hosts store.
// This is a convenience for callers that do not have access to a dependency-injected store.
// If no store can be initialized, the callback rejects every connection (fail-close).
func HostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		globalStoreOnce.Do(func() {
			globalStore, globalStoreErr = NewKnownHostsStore(config.SecretsDir())
		})
		if globalStoreErr != nil {
			return fmt.Errorf("host key verification unavailable: %w", globalStoreErr)
		}
		return globalStore.HostKeyCallback(false)(hostname, remote, key)
	}
}

// KnownHostsStore implements Trust-On-First-Use (TOFU) host key verification.
// It stores host keys in a file inside the configured secrets directory.
type KnownHostsStore struct {
	path string
}

// NewKnownHostsStore creates a store at secretsDir/known_hosts.
func NewKnownHostsStore(secretsDir string) (*KnownHostsStore, error) {
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create secrets dir: %w", err)
	}
	path := filepath.Join(secretsDir, "known_hosts")
	// Ensure file exists with restricted permissions
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create known_hosts: %w", err)
	}
	_ = f.Close()
	return &KnownHostsStore{path: path}, nil
}

// HostKeyCallback returns an ssh.HostKeyCallback that verifies or records host keys.
// If strict=true, unknown hosts are rejected. If strict=false, TOFU is used.
func (s *KnownHostsStore) HostKeyCallback(strict bool) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if key == nil {
			return errors.New("host provided no public key")
		}
		addr := hostname
		if addr == "" {
			addr = remote.String()
		}
		// Normalize hostname: strip port for standard SSH port
		host, port, err := net.SplitHostPort(addr)
		if err == nil && port == "22" {
			addr = host
		}
		keyBytes := key.Marshal()
		known, err := s.lookup(addr)
		if err != nil {
			if errors.Is(err, ErrHostKeyNotFound) {
				if strict {
					return fmt.Errorf("%w: %s", ErrHostKeyNotFound, addr)
				}
				// TOFU: record key
				return s.record(addr, key)
			}
			return err
		}
		if !keysEqual(known, keyBytes) {
			return fmt.Errorf("%w for %s", ErrHostKeyMismatch, addr)
		}
		return nil
	}
}

func (s *KnownHostsStore) lookup(addr string) ([]byte, error) {
	f, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		hosts := strings.Split(fields[0], ",")
		for _, h := range hosts {
			if h == addr || matchWildcard(h, addr) {
				decoded, err := base64.StdEncoding.DecodeString(fields[2])
				if err != nil {
					continue
				}
				return decoded, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, ErrHostKeyNotFound
}

func (s *KnownHostsStore) record(addr string, key ssh.PublicKey) error {
	line := fmt.Sprintf("%s %s %s\n", addr, key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}

func keysEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return hmac.Equal(a, b)
}

func matchWildcard(pattern, value string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == value
	}
	parts := strings.Split(pattern, "*")
	if !strings.HasPrefix(value, parts[0]) {
		return false
	}
	value = value[len(parts[0]):]
	for i := 1; i < len(parts); i++ {
		idx := strings.Index(value, parts[i])
		if idx == -1 {
			return false
		}
		if i == len(parts)-1 && !strings.HasSuffix(value, parts[i]) {
			return false
		}
		value = value[idx+len(parts[i]):]
	}
	return true
}

// HostKeyFingerprint returns a SHA-256 fingerprint of a public key.
func HostKeyFingerprint(key ssh.PublicKey) string {
	h := sha256.Sum256(key.Marshal())
	return base64.StdEncoding.EncodeToString(h[:])
}
