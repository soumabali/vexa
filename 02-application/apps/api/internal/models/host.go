package models

import (
	"time"

	"github.com/google/uuid"
)

type Protocol string

const (
	ProtocolSSH Protocol = "ssh"
	ProtocolRDP Protocol = "rdp"
	ProtocolVNC Protocol = "vnc"
)

type Host struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OwnerID          uuid.UUID  `json:"owner_id" db:"owner_id"`
	Name             string     `json:"name" db:"name"`
	Address          string     `json:"address" db:"address"`
	Protocol         Protocol   `json:"protocol" db:"protocol"`
	Port             int        `json:"port" db:"port"`
	CredentialsID    *uuid.UUID `json:"credentials_id,omitempty" db:"credentials_id"`
	Tags             []string   `json:"tags" db:"tags"`
	GroupPath        string     `json:"group_path" db:"group_path"`
	AllowedUsers     []string   `json:"allowed_users" db:"allowed_users"`
	Description      string     `json:"description,omitempty" db:"description"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	WireGuardEnabled bool       `json:"wireguard_enabled" db:"wireguard_enabled"`
	WireGuardPort    int        `json:"wireguard_port" db:"wireguard_port"`
	Status           string     `json:"status" db:"status"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type HostGroup struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OwnerID   uuid.UUID  `json:"owner_id" db:"owner_id"`
	Name      string     `json:"name" db:"name"`
	Path      string     `json:"path" db:"path"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type CreateHostRequest struct {
	Name          string   `json:"name" binding:"required,min=1,max=255"`
	Address       string   `json:"address" binding:"required,hostname|ip,max=255"`
	Protocol      Protocol `json:"protocol" binding:"required,oneof=ssh rdp vnc"`
	Port          int      `json:"port" binding:"required,min=1,max=65535"`
	CredentialsID *string  `json:"credentials_id,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	GroupPath     string   `json:"group_path,omitempty,max=512"`
	Description   string   `json:"description,omitempty,max=1024"`
}

type UpdateHostRequest struct {
	Name          string   `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Address       string   `json:"address,omitempty" binding:"omitempty,hostname|ip,max=255"`
	Protocol      Protocol `json:"protocol,omitempty" binding:"omitempty,oneof=ssh rdp vnc"`
	Port          int      `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	CredentialsID *string  `json:"credentials_id,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	GroupPath     string   `json:"group_path,omitempty,max=512"`
	Description   string   `json:"description,omitempty,max=1024"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

type ConnectionHistory struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	HostID    uuid.UUID  `json:"host_id" db:"host_id"`
	Protocol  Protocol   `json:"protocol" db:"protocol"`
	StartedAt time.Time  `json:"started_at" db:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	BytesIn   int64      `json:"bytes_in" db:"bytes_in"`
	BytesOut  int64      `json:"bytes_out" db:"bytes_out"`
	Status    string     `json:"status" db:"status"`
}

type HealthCheckResult struct {
	HostID    uuid.UUID `json:"host_id"`
	Reachable bool      `json:"reachable"`
	LatencyMs int64     `json:"latency_ms"`
	Error     string    `json:"error,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}
