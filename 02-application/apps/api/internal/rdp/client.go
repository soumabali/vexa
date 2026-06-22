package rdp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Client manages RDP client connections
type Client struct {
	gateway   *Gateway
	session   *Session
	
	// Connection settings
	reconnectAttempts int
	reconnectDelay    time.Duration
	maxReconnectDelay time.Duration
	
	// State
	mu       sync.RWMutex
	closed   bool
}

// ClientConfig contains RDP client configuration
type ClientConfig struct {
	Hostname          string
	Port              int
	Username          string
	Password          string
	Domain            string
	Width             int
	Height            int
	ColorDepth        int
	ReconnectAttempts int
	ReconnectDelay    time.Duration
}

// NewClient creates a new RDP client
func NewClient(gateway *Gateway, config ClientConfig) (*Client, error) {
	session, err := gateway.CreateSession(
		config.Hostname,
		"",
		config.Hostname,
		config.Port,
		config.Username,
		config.Password,
		config.Domain,
	)
	if err != nil {
		return nil, err
	}
	
	client := &Client{
		gateway:           gateway,
		session:           session,
		reconnectAttempts: config.ReconnectAttempts,
		reconnectDelay:    config.ReconnectDelay,
		maxReconnectDelay: 30 * time.Second,
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
		err := c.session.Connect(ctx)
		if err == nil {
			return nil
		}
		
		attempts++
		if attempts >= c.reconnectAttempts {
			return fmt.Errorf("max reconnect attempts reached: %w", err)
		}
		
		delay := c.reconnectDelay * time.Duration(attempts)
		if delay > c.maxReconnectDelay {
			delay = c.maxReconnectDelay
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			c.session.setStatus(StatusReconnecting)
		}
	}
}

// Disconnect closes the connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	
	return c.gateway.CloseSession(c.session.ID)
}

// SendInput sends input to the RDP session
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

// Resize changes the session resolution
func (c *Client) Resize(width, height int) error {
	return c.session.Resize(width, height)
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
