package security

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	EventType     string                 `json:"event_type"`
	Severity      string                 `json:"severity"`
	UserID        string                 `json:"user_id,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	IP            string                 `json:"ip"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Method        string                 `json:"method,omitempty"`
	Path          string                 `json:"path,omitempty"`
	StatusCode    int                    `json:"status_code,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	backend AuditBackend
}

// AuditBackend defines storage interface
type AuditBackend interface {
	Write(ctx context.Context, event *AuditEvent) error
	Query(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error)
}

// AuditFilter for querying logs
type AuditFilter struct {
	UserID      string
	EventType   string
	Severity    string
	StartTime   time.Time
	EndTime     time.Time
	Limit       int
	Offset      int
}

// NewAuditLogger creates new audit logger
func NewAuditLogger(backend AuditBackend) *AuditLogger {
	return &AuditLogger{backend: backend}
}

// LogEvent logs a single audit event
func (a *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
	event.ID = uuid.New().String()
	event.Timestamp = time.Now().UTC()

	// Set default severity
	if event.Severity == "" {
		event.Severity = "INFO"
	}

	return a.backend.Write(ctx, event)
}

// LogAuth logs authentication events
func (a *AuditLogger) LogAuth(ctx context.Context, success bool, userID, ip, method string, err error) {
	severity := "INFO"
	if !success {
		severity = "WARNING"
	}

	event := &AuditEvent{
		EventType: "auth",
		Severity:  severity,
		UserID:    userID,
		IP:        ip,
		Method:    method,
		Details: map[string]interface{}{
			"success": success,
		},
	}

	if err != nil {
		event.Error = err.Error()
	}

	a.LogEvent(ctx, event)
}

// LogAccess logs API access
func (a *AuditLogger) LogAccess(ctx context.Context, r *http.Request, statusCode int, duration time.Duration) {
	severity := "INFO"
	if statusCode >= 400 {
		severity = "WARNING"
	}
	if statusCode >= 500 {
		severity = "ERROR"
	}

	event := &AuditEvent{
		EventType:  "access",
		Severity:   severity,
		IP:         r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: statusCode,
		Duration:   duration,
	}

	a.LogEvent(ctx, event)
}

// LogSession logs session events
func (a *AuditLogger) LogSession(ctx context.Context, action, sessionID, userID, ip string) {
	event := &AuditEvent{
		EventType: "session",
		Severity:  "INFO",
		UserID:    userID,
		SessionID: sessionID,
		IP:        ip,
		Details: map[string]interface{}{
			"action": action,
		},
	}

	a.LogEvent(ctx, event)
}

// LogAdmin logs admin actions
func (a *AuditLogger) LogAdmin(ctx context.Context, action, adminID, targetType, targetID string, details map[string]interface{}) {
	event := &AuditEvent{
		EventType: "admin",
		Severity:  "INFO",
		UserID:    adminID,
		Details: map[string]interface{}{
			"action":      action,
			"target_type": targetType,
			"target_id":   targetID,
			"details":     details,
		},
	}

	a.LogEvent(ctx, event)
}

// LogSecurity logs security events (WAF, rate limit, etc.)
func (a *AuditLogger) LogSecurity(ctx context.Context, eventType, severity, ip string, details map[string]interface{}) {
	event := &AuditEvent{
		EventType: eventType,
		Severity:  severity,
		IP:        ip,
		Details:   details,
	}

	a.LogEvent(ctx, event)
}

// Query queries audit logs
func (a *AuditLogger) Query(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error) {
	return a.backend.Query(ctx, filter)
}

// AuditMiddleware returns middleware that logs all requests
func AuditMiddleware(logger *AuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.LogAccess(r.Context(), r, rw.statusCode, duration)
		})
	}
}

// responseRecorder captures status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// ConsoleBackend logs to stdout (for development)
type ConsoleBackend struct{}

func (c *ConsoleBackend) Write(ctx context.Context, event *AuditEvent) error {
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
	return nil
}

func (c *ConsoleBackend) Query(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error) {
	return nil, nil
}
