package vnc

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Client manages VNC client connections
type Client struct {
	proxy     *Proxy
	session   *Session
	
	// Settings
	reconnectAttempts int
	reconnectDelay    time.Duration
	
	// State
	mu       sync.RWMutex
	closed   bool
}

// ClientConfig contains VNC client configuration
type ClientConfig struct {
	Hostname          string
	Port              int
	Password          string
	ReconnectAttempts int
	ReconnectDelay    time.Duration
}

// NewClient creates a new VNC client
func NewClient(proxy *Proxy, config ClientConfig) (*Client, error) {
	session, err := proxy.CreateSession(
		config.Hostname,
		"",
		config.Hostname,
		config.Port,
		config.Password,
	)
	if err != nil {
		return nil, err
	}
	
	client := &Client{
		proxy:             proxy,
		session:           session,
		reconnectAttempts: config.ReconnectAttempts,
		reconnectDelay:    config.ReconnectDelay,
	}
	
	if client.reconnectAttempts == 0 {
		client.reconnectAttempts = 3
	}
	if client.reconnectDelay == 0 {
		client.reconnectDelay = 5 * time.Second
	}
	
	return client, nil
}

// Connect establishes connection with auto-reconnect
func (c *Client) Connect(ctx context.Context) error {
	attempts := 0
	
	for {
		err := c.session.Connect()
		if err == nil {
			return nil
		}
		
		attempts++
		if attempts >= c.reconnectAttempts {
			return fmt.Errorf("max reconnect attempts reached: %w", err)
		}
		
		delay := c.reconnectDelay * time.Duration(attempts)
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			c.session.setStatus(StatusDisconnected)
		}
	}
}

// Disconnect closes the connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	
	return c.proxy.CloseSession(c.session.ID)
}

// SendInput sends input to the VNC session
func (c *Client) SendInput(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.closed {
		return fmt.Errorf("client is closed")
	}
	
	select {
	case c.session.inputCh <- data:
		return nil
	default:
		return fmt.Errorf("input channel full")
	}
}

// Stats returns connection statistics
func (c *Client) Stats() map[string]interface{} {
	return c.session.Stats()
}

// SessionID returns the session ID
func (c *Client) SessionID() string {
	return c.session.ID
}

// IsConnected returns true if connected
func (c *Client) IsConnected() bool {
	return c.session.Status() == StatusConnected
}

// RequestUpdate requests a framebuffer update
func (c *Client) RequestUpdate(x, y, w, h uint16, incremental bool) error {
	return c.session.requestFramebufferUpdate(x, y, w, h, incremental)
}
