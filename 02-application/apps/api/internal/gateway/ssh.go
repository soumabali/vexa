package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	sshutil "github.com/soumabali/vexa/internal/ssh"
)

type SSHGateway struct {
	mu       sync.RWMutex
	pools    map[string]*SSHPool
	config   *SSHGatewayConfig
}

type SSHGatewayConfig struct {
	MaxConnectionsPerHost int
	ConnectionTimeout     time.Duration
	IdleTimeout           time.Duration
	KeepAliveInterval     time.Duration
}

type SSHPool struct {
	mu          sync.Mutex
	connections []*SSHConnection
	host        string
	port        int
	maxConns    int
}

type SSHConnection struct {
	ID       string
	Client   *ssh.Client
	Host     string
	Port     int
	CreatedAt time.Time
	LastUsed  time.Time
	InUse     bool
}

func NewSSHGateway(config *SSHGatewayConfig) *SSHGateway {
	return &SSHGateway{
		pools:  make(map[string]*SSHPool),
		config: config,
	}
}

func (g *SSHGateway) GetConnection(ctx context.Context, host string, port int, user, password string) (*SSHConnection, error) {
	poolKey := fmt.Sprintf("%s:%d", host, port)

	g.mu.RLock()
	pool, exists := g.pools[poolKey]
	g.mu.RUnlock()

	if !exists {
		pool = &SSHPool{
			host:     host,
			port:     port,
			maxConns: g.config.MaxConnectionsPerHost,
		}
		g.mu.Lock()
		g.pools[poolKey] = pool
		g.mu.Unlock()
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, conn := range pool.connections {
		if !conn.InUse {
			conn.InUse = true
			conn.LastUsed = time.Now().UTC()
			return conn, nil
		}
	}

	if len(pool.connections) >= pool.maxConns {
		return nil, fmt.Errorf("max connections reached for %s", poolKey)
	}

	client, err := g.dial(host, port, user, password)
	if err != nil {
		return nil, err
	}

	conn := &SSHConnection{
		ID:       uuid.New().String(),
		Client:   client,
		Host:     host,
		Port:     port,
		CreatedAt: time.Now().UTC(),
		LastUsed:  time.Now().UTC(),
		InUse:     true,
	}
	pool.connections = append(pool.connections, conn)

	return conn, nil
}

func (g *SSHGateway) dial(host string, port int, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: sshutil.HostKeyCallback(),
		Timeout:         g.config.ConnectionTimeout,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return nil, fmt.Errorf("SSH dial failed: %w", err)
	}

	return client, nil
}

func (g *SSHGateway) ReleaseConnection(conn *SSHConnection) {
	if conn == nil {
		return
	}
	conn.InUse = false
	conn.LastUsed = time.Now().UTC()
}

func (g *SSHGateway) CloseConnection(conn *SSHConnection) {
	if conn == nil {
		return
	}
	conn.Client.Close()
}

func (g *SSHGateway) CleanupIdleConnections(maxIdle time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UTC()
	for key, pool := range g.pools {
		pool.mu.Lock()
		active := make([]*SSHConnection, 0)
		for _, conn := range pool.connections {
			if !conn.InUse && now.Sub(conn.LastUsed) > maxIdle {
				conn.Client.Close()
			} else {
				active = append(active, conn)
			}
		}
		pool.connections = active
		if len(pool.connections) == 0 {
			delete(g.pools, key)
		}
		pool.mu.Unlock()
	}
}

func (g *SSHGateway) StartCleanupRoutine(ctx context.Context, interval, maxIdle time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				g.CleanupIdleConnections(maxIdle)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
