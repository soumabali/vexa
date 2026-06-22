package models

import (
	"time"

	"github.com/google/uuid"
)

type TunnelStatus string

const (
	TunnelStatusCreating     TunnelStatus = "creating"
	TunnelStatusConnected    TunnelStatus = "connected"
	TunnelStatusDisconnected TunnelStatus = "disconnected"
	TunnelStatusError        TunnelStatus = "error"
	TunnelStatusDisabled     TunnelStatus = "disabled"
)

type WireGuardTunnel struct {
	ID                  uuid.UUID    `json:"id" db:"id"`
	UserID              uuid.UUID    `json:"user_id" db:"user_id"`
	HostID              uuid.UUID    `json:"host_id" db:"host_id"`
	InterfaceName       string       `json:"interface_name" db:"interface_name"`
	ServerPrivateKey    string       `json:"-" db:"server_private_key"`
	ServerPublicKey     string       `json:"server_public_key" db:"server_public_key"`
	ClientPublicKey     string       `json:"client_public_key" db:"client_public_key"`
	PresharedKey        *string      `json:"preshared_key,omitempty" db:"preshared_key"`
	ServerIP            string       `json:"server_ip" db:"server_ip"`
	ClientIP            string       `json:"client_ip" db:"client_ip"`
	ListenPort          int          `json:"listen_port" db:"listen_port"`
	AllowedIPs          []string     `json:"allowed_ips" db:"allowed_ips"`
	DNSServers          []string     `json:"dns_servers,omitempty" db:"dns_servers"`
	MTU                 int          `json:"mtu" db:"mtu"`
	PersistentKeepalive int          `json:"persistent_keepalive" db:"persistent_keepalive"`
	Status              TunnelStatus `json:"status" db:"status"`
	IsEnabled           bool         `json:"is_enabled" db:"is_enabled"`
	LastHandshakeAt     *time.Time   `json:"last_handshake_at,omitempty" db:"last_handshake_at"`
	BytesSent           int64        `json:"bytes_sent" db:"bytes_sent"`
	BytesReceived       int64        `json:"bytes_received" db:"bytes_received"`
	LastRotatedAt       *time.Time   `json:"last_rotated_at,omitempty" db:"last_rotated_at"`
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at" db:"updated_at"`
}

type CreateTunnelRequest struct {
	HostID     string   `json:"host_id" binding:"required,uuid"`
	AllowedIPs []string `json:"allowed_ips"`
	Port       int      `json:"port" binding:"omitempty,min=1,max=65535"`
	DNSServers []string `json:"dns_servers,omitempty"`
	UsePSK     bool     `json:"use_psk"`
	MTU        int      `json:"mtu" binding:"omitempty,min=1280,max=1500"`
}

type UpdateTunnelRequest struct {
	AllowedIPs []string `json:"allowed_ips,omitempty"`
	Port       *int     `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	DNSServers []string `json:"dns_servers,omitempty"`
	MTU        *int     `json:"mtu,omitempty" binding:"omitempty,min=1280,max=1500"`
}

type TunnelConfigExport struct {
	ClientPrivateKey    string   `json:"client_private_key"`
	Address             string   `json:"address"`
	DNS                 []string `json:"dns"`
	ServerPublicKey     string   `json:"server_public_key"`
	PresharedKey        string   `json:"preshared_key,omitempty"`
	AllowedIPs          []string `json:"allowed_ips"`
	Endpoint            string   `json:"endpoint"`
	Port                int      `json:"port"`
	PersistentKeepalive int      `json:"persistent_keepalive"`
	MTU                 int      `json:"mtu"`
}

type TunnelStats struct {
	ID              uuid.UUID    `json:"id"`
	InterfaceName   string       `json:"interface_name"`
	Status          TunnelStatus `json:"status"`
	IsEnabled       bool         `json:"is_enabled"`
	BytesSent       int64        `json:"bytes_sent"`
	BytesReceived   int64        `json:"bytes_received"`
	LastHandshakeAt *time.Time   `json:"last_handshake_at,omitempty"`
	ClientIP        string       `json:"client_ip"`
	ServerIP        string       `json:"server_ip"`
	ListenPort      int          `json:"listen_port"`
}
