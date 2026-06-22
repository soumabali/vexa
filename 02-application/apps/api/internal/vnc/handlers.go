package vnc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var vncUpgrader = websocket.Upgrader{
	ReadBufferSize:  32 * 1024,
	WriteBufferSize: 32 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler handles VNC HTTP/WebSocket requests
type Handler struct {
	proxy *Proxy
}

// NewHandler creates a new VNC handler
func NewHandler(proxy *Proxy) *Handler {
	return &Handler{proxy: proxy}
}

// RegisterRoutes registers VNC routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/vnc/connect", h.HandleWebSocket)
	mux.HandleFunc("/api/vnc/sessions", h.HandleSessions)
	mux.HandleFunc("/api/vnc/sessions/", h.HandleSessionDetail)
}

// HandleWebSocket handles WebSocket upgrade for VNC
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := vncUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade websocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()
	
	// Read initial params
	_, data, err := conn.ReadMessage()
	if err != nil {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "failed to read params"))
		return
	}
	
	var params struct {
		HostID   string `json:"host_id"`
		UserID   string `json:"user_id"`
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}
	
	if err := json.Unmarshal(data, &params); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid params"}`))
		return
	}
	
	// Create session
	session, err := h.proxy.CreateSession(
		params.HostID,
		params.UserID,
		params.Hostname,
		params.Port,
		params.Password,
	)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}
	
	// Attach WebSocket
	session.AttachWebSocket(conn)
	
	// Send session info
	info := map[string]interface{}{
		"type":       "session_info",
		"session_id": session.ID,
		"status":     "connecting",
	}
	infoJSON, _ := json.Marshal(info)
	conn.WriteMessage(websocket.TextMessage, infoJSON)
	
	// Connect
	if err := session.Connect(); err != nil {
		errMsg := map[string]string{"error": err.Error()}
		errJSON, _ := json.Marshal(errMsg)
		conn.WriteMessage(websocket.TextMessage, errJSON)
		return
	}
	
	// Keep alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HandleSessions handles session management
func (h *Handler) HandleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSessions(w, r)
	case http.MethodPost:
		h.createSession(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleSessionDetail handles individual session
func (h *Handler) HandleSessionDetail(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Path[len("/api/vnc/sessions/"):]
	
	switch r.Method {
	case http.MethodGet:
		h.getSession(w, r, sessionID)
	case http.MethodDelete:
		h.closeSession(w, r, sessionID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.proxy.ListSessions()
	
	result := make([]map[string]interface{}, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, s.Stats())
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		HostID   string `json:"host_id"`
		UserID   string `json:"user_id"`
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	session, err := h.proxy.CreateSession(
		req.HostID,
		req.UserID,
		req.Hostname,
		req.Port,
		req.Password,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session.Stats())
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, ok := h.proxy.GetSession(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.Stats())
}

func (h *Handler) closeSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	if err := h.proxy.CloseSession(sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
