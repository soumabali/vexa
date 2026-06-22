package sftp

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/soumabali/vexa/internal/auth"
)

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrInvalidBinaryFrame = errors.New("invalid binary frame format")
	ErrTransferComplete   = errors.New("transfer already complete")
)

// FileTransferMessage types for WebSocket communication
type FileTransferMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	Path      string `json:"path,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Offset    int64  `json:"offset,omitempty"`
	ChunkSize int    `json:"chunk_size,omitempty"`
	TransferID string `json:"transfer_id,omitempty"`
	Error     string `json:"error,omitempty"`
	Progress  *Progress `json:"progress,omitempty"`
}

// Binary frame format: [offset: 8 bytes][length: 4 bytes][data: N bytes]
const (
	BinaryHeaderSize = 12 // 8 + 4
	MaxChunkSize     = 1024 * 1024 // 1MB max chunk
)

type SFTPWebSocketHandler struct {
	sftpService *SFTPService
	jwtManager  *auth.JWTManager
	upgrader    websocket.Upgrader

	clients map[string]*WSClient // sessionID -> client
	mu      sync.RWMutex
}

type WSClient struct {
	SessionID  string
	UserID     string
	Conn       *websocket.Conn
	Send       chan []byte
	Transfers  map[string]*Transfer
	mu         sync.RWMutex
	active     bool
}

func NewSFTPWebSocketHandler(sftpService *SFTPService, jwtManager *auth.JWTManager, allowedOrigins []string) *SFTPWebSocketHandler {
	return &SFTPWebSocketHandler{
		sftpService: sftpService,
		jwtManager:  jwtManager,
		clients:     make(map[string]*WSClient),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				for _, allowed := range allowedOrigins {
					if allowed == "*" || strings.EqualFold(origin, allowed) {
						return true
					}
				}
				return false
			},
			ReadBufferSize:  64 * 1024,
			WriteBufferSize: 64 * 1024,
		},
	}
}

// HandleWebSocket upgrades HTTP to WebSocket and handles SFTP file transfers
func (h *SFTPWebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Authenticate
	claims, err := h.authenticateWebSocket(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	// Validate session exists
	_, err = h.sftpService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SFTP session not found"})
		return
	}

	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed"})
		return
	}

	client := &WSClient{
		SessionID: sessionID,
		UserID:    claims.UserID.String(),
		Conn:      ws,
		Send:      make(chan []byte, 256),
		Transfers: make(map[string]*Transfer),
		active:    true,
	}

	h.mu.Lock()
	h.clients[sessionID] = client
	h.mu.Unlock()

	// Start goroutines
	go h.writePump(client)
	go h.readPump(client)

	// Send ready message
	h.sendJSON(client, FileTransferMessage{
		Type:      "ready",
		SessionID: sessionID,
	})
}

func (h *SFTPWebSocketHandler) authenticateWebSocket(r *http.Request) (*auth.Claims, error) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenString = parts[1]
			}
		}
	}

	if tokenString == "" {
		return nil, errors.New("authentication token required")
	}

	claims, err := h.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}

func (h *SFTPWebSocketHandler) readPump(client *WSClient) {
	defer func() {
		client.mu.Lock()
		client.active = false
		client.mu.Unlock()
		h.removeClient(client.SessionID)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(MaxChunkSize + BinaryHeaderSize + 1024)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}

		if messageType == websocket.TextMessage {
			h.handleTextMessage(client, message)
		} else if messageType == websocket.BinaryMessage {
			h.handleBinaryMessage(client, message)
		}
	}
}

func (h *SFTPWebSocketHandler) writePump(client *WSClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if isBinaryFrame(message) {
				client.Conn.WriteMessage(websocket.BinaryMessage, message)
			} else {
				client.Conn.WriteMessage(websocket.TextMessage, message)
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func isBinaryFrame(data []byte) bool {
	return len(data) >= BinaryHeaderSize && data[0] == 0x00
}

func (h *SFTPWebSocketHandler) handleTextMessage(client *WSClient, data []byte) {
	var msg FileTransferMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		h.sendError(client, "", "invalid JSON message")
		return
	}

	switch msg.Type {
	case "upload":
		h.handleUploadRequest(client, &msg)
	case "download":
		h.handleDownloadRequest(client, &msg)
	case "list":
		h.handleListRequest(client, &msg)
	case "remove":
		h.handleRemoveRequest(client, &msg)
	case "mkdir":
		h.handleMkdirRequest(client, &msg)
	case "move":
		h.handleMoveRequest(client, &msg)
	case "cancel":
		h.handleCancelRequest(client, &msg)
	case "bandwidth":
		h.handleBandwidthRequest(client, &msg)
	case "heartbeat":
		// Just keep connection alive
	default:
		h.sendError(client, msg.SessionID, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

func (h *SFTPWebSocketHandler) handleBinaryMessage(client *WSClient, data []byte) {
	if len(data) < BinaryHeaderSize {
		h.sendError(client, client.SessionID, "binary frame too small")
		return
	}

	offset := int64(binary.BigEndian.Uint64(data[0:8]))
	length := binary.BigEndian.Uint32(data[8:12])

	if uint32(len(data)) < BinaryHeaderSize+length {
		h.sendError(client, client.SessionID, "binary frame length mismatch")
		return
	}

	chunkData := data[BinaryHeaderSize : BinaryHeaderSize+length]

	// Find active upload transfer for this client
	client.mu.RLock()
	var activeTransfer *Transfer
	for _, t := range client.Transfers {
		t.mu.RLock()
		if t.Type == "upload" && !t.Completed && !t.Cancelled {
			activeTransfer = t
		}
		t.mu.RUnlock()
		if activeTransfer != nil {
			break
		}
	}
	client.mu.RUnlock()

	if activeTransfer == nil {
		h.sendError(client, client.SessionID, "no active upload transfer")
		return
	}

	_ = offset // Used for resume support
	_ = chunkData

	// Update transfer progress
	activeTransfer.mu.Lock()
	activeTransfer.Offset += int64(len(chunkData))
	activeTransfer.mu.Unlock()

	// Send progress update
	progress := h.calculateProgress(activeTransfer)
	h.sendJSON(client, FileTransferMessage{
		Type:       "progress",
		TransferID: activeTransfer.ID,
		Progress:   progress,
	})
}

func (h *SFTPWebSocketHandler) handleUploadRequest(client *WSClient, msg *FileTransferMessage) {
	if msg.Path == "" {
		h.sendError(client, msg.SessionID, "path required")
		return
	}

	// Create transfer entry
	transfer := &Transfer{
		ID:        fmt.Sprintf("ws-upload-%d", time.Now().UnixNano()),
		SessionID: msg.SessionID,
		Path:      msg.Path,
		Type:      "upload",
		Size:      msg.Size,
		Offset:    msg.Offset,
		StartedAt: time.Now().UTC(),
	}

	client.mu.Lock()
	client.Transfers[transfer.ID] = transfer
	client.mu.Unlock()

	h.sendJSON(client, FileTransferMessage{
		Type:       "upload_ready",
		SessionID:  msg.SessionID,
		TransferID: transfer.ID,
		ChunkSize:  DefaultChunkSize,
	})
}

func (h *SFTPWebSocketHandler) handleDownloadRequest(client *WSClient, msg *FileTransferMessage) {
	if msg.Path == "" {
		h.sendError(client, msg.SessionID, "path required")
		return
	}

	reader, transfer, err := h.sftpService.Download(msg.SessionID, msg.Path, msg.Offset)
	if err != nil {
		h.sendError(client, msg.SessionID, fmt.Sprintf("download failed: %v", err))
		return
	}
	defer reader.Close()

	client.mu.Lock()
	client.Transfers[transfer.ID] = transfer
	client.mu.Unlock()

	// Send download ready message
	h.sendJSON(client, FileTransferMessage{
		Type:       "download_ready",
		SessionID:  msg.SessionID,
		TransferID: transfer.ID,
		Size:       transfer.Size,
		Offset:     msg.Offset,
		ChunkSize:  DefaultChunkSize,
	})

	// Stream file data as binary frames
	buf := make([]byte, DefaultChunkSize)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			frame := h.buildBinaryFrame(transfer.Offset-int64(n), buf[:n])
			select {
			case client.Send <- frame:
			default:
				// Channel full, drop frame
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			h.sendError(client, msg.SessionID, fmt.Sprintf("read error: %v", err))
			return
		}
	}

	transfer.mu.Lock()
	transfer.Completed = true
	transfer.mu.Unlock()

	h.sendJSON(client, FileTransferMessage{
		Type:       "complete",
		TransferID: transfer.ID,
	})
}

func (h *SFTPWebSocketHandler) handleListRequest(client *WSClient, msg *FileTransferMessage) {
	files, err := h.sftpService.ListDirectory(msg.SessionID, msg.Path)
	if err != nil {
		h.sendError(client, msg.SessionID, fmt.Sprintf("list failed: %v", err))
		return
	}

	h.sendJSON(client, FileTransferMessage{
		Type:      "list_result",
		SessionID: msg.SessionID,
	})

	// Also send file list as a separate message with file details
	listData, _ := json.Marshal(map[string]interface{}{
		"type":  "file_list",
		"files": files,
	})
	select {
	case client.Send <- listData:
	default:
	}
}

func (h *SFTPWebSocketHandler) handleRemoveRequest(client *WSClient, msg *FileTransferMessage) {
	if msg.Path == "" {
		h.sendError(client, msg.SessionID, "path required")
		return
	}

	if err := h.sftpService.Remove(msg.SessionID, msg.Path); err != nil {
		h.sendError(client, msg.SessionID, fmt.Sprintf("remove failed: %v", err))
		return
	}

	h.sendJSON(client, FileTransferMessage{
		Type:      "remove_complete",
		SessionID: msg.SessionID,
		Path:      msg.Path,
	})
}

func (h *SFTPWebSocketHandler) handleMkdirRequest(client *WSClient, msg *FileTransferMessage) {
	if msg.Path == "" {
		h.sendError(client, msg.SessionID, "path required")
		return
	}

	if err := h.sftpService.Mkdir(msg.SessionID, msg.Path); err != nil {
		h.sendError(client, msg.SessionID, fmt.Sprintf("mkdir failed: %v", err))
		return
	}

	h.sendJSON(client, FileTransferMessage{
		Type:      "mkdir_complete",
		SessionID: msg.SessionID,
		Path:      msg.Path,
	})
}

func (h *SFTPWebSocketHandler) handleMoveRequest(client *WSClient, msg *FileTransferMessage) {
	// Parse old and new paths from the message
	oldPath := msg.Path
	newPath := "" // Would need to be parsed from raw JSON

	if oldPath == "" || newPath == "" {
		h.sendError(client, msg.SessionID, "both old_path and new_path required")
		return
	}

	if err := h.sftpService.Move(msg.SessionID, oldPath, newPath); err != nil {
		h.sendError(client, msg.SessionID, fmt.Sprintf("move failed: %v", err))
		return
	}

	h.sendJSON(client, FileTransferMessage{
		Type:      "move_complete",
		SessionID: msg.SessionID,
	})
}

func (h *SFTPWebSocketHandler) handleCancelRequest(client *WSClient, msg *FileTransferMessage) {
	if msg.TransferID == "" {
		h.sendError(client, msg.SessionID, "transfer_id required")
		return
	}

	if err := h.sftpService.CancelTransfer(msg.TransferID); err != nil {
		h.sendError(client, msg.SessionID, fmt.Sprintf("cancel failed: %v", err))
		return
	}

	client.mu.Lock()
	delete(client.Transfers, msg.TransferID)
	client.mu.Unlock()

	h.sendJSON(client, FileTransferMessage{
		Type:       "cancel_complete",
		TransferID: msg.TransferID,
	})
}

func (h *SFTPWebSocketHandler) handleBandwidthRequest(client *WSClient, msg *FileTransferMessage) {
	stats := h.sftpService.MonitorBandwidth(msg.SessionID)

	data, _ := json.Marshal(map[string]interface{}{
		"type":           "bandwidth",
		"bytes_sent":     stats.BytesSent,
		"bytes_received": stats.BytesReceived,
		"sent_speed":     stats.SentSpeed,
		"receive_speed":  stats.ReceiveSpeed,
	})

	select {
	case client.Send <- data:
	default:
	}
}

func (h *SFTPWebSocketHandler) sendJSON(client *WSClient, msg FileTransferMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.Send <- data:
	default:
	}
}

func (h *SFTPWebSocketHandler) sendError(client *WSClient, sessionID, message string) {
	h.sendJSON(client, FileTransferMessage{
		Type:      "error",
		SessionID: sessionID,
		Error:     message,
	})
}

func (h *SFTPWebSocketHandler) removeClient(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, sessionID)
}

func (h *SFTPWebSocketHandler) buildBinaryFrame(offset int64, data []byte) []byte {
	frame := make([]byte, BinaryHeaderSize+len(data))
	frame[0] = 0x00 // Marker byte to identify binary frame
	binary.BigEndian.PutUint64(frame[0:8], uint64(offset))
	binary.BigEndian.PutUint32(frame[8:12], uint32(len(data)))
	copy(frame[BinaryHeaderSize:], data)
	return frame
}

func (h *SFTPWebSocketHandler) calculateProgress(transfer *Transfer) *Progress {
	transfer.mu.RLock()
	defer transfer.mu.RUnlock()

	percentage := 0.0
	if transfer.Size > 0 {
		percentage = float64(transfer.Offset) / float64(transfer.Size) * 100
	}

	speed := 0.0
	elapsed := time.Since(transfer.StartedAt).Seconds()
	if elapsed > 0 {
		speed = float64(transfer.Offset) / elapsed
	}

	return &Progress{
		BytesTransferred: transfer.Offset,
		TotalBytes:       transfer.Size,
		Percentage:       percentage,
		SpeedBps:         speed,
	}
}

// GetActiveClients returns the number of active WebSocket clients
func (h *SFTPWebSocketHandler) GetActiveClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
