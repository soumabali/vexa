package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/soumabali/vexa/internal/wireguard"
)

type TunnelEventType string

const (
	TunnelEventCreate    TunnelEventType = "tunnel:create"
	TunnelEventCreated   TunnelEventType = "tunnel:created"
	TunnelEventRotate    TunnelEventType = "tunnel:rotate"
	TunnelEventRotated   TunnelEventType = "tunnel:rotated"
	TunnelEventDestroy   TunnelEventType = "tunnel:destroy"
	TunnelEventDestroyed TunnelEventType = "tunnel:destroyed"
	TunnelEventStats     TunnelEventType = "tunnel:stats"
)

type TunnelEvent struct {
	Type     TunnelEventType `json:"type"`
	TunnelID string          `json:"tunnel_id"`
	Payload  interface{}     `json:"payload,omitempty"`
	Time     int64           `json:"time"`
}

type TunnelStatsPayload struct {
	TunnelID      string `json:"tunnel_id"`
	BytesSent     int64  `json:"bytes_sent"`
	BytesRcvd     int64  `json:"bytes_rcvd"`
	LastHandshake string `json:"last_handshake,omitempty"`
}

type TunnelWSEvent struct {
	Type       TunnelEventType `json:"type"`
	TunnelID   string          `json:"tunnel_id"`
	PublicKey  string          `json:"public_key,omitempty"`
	AssignedIP string          `json:"assigned_ip,omitempty"`
	Config     string          `json:"config,omitempty"`
	Timestamp  int64           `json:"timestamp"`
}

type TunnelWSHandler struct {
	svc         *wireguard.WireGuardService
	upgrader    websocket.Upgrader
	connections map[string]map[*websocket.Conn]bool
	mu          sync.RWMutex
}

func NewTunnelWSHandler(svc *wireguard.WireGuardService) *TunnelWSHandler {
	return &TunnelWSHandler{
		svc: svc,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connections: make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *TunnelWSHandler) HandleTunnelWS(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed"})
		return
	}

	h.register(uid.String(), ws)
	defer h.unregister(uid.String(), ws)

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				done <- struct{}{}
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-pingTicker.C:
			if err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}

func (h *TunnelWSHandler) Broadcast(event TunnelEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[tunnel-ws] failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, conns := range h.connections {
		for conn := range conns {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[tunnel-ws] failed to send event to %s: %v", userID, err)
				conn.Close()
				delete(conns, conn)
			}
		}
	}
}

func (h *TunnelWSHandler) BroadcastToUser(userID string, event TunnelEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[tunnel-ws] failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, ok := h.connections[userID]; ok {
		for conn := range conns {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[tunnel-ws] failed to send event to %s: %v", userID, err)
				conn.Close()
				delete(conns, conn)
			}
		}
	}
}

func (h *TunnelWSHandler) register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[userID] == nil {
		h.connections[userID] = make(map[*websocket.Conn]bool)
	}
	h.connections[userID][conn] = true
}

func (h *TunnelWSHandler) unregister(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.connections[userID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.connections, userID)
		}
	}
}
