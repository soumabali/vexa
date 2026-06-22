package ssh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"github.com/soumabali/vexa/internal/db"
)

// Message types for binary WebSocket protocol
const (
	MsgData       byte = 0x01
	MsgResize     byte = 0x02
	MsgHeartbeat  byte = 0x03
	MsgError      byte = 0x04
	MsgStatus     byte = 0x05
	MsgSFTP       byte = 0x06
)

// ProxyServer manages SSH connections via WebSocket
type ProxyServer struct {
	db          *db.DB
	upgrader    websocket.Upgrader
	sessions    map[string]*Session
	sessionsMu  sync.RWMutex
	maxSessions int
}

// NewProxyServer creates a new SSH proxy server
func NewProxyServer(database *db.DB) *ProxyServer {
	return &ProxyServer{
		db:       database,
		sessions: make(map[string]*Session),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: restrict to allowed origins
				return true
			},
		},
		maxSessions: 10,
	}
}

// Session represents an active SSH session
type Session struct {
	ID          string
	UserID      string
	HostID      string
	HostAddress string
	SSHConfig   *ssh.ClientConfig
	Client      *ssh.Client
	Session     *ssh.Session
	Stdin       io.WriteCloser
	Stdout      io.Reader
	WSConn      *websocket.Conn
	PTY         *PTY
	CreatedAt   time.Time
	LastActive  time.Time
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// CreateSession creates a new SSH session
func (ps *ProxyServer) CreateSession(userID, hostID string, sshConfig *ssh.ClientConfig) (*Session, error) {
	ps.sessionsMu.Lock()
	defer ps.sessionsMu.Unlock()

	// Check max sessions per user
	userSessionCount := 0
	for _, s := range ps.sessions {
		if s.UserID == userID {
			userSessionCount++
		}
	}
	if userSessionCount >= ps.maxSessions {
		return nil, fmt.Errorf("maximum sessions (%d) reached for user", ps.maxSessions)
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := &Session{
		ID:         uuid.New().String(),
		UserID:     userID,
		HostID:     hostID,
		SSHConfig:  sshConfig,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}

	ps.sessions[session.ID] = session
	return session, nil
}

// Connect establishes SSH connection to host
func (s *Session) Connect(ctx context.Context, hostAddress string) error {
	s.HostAddress = hostAddress

	// Connect to SSH server
	client, err := ssh.Dial("tcp", hostAddress, s.SSHConfig)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	s.Client = client

	// Create SSH session
	sshSession, err := client.NewSession()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	s.Session = sshSession

	// Allocate PTY
	pty, err := NewPTY(80, 24)
	if err != nil {
		return fmt.Errorf("failed to allocate PTY: %w", err)
	}
	s.PTY = pty

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := sshSession.RequestPty("xterm-256color", pty.Width, pty.Height, modes); err != nil {
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	// Get stdin/stdout pipes
	stdin, err := sshSession.StdinPipe()
	if err != nil {
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to get stdin: %w", err)
	}
	s.Stdin = stdin

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to get stdout: %w", err)
	}
	s.Stdout = stdout

	// Start shell
	if err := sshSession.Shell(); err != nil {
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to start shell: %w", err)
	}

	return nil
}

// HandleWebSocket handles WebSocket connection for this session
func (s *Session) HandleWebSocket(ws *websocket.Conn) {
	s.WSConn = ws
	s.LastActive = time.Now()

	// Start goroutines for bidirectional communication
	go s.readFromWebSocket()
	go s.writeToWebSocket()
	go s.heartbeat()
}

// readFromWebSocket reads from WebSocket and writes to SSH stdin
func (s *Session) readFromWebSocket() {
	defer s.Close()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		_, data, err := s.WSConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			return
		}

		if len(data) == 0 {
			continue
		}

		msgType := data[0]
		payload := data[1:]

		switch msgType {
		case MsgData:
			if _, err := s.Stdin.Write(payload); err != nil {
				fmt.Printf("Failed to write to stdin: %v\n", err)
				return
			}
			s.LastActive = time.Now()

		case MsgResize:
			var resize struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			}
			if err := json.Unmarshal(payload, &resize); err != nil {
				fmt.Printf("Failed to unmarshal resize: %v\n", err)
				continue
			}
			s.PTY.Resize(resize.Width, resize.Height)
			if s.Session != nil {
				s.Session.WindowChange(resize.Height, resize.Width)
			}

		case MsgHeartbeat:
			// Respond with heartbeat
			s.sendMessage(MsgHeartbeat, []byte("pong"))

		default:
			fmt.Printf("Unknown message type: %d\n", msgType)
		}
	}
}

// writeToWebSocket reads from SSH stdout and writes to WebSocket
func (s *Session) writeToWebSocket() {
	defer s.Close()

	buf := make([]byte, 4096)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		n, err := s.Stdout.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("SSH stdout error: %v\n", err)
			}
			return
		}

		if n > 0 {
			if err := s.sendMessage(MsgData, buf[:n]); err != nil {
				fmt.Printf("Failed to send to WebSocket: %v\n", err)
				return
			}
			s.LastActive = time.Now()
		}
	}
}

// heartbeat sends periodic ping
func (s *Session) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.sendMessage(MsgHeartbeat, []byte("ping")); err != nil {
				return
			}
		}
	}
}

// sendMessage sends a binary message to WebSocket
func (s *Session) sendMessage(msgType byte, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.WSConn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	// Build binary frame: 1 byte type + payload
	frame := make([]byte, 1+len(payload))
	frame[0] = msgType
	copy(frame[1:], payload)

	return s.WSConn.WriteMessage(websocket.BinaryMessage, frame)
}

// Close closes the session and cleans up
func (s *Session) Close() error {
	s.cancel()

	if s.WSConn != nil {
		s.WSConn.Close()
	}
	if s.Session != nil {
		s.Session.Close()
	}
	if s.Client != nil {
		s.Client.Close()
	}

	return nil
}

// GetSession returns a session by ID
func (ps *ProxyServer) GetSession(id string) (*Session, bool) {
	ps.sessionsMu.RLock()
	defer ps.sessionsMu.RUnlock()
	session, ok := ps.sessions[id]
	return session, ok
}

// ListUserSessions returns all sessions for a user
func (ps *ProxyServer) ListUserSessions(userID string) []*Session {
	ps.sessionsMu.RLock()
	defer ps.sessionsMu.RUnlock()

	var sessions []*Session
	for _, s := range ps.sessions {
		if s.UserID == userID {
			sessions = append(sessions, s)
		}
	}
	return sessions
}

// RemoveSession removes a session
func (ps *ProxyServer) RemoveSession(id string) {
	ps.sessionsMu.Lock()
	defer ps.sessionsMu.Unlock()

	if session, ok := ps.sessions[id]; ok {
		session.Close()
		delete(ps.sessions, id)
	}
}

// CleanupInactive removes sessions inactive for longer than duration
func (ps *ProxyServer) CleanupInactive(maxAge time.Duration) {
	ps.sessionsMu.Lock()
	defer ps.sessionsMu.Unlock()

	now := time.Now()
	for id, session := range ps.sessions {
		if now.Sub(session.LastActive) > maxAge {
			session.Close()
			delete(ps.sessions, id)
		}
	}
}
