package models

import (
	"time"

	"github.com/google/uuid"
)

type TerminalSession struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	HostID      uuid.UUID  `json:"host_id"`
	Protocol    string     `json:"protocol"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	BytesIn     int64      `json:"bytes_in"`
	BytesOut    int64      `json:"bytes_out"`
	Status      string     `json:"status"`
	RecordingPath *string  `json:"recording_path,omitempty"`
	TerminatedBy *uuid.UUID `json:"terminated_by,omitempty"`
}

type AuditEvent struct {
	ID            int64          `json:"id"`
	Timestamp     time.Time      `json:"timestamp"`
	EventType     string         `json:"event_type"`
	UserID        *uuid.UUID     `json:"user_id,omitempty"`
	DeviceID      *uuid.UUID     `json:"device_id,omitempty"`
	IPAddress     string         `json:"ip_address,omitempty"`
	Details       map[string]interface{} `json:"details"`
	IntegrityHash string         `json:"integrity_hash"`
}

type AuditQuery struct {
	EventType   string    `form:"event_type"`
	UserID      string    `form:"user_id"`
	StartTime   time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime     time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Limit       int       `form:"limit,default=100"`
	Offset      int       `form:"offset,default=0"`
}
