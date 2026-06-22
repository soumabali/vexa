package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	"github.com/soumabali/vexa/internal/models"
	sshutil "github.com/soumabali/vexa/internal/ssh"
)

var (
	ErrSessionNotFound = errors.New("terminal session not found")
	ErrInvalidMessage  = errors.New("invalid websocket message")
	maxInputLength     = 4096
	ErrInputTooLong    = errors.New("terminal input exceeds maximum length")
)

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

type Session struct {
	ID          string
	UserID      uuid.UUID
	HostID      uuid.UUID
	Protocol    string
	WebSocket   *websocket.Conn
	SSHClient   *ssh.Client
	SSHSession  *ssh.Session
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	mu          sync.Mutex
	active      bool
	bytesIn     int64
	bytesOut    int64
	startedAt   time.Time
	lastActivity time.Time
}

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage  `json:"data"`
}

type InitData struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

	type ResizeData struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// InputValidator defines the interface for terminal input validation
type InputValidator interface {
	Validate(data string) error
}

// defaultValidator is the global default validator (can be overridden for testing)
var defaultValidator InputValidator = &defaultInputValidator{maxLen: maxInputLength}

// DefaultValidator returns the global default validator (exported for testing)
func DefaultValidator() InputValidator {
	return defaultValidator
}

type defaultInputValidator struct {
	maxLen int
}

func (v *defaultInputValidator) Validate(data string) error {
	if len(data) > v.maxLen {
		return fmt.Errorf("%w: %d > %d", ErrInputTooLong, len(data), v.maxLen)
	}
	// Reject newlines and null bytes
	if strings.ContainsAny(data, "\n\r\x00") {
		return fmt.Errorf("input contains forbidden control characters")
	}
	// Reject dangerous shell metacharacters
	forbidden := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "\u003c", "\u003e", "!", "*", "?", "~", "#", "%", "\\"}
	for _, char := range forbidden {
		if strings.Contains(data, char) {
			return fmt.Errorf("input contains forbidden character: %s", char)
		}
	}
	// Reject command words that are commonly used for destructive operations.
	lower := strings.ToLower(data)
	dangerousWords := []string{"rm -rf", "mkfs", "dd if=/dev/zero", ":(){ :|:", "../etc/shadow", "../etc/passwd"}
	for _, word := range dangerousWords {
		if strings.Contains(lower, word) {
			return fmt.Errorf("input contains dangerous command: %s", word)
		}
	}
	return nil
}

func SanitizeInput(input string) string {
	// Remove binary control characters
	replacer := strings.NewReplacer(
		"\x00", "",
		"\x01", "", "\x02", "", "\x03", "", "\x04", "", "\x05", "", "\x06", "", "\x07", "",
		"\x08", "", "\x09", "", "\x0a", "", "\x0b", "", "\x0c", "", "\x0d", "", "\x0e", "",
		"\x0f", "", "\x10", "", "\x11", "", "\x12", "", "\x13", "", "\x14", "", "\x15", "",
		"\x16", "", "\x17", "", "\x18", "", "\x19", "", "\x1a", "", "\x1b", "", "\x1c", "",
		"\x1d", "", "\x1e", "", "\x1f", "",
	)
	input = replacer.Replace(input)

	// Remove dangerous shell characters
	replacer2 := strings.NewReplacer(
		";", "",
		"&", "",
		"|", "",
		"`", "",
		"$", "",
		"<", "",
		">", "",
		"(", "",
		")", "",
		"{", "",
		"}", "",
		"[", "",
		"]", "",
		"\\", "",
		"!", "",
		"*", "",
		"?", "",
		"~", "",
		"#", "",
		"%", "",
	)
	return replacer2.Replace(input)
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Require explicit origin matching for WebSocket connections
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // same-origin request
				}
				// In production, this should match allowed origins from config
				// For now, reject obvious bad patterns
				if strings.Contains(origin, "javascript:") || strings.Contains(origin, "data:") {
					return false
				}
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (sm *SessionManager) CreateSession(userID, hostID uuid.UUID, protocol string, ws *websocket.Conn) (*Session, error) {
	session := &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		HostID:       hostID,
		Protocol:     protocol,
		WebSocket:    ws,
		active:       true,
		startedAt:    time.Now().UTC(),
		lastActivity: time.Now().UTC(),
	}

	sm.mu.Lock()
	sm.sessions[session.ID] = session
	sm.mu.Unlock()

	return session, nil
}

func (sm *SessionManager) GetSession(id string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (sm *SessionManager) RemoveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[id]; exists {
		session.Close()
		delete(sm.sessions, id)
	}
}

func (sm *SessionManager) ListSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (s *Session) ConnectSSH(host, user, password string, port int) error {
	return s.connectWithConfig(host, user, []ssh.AuthMethod{ssh.Password(password)}, port)
}

func (s *Session) ConnectSSHWithAuth(host, user string, authMethods []ssh.AuthMethod, port int) error {
	return s.connectWithConfig(host, user, authMethods, port)
}

func (s *Session) connectWithConfig(host, user string, authMethods []ssh.AuthMethod, port int) error {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: sshutil.HostKeyCallback(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	s.SSHClient = client

	sshSession, err := client.NewSession()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create SSH session: %w", err)
	}

	stdin, err := sshSession.StdinPipe()
	if err != nil {
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		stdin.Close()
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := sshSession.StderrPipe()
	if err != nil {
		stdin.Close()
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := sshSession.RequestPty("xterm-256color", 80, 24, modes); err != nil {
		stdin.Close()
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	if err := sshSession.Shell(); err != nil {
		stdin.Close()
		sshSession.Close()
		client.Close()
		return fmt.Errorf("failed to start shell: %w", err)
	}

	s.SSHSession = sshSession
	s.stdin = stdin
	s.stdout = stdout.(io.ReadCloser)
	s.stderr = stderr.(io.ReadCloser)

	go s.readLoop()
	go s.writeLoop()

	return nil
}

func (s *Session) readLoop() {
	buf := make([]byte, 4096)
	for {
		n, err := s.stdout.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			s.mu.Lock()
			s.bytesOut += int64(n)
			s.lastActivity = time.Now().UTC()
			s.mu.Unlock()

			msg := Message{Type: "data", Data: json.RawMessage(fmt.Sprintf(`"%s"`, string(buf[:n])))}
			data, _ := json.Marshal(msg)
			s.WebSocket.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (s *Session) writeLoop() {
	for {
		_, message, err := s.WebSocket.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "data":
			var data string
			if err := json.Unmarshal(msg.Data, &data); err == nil {
				// Reject only oversized input and null bytes. Terminal keystrokes legitimately contain
				// shell metacharacters as the user types them into the remote shell.
				if len(data) > maxInputLength {
					errMsg := Message{Type: "error", Data: json.RawMessage(`"input too long"`)}
					errData, _ := json.Marshal(errMsg)
					s.WebSocket.WriteMessage(websocket.TextMessage, errData)
					continue
				}
				if strings.ContainsRune(data, '\x00') {
					continue
				}
				s.mu.Lock()
				s.stdin.Write([]byte(data))
				s.bytesIn += int64(len(data))
				s.lastActivity = time.Now().UTC()
				s.mu.Unlock()
			}
		case "resize":
			var resize ResizeData
			if err := json.Unmarshal(msg.Data, &resize); err == nil {
				s.Resize(resize.Cols, resize.Rows)
			}
		case "heartbeat":
			s.mu.Lock()
			s.lastActivity = time.Now().UTC()
			s.mu.Unlock()
		}
	}
}

func (s *Session) Resize(cols, rows int) error {
	if s.SSHSession == nil {
		return errors.New("no active SSH session")
	}
	return s.SSHSession.WindowChange(rows, cols)
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}
	s.active = false

	if s.stdin != nil {
		s.stdin.Close()
	}
	if s.SSHSession != nil {
		s.SSHSession.Close()
	}
	if s.SSHClient != nil {
		s.SSHClient.Close()
	}
	if s.WebSocket != nil {
		s.WebSocket.Close()
	}
}

func (s *Session) GetStats() models.TerminalSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	return models.TerminalSession{
		ID:        uuid.MustParse(s.ID),
		UserID:    s.UserID,
		HostID:    s.HostID,
		Protocol:  s.Protocol,
		StartedAt: s.startedAt,
		BytesIn:   s.bytesIn,
		BytesOut:  s.bytesOut,
		Status:    map[bool]string{true: "active", false: "closed"}[s.active],
	}
}
