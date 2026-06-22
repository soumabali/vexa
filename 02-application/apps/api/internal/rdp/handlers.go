package rdp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  32 * 1024,
	WriteBufferSize: 32 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin check
		return true
	},
}

// Handler handles RDP HTTP/WebSocket requests
type Handler struct {
	gateway *Gateway
}

// NewHandler creates a new RDP handler
func NewHandler(gateway *Gateway) *Handler {
	return &Handler{gateway: gateway}
}

// RegisterRoutes registers RDP routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/rdp/connect", h.HandleWebSocket)
	mux.HandleFunc("/api/rdp/sessions", h.HandleSessions)
	mux.HandleFunc("/api/rdp/sessions/", h.HandleSessionDetail)
}

// HandleWebSocket handles WebSocket upgrade for RDP
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade websocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()
	
	// Read initial connection params
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
		Username string `json:"username"`
		Password string `json:"password"`
		Domain   string `json:"domain"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
	}
	
	if err := json.Unmarshal(data, &params); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid params"}`))
		return
	}
	
	// Create session
	session, err := h.gateway.CreateSession(
		params.HostID,
		params.UserID,
		params.Hostname,
		params.Port,
		params.Username,
		params.Password,
		params.Domain,
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
	ctx := r.Context()
	if err := session.Connect(ctx); err != nil {
		errMsg := map[string]string{"error": err.Error()}
		errJSON, _ := json.Marshal(errMsg)
		conn.WriteMessage(websocket.TextMessage, errJSON)
		return
	}
	
	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
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

// HandleSessionDetail handles individual session operations
func (h *Handler) HandleSessionDetail(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Path[len("/api/rdp/sessions/"):]
	
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
	sessions := h.gateway.ListSessions()
	
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
		Username string `json:"username"`
		Password string `json:"password"`
		Domain   string `json:"domain"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	session, err := h.gateway.CreateSession(
		req.HostID,
		req.UserID,
		req.Hostname,
		req.Port,
		req.Username,
		req.Password,
		req.Domain,
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
	session, ok := h.gateway.GetSession(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.Stats())
}

func (h *Handler) closeSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	if err := h.gateway.CloseSession(sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
