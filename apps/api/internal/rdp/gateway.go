package rdp

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

// Session represents an active RDP session
type Session struct {
	ID           string
	HostID       string
	UserID       string
	ConnectionID string
	
	// RDP connection params
	Hostname     string
	Port         int
	Username     string
	Password     string
	Domain       string
	Width        int
	Height       int
	ColorDepth   int
	
	// Runtime
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	
	// WebSocket
	wsConn       *websocket.Conn
	wsMutex      sync.Mutex
	
	// State
	mu           sync.RWMutex
	status       SessionStatus
	connectedAt  time.Time
	lastActivity time.Time
	
	// Channels
	done         chan struct{}
	inputCh      chan []byte
	outputCh     chan []byte
	
	// Metrics
	bytesSent    int64
	bytesRecv    int64
	framesSent   int64
}

// SessionStatus represents RDP session state
type SessionStatus int

const (
	StatusDisconnected SessionStatus = iota
	StatusConnecting
	StatusConnected
	StatusReconnecting
	StatusError
	StatusClosed
)

func (s SessionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusReconnecting:
		return "reconnecting"
	case StatusError:
		return "error"
	case StatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Gateway manages RDP sessions
type Gateway struct {
	sessions    map[string]*Session
	sessionsMu  sync.RWMutex
	
	// Configuration
	freerdpPath string
	defaultWidth  int
	defaultHeight int
	defaultDepth  int
	
	// Metrics
	totalSessions int64
	activeSessions int64
}

// NewGateway creates a new RDP gateway
func NewGateway() *Gateway {
	return &Gateway{
		sessions:      make(map[string]*Session),
		freerdpPath:   "xfreerdp3",
		defaultWidth:  1920,
		defaultHeight: 1080,
		defaultDepth:  32,
	}
}

// SetFreeRDPPath sets the FreeRDP binary path
func (g *Gateway) SetFreeRDPPath(path string) {
	g.freerdpPath = path
}

// CreateSession creates a new RDP session
func (g *Gateway) CreateSession(hostID, userID, hostname string, port int, username, password, domain string) (*Session, error) {
	session := &Session{
		ID:           generateSessionID(),
		HostID:       hostID,
		UserID:       userID,
		Hostname:     hostname,
		Port:         port,
		Username:     username,
		Password:     password,
		Domain:       domain,
		Width:        g.defaultWidth,
		Height:       g.defaultHeight,
		ColorDepth:   g.defaultDepth,
		status:       StatusDisconnected,
		done:         make(chan struct{}),
		inputCh:      make(chan []byte, 100),
		outputCh:     make(chan []byte, 100),
	}
	
	g.sessionsMu.Lock()
	g.sessions[session.ID] = session
	g.totalSessions++
	g.activeSessions++
	g.sessionsMu.Unlock()
	
	return session, nil
}

// GetSession retrieves a session by ID
func (g *Gateway) GetSession(id string) (*Session, bool) {
	g.sessionsMu.RLock()
	defer g.sessionsMu.RUnlock()
	session, ok := g.sessions[id]
	return session, ok
}

// ListSessions returns all active sessions
func (g *Gateway) ListSessions() []*Session {
	g.sessionsMu.RLock()
	defer g.sessionsMu.RUnlock()
	
	sessions := make([]*Session, 0, len(g.sessions))
	for _, s := range g.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// CloseSession terminates a session
func (g *Gateway) CloseSession(id string) error {
	g.sessionsMu.Lock()
	session, ok := g.sessions[id]
	if !ok {
		g.sessionsMu.Unlock()
		return fmt.Errorf("session not found: %s", id)
	}
	delete(g.sessions, id)
	g.activeSessions--
	g.sessionsMu.Unlock()
	
	return session.Close()
}

// Connect establishes RDP connection via FreeRDP
func (s *Session) Connect(ctx context.Context) error {
	s.mu.Lock()
	if s.status != StatusDisconnected {
		s.mu.Unlock()
		return fmt.Errorf("session not in disconnected state: %s", s.status)
	}
	s.status = StatusConnecting
	s.mu.Unlock()
	
	// Build FreeRDP command
	args := []string{
		"/v:" + s.Hostname + ":" + fmt.Sprintf("%d", s.Port),
		"/u:" + s.Username,
		"/p:" + s.Password,
		"/w:" + fmt.Sprintf("%d", s.Width),
		"/h:" + fmt.Sprintf("%d", s.Height),
		"/bpp:" + fmt.Sprintf("%d", s.ColorDepth),
		"/cert:ignore",
		"/gfx",
		"/gfx:AVC444",
		"/network:autodetect",
		"/compression",
		"/clipboard",
		"/audio:sys:pulse",
		"/mic:sys:pulse",
		"+fonts",
		"+aero",
		"+window-drag",
		"+menu-anims",
		"+themes",
		"/dynamic-resolution",
		"/echo",
		"/from-stdin", // Enable stdin for input forwarding
	}
	
	if s.Domain != "" {
		args = append(args, "/d:"+s.Domain)
	}
	
	cmd := exec.CommandContext(ctx, "xfreerdp3", args...)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	s.cmd = cmd
	s.stdin = stdin
	s.stdout = stdout
	s.stderr = stderr
	
	if err := cmd.Start(); err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("failed to start FreeRDP: %w", err)
	}
	
	s.connectedAt = time.Now()
	s.lastActivity = time.Now()
	s.setStatus(StatusConnected)
	
	// Start I/O goroutines
	g, ctx := errgroup.WithContext(ctx)
	
	// Read output from FreeRDP
	g.Go(func() error {
		return s.readOutput(ctx)
	})
	
	// Write input to FreeRDP
	g.Go(func() error {
		return s.writeInput(ctx)
	})
	
	// Handle WebSocket
	g.Go(func() error {
		return s.handleWebSocket(ctx)
	})
	
	// Monitor process
	g.Go(func() error {
		return s.monitorProcess(ctx)
	})
	
	go func() {
		if err := g.Wait(); err != nil {
			s.setStatus(StatusError)
		}
		s.Close()
	}()
	
	return nil
}

// AttachWebSocket attaches a WebSocket connection to the session
func (s *Session) AttachWebSocket(conn *websocket.Conn) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	
	if s.wsConn != nil {
		s.wsConn.Close()
	}
	s.wsConn = conn
}

// SendFrame sends a frame to the WebSocket
func (s *Session) SendFrame(frameType byte, data []byte) error {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	
	if s.wsConn == nil {
		return fmt.Errorf("no websocket attached")
	}
	
	// Binary frame: [1 byte type][payload]
	frame := make([]byte, 1+len(data))
	frame[0] = frameType
	copy(frame[1:], data)
	
	s.wsMutex.Unlock()
	err := s.wsConn.WriteMessage(websocket.BinaryMessage, frame)
	s.wsMutex.Lock()
	
	if err != nil {
		return err
	}
	
	s.framesSent++
	s.bytesSent += int64(len(frame))
	return nil
}

// readOutput reads from FreeRDP stdout
func (s *Session) readOutput(ctx context.Context) error {
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		default:
		}
		
		n, err := s.stdout.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		
		if n > 0 {
			s.bytesRecv += int64(n)
			s.lastActivity = time.Now()
			
			// Forward to WebSocket
			select {
			case s.outputCh <- append([]byte{}, buf[:n]...):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// writeInput writes to FreeRDP stdin
func (s *Session) writeInput(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		case data := <-s.inputCh:
			if _, err := s.stdin.Write(data); err != nil {
				return err
			}
			s.bytesSent += int64(len(data))
			s.lastActivity = time.Now()
		}
	}
}

// handleWebSocket handles WebSocket messages
func (s *Session) handleWebSocket(ctx context.Context) error {
	if s.wsConn == nil {
		return fmt.Errorf("no websocket connection")
	}
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		default:
		}
		
		msgType, data, err := s.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			return err
		}
		
		if msgType == websocket.BinaryMessage && len(data) > 0 {
			frameType := data[0]
			payload := data[1:]
			
			switch frameType {
			case FrameTypeInput:
				// Forward input to FreeRDP
				select {
				case s.inputCh <- payload:
				case <-ctx.Done():
					return ctx.Err()
				}
			case FrameTypeResize:
				// Handle resize
				if len(payload) >= 4 {
					width := int(payload[0])<<8 | int(payload[1])
					height := int(payload[2])<<8 | int(payload[3])
					s.Resize(width, height)
				}
			case FrameTypeClipboard:
				// Handle clipboard sync
				s.handleClipboard(payload)
			}
		}
	}
}

// monitorProcess monitors the FreeRDP process
func (s *Session) monitorProcess(ctx context.Context) error {
	if s.cmd == nil {
		return fmt.Errorf("no command running")
	}
	
	err := s.cmd.Wait()
	if err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("freerdp process exited: %w", err)
	}
	
	s.setStatus(StatusDisconnected)
	return nil
}

// Resize changes the session resolution
func (s *Session) Resize(width, height int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.Width = width
	s.Height = height
	
	// Send resize command to FreeRDP via stdin
	resizeCmd := fmt.Sprintf("/size:%dx%d\n", width, height)
	if s.stdin != nil {
		_, err := s.stdin.Write([]byte(resizeCmd))
		return err
	}
	
	return nil
}

// handleClipboard handles clipboard sync
func (s *Session) handleClipboard(data []byte) {
	// TODO: Implement clipboard sync
	// Forward to FreeRDP clipboard channel
}

// Close closes the session
func (s *Session) Close() error {
	s.mu.Lock()
	if s.status == StatusClosed {
		s.mu.Unlock()
		return nil
	}
	s.status = StatusClosed
	s.mu.Unlock()
	
	close(s.done)
	
	if s.wsConn != nil {
		s.wsConn.Close()
	}
	
	if s.stdin != nil {
		s.stdin.Close()
	}
	
	if s.stdout != nil {
		s.stdout.Close()
	}
	
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
	
	return nil
}

// setStatus updates session status
func (s *Session) setStatus(status SessionStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

// Status returns current session status
func (s *Session) Status() SessionStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// Stats returns session statistics
func (s *Session) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return map[string]interface{}{
		"id":            s.ID,
		"status":        s.status.String(),
		"host_id":       s.HostID,
		"connected_at":  s.connectedAt,
		"last_activity": s.lastActivity,
		"bytes_sent":    s.bytesSent,
		"bytes_recv":    s.bytesRecv,
		"frames_sent":   s.framesSent,
	}
}

// Frame types for binary WebSocket protocol
const (
	FrameTypeScreen  byte = 0x01 // Screen update
	FrameTypeInput   byte = 0x02 // Input event
	FrameTypeAudio   byte = 0x03 // Audio data
	FrameTypeResize  byte = 0x04 // Resize request
	FrameTypeClipboard byte = 0x05 // Clipboard data
	FrameTypeCursor  byte = 0x06 // Cursor update
	FrameTypeError   byte = 0xFF // Error
)

func generateSessionID() string {
	return fmt.Sprintf("rdp-%d", time.Now().UnixNano())
}
