package wireguard

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type RotationScheduler struct {
	repo      *Repository
	ifaceCtrl *InterfaceController
	maxAge    time.Duration
	interval  time.Duration
	stopCh    chan struct{}
	running   bool
	mu        sync.Mutex
}

func NewRotationScheduler(repo *Repository, ifaceCtrl *InterfaceController, maxAgeDays int) *RotationScheduler {
	return &RotationScheduler{
		repo:      repo,
		ifaceCtrl: ifaceCtrl,
		maxAge:    time.Duration(maxAgeDays) * 24 * time.Hour,
		interval:  24 * time.Hour,
		stopCh:    make(chan struct{}),
	}
}

func (rs *RotationScheduler) Start() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.running {
		return
	}

	rs.running = true
	rs.stopCh = make(chan struct{})
	go rs.loop()

	log.Printf("[wg-rotate] key rotation scheduler started (max_age=%s, check_interval=%s)",
		rs.maxAge, rs.interval)
}

func (rs *RotationScheduler) Stop() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if !rs.running {
		return
	}

	close(rs.stopCh)
	rs.running = false
	log.Printf("[wg-rotate] key rotation scheduler stopped")
}

func (rs *RotationScheduler) SetInterval(d time.Duration) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.interval = d
}

func (rs *RotationScheduler) IsRunning() bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.running
}

func (rs *RotationScheduler) loop() {
	ticker := time.NewTicker(rs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rs.stopCh:
			return
		case <-ticker.C:
			rs.rotateDue()
		}
	}
}

func (rs *RotationScheduler) rotateDue() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	maxAgeDays := int(rs.maxAge.Hours() / 24)

	tunnels, err := rs.repo.GetTunnelsDueForRotation(ctx, maxAgeDays)
	if err != nil {
		log.Printf("[wg-rotate] failed to query tunnels due for rotation: %v", err)
		return
	}

	for _, tunnel := range tunnels {
		if err := rs.rotateTunnelKeys(ctx, tunnel.ID.String()); err != nil {
			log.Printf("[wg-rotate] failed to rotate keys for tunnel %s: %v", tunnel.ID, err)
		}
	}

	if len(tunnels) > 0 {
		log.Printf("[wg-rotate] rotated keys for %d tunnels", len(tunnels))
	}
}

func (rs *RotationScheduler) rotateTunnelKeys(ctx context.Context, tunnelID string) error {
	id, err := uuid.Parse(tunnelID)
	if err != nil {
		return err
	}

	newPriv, newPub, err := generateKeyPair()
	if err != nil {
		return err
	}

	if err := rs.repo.Update(ctx, id, map[string]interface{}{
		"server_private_key": newPriv,
		"server_public_key":  newPub,
		"last_rotated_at":    time.Now().UTC(),
	}); err != nil {
		return err
	}

	log.Printf("[wg-rotate] rotated keys for tunnel %s", tunnelID)
	return nil
}
