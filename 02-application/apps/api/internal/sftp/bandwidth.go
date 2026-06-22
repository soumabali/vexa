package sftp

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// BandwidthMonitor provides real-time bandwidth monitoring for SFTP sessions
type BandwidthMonitor struct {
	sessions map[string]*SessionBandwidth
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// SessionBandwidth tracks bandwidth for a single session
type SessionBandwidth struct {
	SessionID     string
	BytesSent     int64
	BytesReceived int64
	SentSpeed     float64
	ReceiveSpeed  float64
	WindowStart   time.Time
	WindowBytesIn int64
	WindowBytesOut int64
	mu            sync.RWMutex
}

// NewBandwidthMonitor creates a new bandwidth monitor
func NewBandwidthMonitor() *BandwidthMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	bm := &BandwidthMonitor{
		sessions: make(map[string]*SessionBandwidth),
		ctx:      ctx,
		cancel:   cancel,
	}
	go bm.monitoringLoop()
	return bm
}

// RegisterSession registers a new session for monitoring
func (bm *BandwidthMonitor) RegisterSession(sessionID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.sessions[sessionID] = &SessionBandwidth{
		SessionID:   sessionID,
		WindowStart: time.Now().UTC(),
	}
}

// UnregisterSession removes a session from monitoring
func (bm *BandwidthMonitor) UnregisterSession(sessionID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.sessions, sessionID)
}

// RecordBytes records bytes transferred for a session
func (bm *BandwidthMonitor) RecordBytes(sessionID string, bytesReceived, bytesSent int64) {
	bm.mu.RLock()
	session, exists := bm.sessions[sessionID]
	bm.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	atomic.AddInt64(&session.BytesReceived, bytesReceived)
	atomic.AddInt64(&session.BytesSent, bytesSent)
	atomic.AddInt64(&session.WindowBytesIn, bytesReceived)
	atomic.AddInt64(&session.WindowBytesOut, bytesSent)
}

// GetStats returns current bandwidth stats for a session
func (bm *BandwidthMonitor) GetStats(sessionID string) *BandwidthStats {
	bm.mu.RLock()
	session, exists := bm.sessions[sessionID]
	bm.mu.RUnlock()

	if !exists {
		return &BandwidthStats{}
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	return &BandwidthStats{
		BytesSent:     atomic.LoadInt64(&session.BytesSent),
		BytesReceived: atomic.LoadInt64(&session.BytesReceived),
		SentSpeed:     session.SentSpeed,
		ReceiveSpeed:  session.ReceiveSpeed,
		LastUpdated:   time.Now().UTC(),
	}
}

// GetAllStats returns bandwidth stats for all active sessions
func (bm *BandwidthMonitor) GetAllStats() map[string]*BandwidthStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	stats := make(map[string]*BandwidthStats, len(bm.sessions))
	for id, session := range bm.sessions {
		session.mu.RLock()
		stats[id] = &BandwidthStats{
			BytesSent:     atomic.LoadInt64(&session.BytesSent),
			BytesReceived: atomic.LoadInt64(&session.BytesReceived),
			SentSpeed:     session.SentSpeed,
			ReceiveSpeed:  session.ReceiveSpeed,
			LastUpdated:   time.Now().UTC(),
		}
		session.mu.RUnlock()
	}
	return stats
}

// monitoringLoop periodically calculates bandwidth speeds
func (bm *BandwidthMonitor) monitoringLoop() {
	ticker := time.NewTicker(BandwidthWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bm.calculateSpeeds()
		case <-bm.ctx.Done():
			return
		}
	}
}

// calculateSpeeds calculates bandwidth speeds for all sessions
func (bm *BandwidthMonitor) calculateSpeeds() {
	bm.mu.RLock()
	sessions := make([]*SessionBandwidth, 0, len(bm.sessions))
	for _, s := range bm.sessions {
		sessions = append(sessions, s)
	}
	bm.mu.RUnlock()

	now := time.Now().UTC()

	for _, session := range sessions {
		session.mu.Lock()

		elapsed := now.Sub(session.WindowStart).Seconds()
		if elapsed > 0 {
			session.ReceiveSpeed = float64(session.WindowBytesIn) / elapsed
			session.SentSpeed = float64(session.WindowBytesOut) / elapsed
		}

		// Reset window
		session.WindowStart = now
		session.WindowBytesIn = 0
		session.WindowBytesOut = 0

		session.mu.Unlock()
	}
}

// Stop stops the bandwidth monitor
func (bm *BandwidthMonitor) Stop() {
	bm.cancel()
}

// Global bandwidth monitor instance
var globalMonitor *BandwidthMonitor
var globalMonitorOnce sync.Once

// GetGlobalBandwidthMonitor returns the singleton bandwidth monitor
func GetGlobalBandwidthMonitor() *BandwidthMonitor {
	globalMonitorOnce.Do(func() {
		globalMonitor = NewBandwidthMonitor()
	})
	return globalMonitor
}
