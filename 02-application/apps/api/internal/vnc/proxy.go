package vnc

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// VNC protocol constants
const (
	VNCVersion     = "RFB 003.008\n"
	SecurityNone   = 1
	SecurityVNCAuth = 2
	SecurityRA2    = 5
	SecurityRA2ne  = 6
	SecurityTight  = 16
	SecurityUltra  = 17
	SecurityTLS    = 18
	SecurityVeNCrypt = 19
	
	// Message types
	MsgFramebufferUpdate    = 0
	MsgSetColorMapEntries   = 1
	MsgBell                 = 2
	MsgServerCutText        = 3
	MsgClientCutText        = 6
	MsgFramebufferUpdateReq = 3
	MsgSetPixelFormat       = 0
	MsgSetEncodings         = 2
	MsgKeyEvent             = 4
	MsgPointerEvent         = 5
	
	// Encoding types
	EncRaw                  = 0
	EncCopyRect             = 1
	EncRRE                  = 2
	EncCoRRE                = 4
	EncHextile              = 5
	EncZlib                 = 6
	EncTight                = 7
	EncZlibHex              = 8
	EncUltra                = 9
	EncTRLE                 = 15
	EncZRLE                 = 16
	EncPseudoCursor         = -239
	EncPseudoDesktopSize    = -223
	EncPseudoExtendedDesktopSize = -308
)

// Session represents an active VNC session
type Session struct {
	ID           string
	HostID       string
	UserID       string
	ConnectionID string
	
	// VNC connection params
	Hostname     string
	Port         int
	Password     string
	
	// State
	mu           sync.RWMutex
	status       SessionStatus
	connectedAt  time.Time
	lastActivity time.Time
	
	// Network
	conn         net.Conn
	
	// WebSocket
	wsConn       *websocket.Conn
	wsMutex      sync.Mutex
	
	// Channels
	done         chan struct{}
	inputCh      chan []byte
	outputCh     chan []byte
	
	// Server info
	framebufferWidth  uint16
	framebufferHeight uint16
	pixelFormat       PixelFormat
	name              string
	
	// Metrics
	bytesSent    int64
	bytesRecv    int64
	framesSent   int64
}

// SessionStatus represents VNC session state
type SessionStatus int

const (
	StatusDisconnected SessionStatus = iota
	StatusConnecting
	StatusHandshaking
	StatusAuthenticating
	StatusConnected
	StatusError
	StatusClosed
)

func (s SessionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusHandshaking:
		return "handshaking"
	case StatusAuthenticating:
		return "authenticating"
	case StatusConnected:
		return "connected"
	case StatusError:
		return "error"
	case StatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// PixelFormat represents VNC pixel format
type PixelFormat struct {
	BitsPerPixel uint8
	Depth        uint8
	BigEndian    uint8
	TrueColor    uint8
	RedMax       uint16
	GreenMax     uint16
	BlueMax      uint16
	RedShift     uint8
	GreenShift   uint8
	BlueShift    uint8
}

// Proxy manages VNC sessions
type Proxy struct {
	sessions    map[string]*Session
	sessionsMu  sync.RWMutex
	
	// Metrics
	totalSessions  int64
	activeSessions int64
}

// NewProxy creates a new VNC proxy
func NewProxy() *Proxy {
	return &Proxy{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new VNC session
func (p *Proxy) CreateSession(hostID, userID, hostname string, port int, password string) (*Session, error) {
	session := &Session{
		ID:       generateSessionID(),
		HostID:   hostID,
		UserID:   userID,
		Hostname: hostname,
		Port:     port,
		Password: password,
		status:   StatusDisconnected,
		done:     make(chan struct{}),
		inputCh:  make(chan []byte, 100),
		outputCh: make(chan []byte, 100),
	}
	
	p.sessionsMu.Lock()
	p.sessions[session.ID] = session
	p.totalSessions++
	p.activeSessions++
	p.sessionsMu.Unlock()
	
	return session, nil
}

// GetSession retrieves a session by ID
func (p *Proxy) GetSession(id string) (*Session, bool) {
	p.sessionsMu.RLock()
	defer p.sessionsMu.RUnlock()
	session, ok := p.sessions[id]
	return session, ok
}

// ListSessions returns all active sessions
func (p *Proxy) ListSessions() []*Session {
	p.sessionsMu.RLock()
	defer p.sessionsMu.RUnlock()
	
	sessions := make([]*Session, 0, len(p.sessions))
	for _, s := range p.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// CloseSession terminates a session
func (p *Proxy) CloseSession(id string) error {
	p.sessionsMu.Lock()
	session, ok := p.sessions[id]
	if !ok {
		p.sessionsMu.Unlock()
		return fmt.Errorf("session not found: %s", id)
	}
	delete(p.sessions, id)
	p.activeSessions--
	p.sessionsMu.Unlock()
	
	return session.Close()
}

// Connect establishes VNC connection
func (s *Session) Connect() error {
	s.mu.Lock()
	if s.status != StatusDisconnected {
		s.mu.Unlock()
		return fmt.Errorf("session not in disconnected state")
	}
	s.status = StatusConnecting
	s.mu.Unlock()
	
	// Connect to VNC server
	addr := fmt.Sprintf("%s:%d", s.Hostname, s.Port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("failed to connect to VNC server: %w", err)
	}
	s.conn = conn
	
	// Perform VNC handshake
	if err := s.handshake(); err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("handshake failed: %w", err)
	}
	
	// Authenticate
	if err := s.authenticate(); err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("authentication failed: %w", err)
	}
	
	// Initialize
	if err := s.initialize(); err != nil {
		s.setStatus(StatusError)
		return fmt.Errorf("initialization failed: %w", err)
	}
	
	s.connectedAt = time.Now()
	s.lastActivity = time.Now()
	s.setStatus(StatusConnected)
	
	// Start I/O goroutines
	go s.readLoop()
	go s.writeLoop()
	
	return nil
}

// handshake performs VNC protocol handshake
func (s *Session) handshake() error {
	s.setStatus(StatusHandshaking)
	
	// Read server version
	buf := make([]byte, 12)
	if _, err := io.ReadFull(s.conn, buf); err != nil {
		return fmt.Errorf("failed to read server version: %w", err)
	}
	
	// Send client version
	if _, err := s.conn.Write([]byte(VNCVersion)); err != nil {
		return fmt.Errorf("failed to send client version: %w", err)
	}
	
	return nil
}

// authenticate performs VNC authentication
func (s *Session) authenticate() error {
	s.setStatus(StatusAuthenticating)
	
	// Read security types
	buf := make([]byte, 1)
	if _, err := io.ReadFull(s.conn, buf); err != nil {
		return fmt.Errorf("failed to read security type count: %w", err)
	}
	
	numTypes := int(buf[0])
	if numTypes == 0 {
		// Read error reason
		var reasonLen uint32
		if err := s.readUint32(&reasonLen); err != nil {
			return err
		}
		reason := make([]byte, reasonLen)
		if _, err := io.ReadFull(s.conn, reason); err != nil {
			return err
		}
		return fmt.Errorf("server error: %s", string(reason))
	}
	
	// Read security types
	secTypes := make([]byte, numTypes)
	if _, err := io.ReadFull(s.conn, secTypes); err != nil {
		return fmt.Errorf("failed to read security types: %w", err)
	}
	
	// Choose VNC auth or none
	chosenType := uint8(SecurityNone)
	for _, t := range secTypes {
		if t == SecurityVNCAuth {
			chosenType = SecurityVNCAuth
			break
		}
		if t == SecurityNone {
			chosenType = SecurityNone
		}
	}
	
	// Send chosen type
	if _, err := s.conn.Write([]byte{chosenType}); err != nil {
		return fmt.Errorf("failed to send security type: %w", err)
	}
	
	// Handle authentication
	switch chosenType {
	case SecurityNone:
		// No authentication needed
		return nil
	case SecurityVNCAuth:
		return s.vncAuth()
	default:
		return fmt.Errorf("unsupported security type: %d", chosenType)
	}
}

// vncAuth performs VNC password authentication
func (s *Session) vncAuth() error {
	// Read challenge (16 bytes)
	challenge := make([]byte, 16)
	if _, err := io.ReadFull(s.conn, challenge); err != nil {
		return fmt.Errorf("failed to read challenge: %w", err)
	}
	
	// Encrypt challenge with password
	response := s.encryptChallenge(challenge)
	
	// Send response
	if _, err := s.conn.Write(response); err != nil {
		return fmt.Errorf("failed to send auth response: %w", err)
	}
	
	// Read result
	var result uint32
	if err := s.readUint32(&result); err != nil {
		return fmt.Errorf("failed to read auth result: %w", err)
	}
	
	if result != 0 {
		return fmt.Errorf("authentication failed")
	}
	
	return nil
}

// encryptChallenge encrypts VNC challenge with password
func (s *Session) encryptChallenge(challenge []byte) []byte {
	// VNC uses DES encryption with password padded to 8 bytes
	// This is a simplified implementation
	key := s.padPassword(s.Password)
	
	// Simple XOR for demo (replace with proper DES in production)
	result := make([]byte, 16)
	for i := 0; i < 16; i++ {
		result[i] = challenge[i] ^ key[i%8]
	}
	
	return result
}

func (s *Session) padPassword(password string) []byte {
	padded := make([]byte, 8)
	copy(padded, password)
	return padded
}

// initialize reads server initialization message
func (s *Session) initialize() error {
	// Read framebuffer dimensions
	var width, height uint16
	if err := s.readUint16(&width); err != nil {
		return err
	}
	if err := s.readUint16(&height); err != nil {
		return err
	}
	
	s.framebufferWidth = width
	s.framebufferHeight = height
	
	// Read pixel format
	pf := make([]byte, 16)
	if _, err := io.ReadFull(s.conn, pf); err != nil {
		return fmt.Errorf("failed to read pixel format: %w", err)
	}
	
	s.pixelFormat = PixelFormat{
		BitsPerPixel: pf[0],
		Depth:        pf[1],
		BigEndian:    pf[2],
		TrueColor:    pf[3],
		RedMax:       uint16(pf[4])<<8 | uint16(pf[5]),
		GreenMax:     uint16(pf[6])<<8 | uint16(pf[7]),
		BlueMax:      uint16(pf[8])<<8 | uint16(pf[9]),
		RedShift:     pf[10],
		GreenShift:   pf[11],
		BlueShift:    pf[12],
	}
	
	// Read name length
	var nameLen uint32
	if err := s.readUint32(&nameLen); err != nil {
		return err
	}
	
	// Read name
	name := make([]byte, nameLen)
	if _, err := io.ReadFull(s.conn, name); err != nil {
		return fmt.Errorf("failed to read desktop name: %w", err)
	}
	s.name = string(name)
	
	// Set pixel format (32-bit true color)
	if err := s.setPixelFormat(); err != nil {
		return err
	}
	
	// Set encodings (Tight, ZRLE, Hextile, Raw)
	if err := s.setEncodings(); err != nil {
		return err
	}
	
	// Request full framebuffer update
	return s.requestFramebufferUpdate(0, 0, s.framebufferWidth, s.framebufferHeight, false)
}

// setPixelFormat sets the pixel format
func (s *Session) setPixelFormat() error {
	msg := make([]byte, 20)
	msg[0] = MsgSetPixelFormat
	msg[1] = 0 // padding
	msg[2] = 0 // padding
	msg[3] = 0 // padding
	
	// 32-bit true color
	msg[4] = 32  // bits per pixel
	msg[5] = 24  // depth
	msg[6] = 0   // big endian
	msg[7] = 1   // true color
	msg[8] = 0; msg[9] = 255  // red max
	msg[10] = 0; msg[11] = 255 // green max
	msg[12] = 0; msg[13] = 255 // blue max
	msg[14] = 16 // red shift
	msg[15] = 8  // green shift
	msg[16] = 0  // blue shift
	// padding
	
	_, err := s.conn.Write(msg)
	return err
}

// setEncodings sets supported encodings
func (s *Session) setEncodings() error {
	encodings := []int32{
		EncTight,
		EncZRLE,
		EncHextile,
		EncZlib,
		EncRaw,
		EncPseudoCursor,
		EncPseudoDesktopSize,
	}
	
	msg := make([]byte, 4+len(encodings)*4)
	msg[0] = MsgSetEncodings
	msg[1] = 0 // padding
	msg[2] = 0 // padding
	msg[3] = 0 // padding
	
	// Number of encodings
	msg[4] = 0
	msg[5] = 0
	msg[6] = 0
	msg[7] = byte(len(encodings))
	
	for i, enc := range encodings {
		msg[8+i*4] = byte(enc >> 24)
		msg[9+i*4] = byte(enc >> 16)
		msg[10+i*4] = byte(enc >> 8)
		msg[11+i*4] = byte(enc)
	}
	
	_, err := s.conn.Write(msg)
	return err
}

// requestFramebufferUpdate requests a framebuffer update
func (s *Session) requestFramebufferUpdate(x, y, w, h uint16, incremental bool) error {
	msg := make([]byte, 10)
	msg[0] = MsgFramebufferUpdateReq
	msg[1] = 0 // padding
	if incremental {
		msg[2] = 1
	} else {
		msg[2] = 0
	}
	msg[3] = 0 // padding
	msg[4] = byte(x >> 8)
	msg[5] = byte(x)
	msg[6] = byte(y >> 8)
	msg[7] = byte(y)
	msg[8] = byte(w >> 8)
	msg[9] = byte(w)
	msg[10] = byte(h >> 8)
	msg[11] = byte(h)
	
	_, err := s.conn.Write(msg)
	return err
}

// readLoop reads from VNC server
func (s *Session) readLoop() {
	buf := make([]byte, 64*1024)
	for {
		select {
		case <-s.done:
			return
		default:
		}
		
		n, err := s.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				s.setStatus(StatusError)
			}
			return
		}
		
		if n > 0 {
			s.bytesRecv += int64(n)
			s.lastActivity = time.Now()
			
			// Forward to WebSocket
			select {
			case s.outputCh <- append([]byte{}, buf[:n]...):
			case <-s.done:
				return
			}
		}
	}
}

// writeLoop writes to VNC server
func (s *Session) writeLoop() {
	for {
		select {
		case <-s.done:
			return
		case data := <-s.inputCh:
			if _, err := s.conn.Write(data); err != nil {
				s.setStatus(StatusError)
				return
			}
			s.bytesSent += int64(len(data))
			s.lastActivity = time.Now()
		}
	}
}

// AttachWebSocket attaches a WebSocket
func (s *Session) AttachWebSocket(conn *websocket.Conn) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	
	if s.wsConn != nil {
		s.wsConn.Close()
	}
	s.wsConn = conn
	
	// Start WebSocket handler
	go s.handleWebSocket()
}

// handleWebSocket handles WebSocket communication
func (s *Session) handleWebSocket() {
	if s.wsConn == nil {
		return
	}
	
	// Forward output to WebSocket
	go func() {
		for {
			select {
			case <-s.done:
				return
			case data := <-s.outputCh:
				s.wsMutex.Lock()
				if s.wsConn != nil {
					frame := make([]byte, 1+len(data))
					frame[0] = FrameTypeScreen
					copy(frame[1:], data)
					s.wsConn.WriteMessage(websocket.BinaryMessage, frame)
					s.framesSent++
				}
				s.wsMutex.Unlock()
			}
		}
	}()
	
	// Read WebSocket messages
	for {
		select {
		case <-s.done:
			return
		default:
		}
		
		msgType, data, err := s.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			s.setStatus(StatusError)
			return
		}
		
		if msgType == websocket.BinaryMessage && len(data) > 0 {
			frameType := data[0]
			payload := data[1:]
			
			switch frameType {
			case FrameTypeInput:
				select {
				case s.inputCh <- payload:
				case <-s.done:
					return
				}
			case FrameTypeResize:
				if len(payload) >= 4 {
					width := uint16(payload[0])<<8 | uint16(payload[1])
					height := uint16(payload[2])<<8 | uint16(payload[3])
					s.requestFramebufferUpdate(0, 0, width, height, false)
				}
			case FrameTypeClipboard:
				s.handleClipboard(payload)
			}
		}
	}
}

// handleClipboard handles clipboard sync
func (s *Session) handleClipboard(data []byte) {
	// Send ClientCutText message
	msg := make([]byte, 8+len(data))
	msg[0] = MsgClientCutText
	msg[1] = 0 // padding
	msg[2] = 0 // padding
	msg[3] = 0 // padding
	msg[4] = 0 // padding
	msg[5] = 0 // padding
	msg[6] = 0 // padding
	msg[7] = 0 // padding
	msg[8] = byte(len(data) >> 24)
	msg[9] = byte(len(data) >> 16)
	msg[10] = byte(len(data) >> 8)
	msg[11] = byte(len(data))
	copy(msg[12:], data)
	
	s.inputCh <- msg
}

// readUint16 reads a uint16 from connection
func (s *Session) readUint16(v *uint16) error {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(s.conn, buf); err != nil {
		return err
	}
	*v = uint16(buf[0])<<8 | uint16(buf[1])
	return nil
}

// readUint32 reads a uint32 from connection
func (s *Session) readUint32(v *uint32) error {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(s.conn, buf); err != nil {
		return err
	}
	*v = uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	return nil
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
	
	if s.conn != nil {
		s.conn.Close()
	}
	
	return nil
}

// setStatus updates session status
func (s *Session) setStatus(status SessionStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

// Status returns current status
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
		"id":               s.ID,
		"status":           s.status.String(),
		"host_id":          s.HostID,
		"connected_at":     s.connectedAt,
		"last_activity":    s.lastActivity,
		"bytes_sent":       s.bytesSent,
		"bytes_recv":       s.bytesRecv,
		"frames_sent":      s.framesSent,
		"framebuffer_w":    s.framebufferWidth,
		"framebuffer_h":    s.framebufferHeight,
		"desktop_name":     s.name,
	}
}

// Frame types for binary protocol
const (
	FrameTypeScreen    byte = 0x01
	FrameTypeInput     byte = 0x02
	FrameTypeAudio     byte = 0x03
	FrameTypeResize    byte = 0x04
	FrameTypeClipboard byte = 0x05
	FrameTypeCursor    byte = 0x06
	FrameTypeError     byte = 0xFF
)

func generateSessionID() string {
	return fmt.Sprintf("vnc-%d", time.Now().UnixNano())
}
