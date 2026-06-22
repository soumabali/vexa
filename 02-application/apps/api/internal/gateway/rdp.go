package gateway

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

type RDPGateway struct {
	mu      sync.RWMutex
	sessions map[string]*RDPSession
	config  *RDPGatewayConfig
}

type RDPGatewayConfig struct {
	ConnectionTimeout time.Duration
	MaxSessions       int
}

type RDPSession struct {
	ID       string
	Host     string
	Port     int
	User     string
	Password string
	Conn     net.Conn
	WS       *websocket.Conn
	Active   bool
	mu       sync.Mutex
	startedAt time.Time
	bytesIn  int64
	bytesOut int64
}

func NewRDPGateway(config *RDPGatewayConfig) *RDPGateway {
	return &RDPGateway{
		sessions: make(map[string]*RDPSession),
		config:   config,
	}
}

func (g *RDPGateway) CreateSession(ctx context.Context, host string, port int, user, password string, ws *websocket.Conn) (*RDPSession, error) {
	if port == 0 {
		port = 3389
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), g.config.ConnectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("RDP connection failed: %w", err)
	}

	session := &RDPSession{
		ID:        uuid.New().String(),
		Host:      host,
		Port:      port,
		User:      user,
		Password:  password,
		Conn:      conn,
		WS:        ws,
		Active:    true,
		startedAt: time.Now().UTC(),
	}

	g.mu.Lock()
	if len(g.sessions) >= g.config.MaxSessions {
		g.mu.Unlock()
		conn.Close()
		return nil, fmt.Errorf("max RDP sessions reached")
	}
	g.sessions[session.ID] = session
	g.mu.Unlock()

	go g.proxyWebSocketToTCP(session)
	go g.proxyTCPToWebSocket(session)

	return session, nil
}

func (g *RDPGateway) proxyWebSocketToTCP(session *RDPSession) {
	defer session.Close()
	for {
		messageType, data, err := session.WS.ReadMessage()
		if err != nil {
			return
		}
		if messageType == websocket.BinaryMessage {
			session.mu.Lock()
			_, err := session.Conn.Write(data)
			session.bytesIn += int64(len(data))
			session.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (g *RDPGateway) proxyTCPToWebSocket(session *RDPSession) {
	defer session.Close()
	buf := make([]byte, 4096)
	for {
		n, err := session.Conn.Read(buf)
		if err != nil {
			return
		}
		if n > 0 {
			session.mu.Lock()
			err := session.WS.WriteMessage(websocket.BinaryMessage, buf[:n])
			session.bytesOut += int64(n)
			session.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (g *RDPGateway) GetSession(id string) (*RDPSession, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	session, exists := g.sessions[id]
	return session, exists
}

func (g *RDPGateway) RemoveSession(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if session, exists := g.sessions[id]; exists {
		session.Close()
		delete(g.sessions, id)
	}
}

func (g *RDPGateway) ListSessions() []*RDPSession {
	g.mu.RLock()
	defer g.mu.RUnlock()
	sessions := make([]*RDPSession, 0, len(g.sessions))
	for _, s := range g.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (s *RDPSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.Active {
		return
	}
	s.Active = false
	if s.Conn != nil {
		s.Conn.Close()
	}
	if s.WS != nil {
		s.WS.Close()
	}
}
