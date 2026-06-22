package wireguard

import (
	"errors"
	"time"
)

var (
	ErrTunnelNotFound    = errors.New("wireguard tunnel not found")
	ErrTunnelExists      = errors.New("wireguard tunnel already exists for this host")
	ErrMaxTunnelsReached = errors.New("maximum tunnels per user reached")
	ErrPortUnavailable   = errors.New("listen port unavailable")
	ErrInterfaceNotFound = errors.New("wireguard interface not found on system")
	ErrNoAvailableIP     = errors.New("no available IP in subnet pool")
	ErrInvalidSubnet     = errors.New("invalid subnet configuration")
)

type ServerConfig struct {
	Subnet            string `json:"subnet"`
	WGPortRange       [2]int `json:"wg_port_range"`
	RotationDays      int    `json:"rotation_days"`
	MaxTunnelsPerUser int    `json:"max_tunnels_per_user"`
	MaxTotalTunnels   int    `json:"max_total_tunnels"`
	Enabled           bool   `json:"enabled"`
	BinPath           string `json:"bin_path"`
}

type RotationEvent struct {
	TunnelID     string
	OldPublicKey string
	NewPublicKey string
	RotatedAt    time.Time
}
