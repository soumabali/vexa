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

type VNCGateway struct {
	mu       sync.RWMutex
	sessions map[string]*VNCSession
	config   *VNCGatewayConfig
}

type VNCGatewayConfig struct {
	ConnectionTimeout time.Duration
	MaxSessions       int
	PasswordLength    int
}

type VNCSession struct {
	ID       string
	Host     string
	Port     int
	Password string
	Conn     net.Conn
	WS       *websocket.Conn
	Active   bool
	mu       sync.Mutex
	startedAt time.Time
	bytesIn  int64
	bytesOut int64
}

func NewVNCGateway(config *VNCGatewayConfig) *VNCGateway {
	return &VNCGateway{
		sessions: make(map[string]*VNCSession),
		config:   config,
	}
}

func (g *VNCGateway) CreateSession(ctx context.Context, host string, port int, password string, ws *websocket.Conn) (*VNCSession, error) {
	if port == 0 {
		port = 5900
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), g.config.ConnectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("VNC connection failed: %w", err)
	}

	session := &VNCSession{
		ID:        uuid.New().String(),
		Host:      host,
		Port:      port,
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
		return nil, fmt.Errorf("max VNC sessions reached")
	}
	g.sessions[session.ID] = session
	g.mu.Unlock()

	go g.proxyWebSocketToTCP(session)
	go g.proxyTCPToWebSocket(session)

	return session, nil
}

func (g *VNCGateway) proxyWebSocketToTCP(session *VNCSession) {
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

func (g *VNCGateway) proxyTCPToWebSocket(session *VNCSession) {
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

func (g *VNCGateway) GetSession(id string) (*VNCSession, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	session, exists := g.sessions[id]
	return session, exists
}

func (g *VNCGateway) RemoveSession(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if session, exists := g.sessions[id]; exists {
		session.Close()
		delete(g.sessions, id)
	}
}

func (g *VNCGateway) ListSessions() []*VNCSession {
	g.mu.RLock()
	defer g.mu.RUnlock()
	sessions := make([]*VNCSession, 0, len(g.sessions))
	for _, s := range g.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (s *VNCSession) Close() {
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
