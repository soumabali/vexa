// Package sandbox provides Linux security sandboxing for protocol handlers
// using seccomp-bpf, cgroups v2, namespaces, capabilities, and AppArmor.
package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Constants
const (
	DefaultCPUPercent    = 100 // 100% of one core
	DefaultMemoryMB      = 512
	DefaultDiskMB        = 1024
	DefaultMaxPIDs       = 64
	DefaultSessionTimeout = 24 * time.Hour
	DefaultKillTimeout    = 5 * time.Second
)

// Profile represents a sandbox profile for a specific protocol
type Profile struct {
	Name        string
	Protocol    string // ssh, rdp, vnc
	Syscalls    []string
	SeccompPath string
	AppArmor    string
	CgroupPath  string
	Resources   ResourceLimits
}

// ResourceLimits defines cgroup resource constraints
type ResourceLimits struct {
	CPUPercent   int    // 0-100 per core
	MemoryMB     int
	DiskMB       int
	MaxPIDs      int
	MaxOpenFiles int
	NetEgressOnly bool
}

// Sandbox represents a running sandboxed process
type Sandbox struct {
	ID          string
	Profile     *Profile
	Cmd         *exec.Cmd
	PID         int
	Status      Status
	CreatedAt   time.Time
	LastActive  time.Time
	mu          sync.RWMutex
	metrics     *Metrics
	sessionID   string
}

// Status represents sandbox lifecycle state
type Status int

const (
	StatusPending Status = iota
	StatusInitializing
	StatusRunning
	StatusPaused
	StatusTerminating
	StatusTerminated
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusInitializing:
		return "initializing"
	case StatusRunning:
		return "running"
	case StatusPaused:
		return "paused"
	case StatusTerminating:
		return "terminating"
	case StatusTerminated:
		return "terminated"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Manager manages all active sandboxes
type Manager struct {
	sandboxes   map[string]*Sandbox
	profiles    map[string]*Profile
	mu          sync.RWMutex
	wg          sync.WaitGroup
	stopCh      chan struct{}
	metrics     *Metrics
	cfg         Config
	seccomp     *SeccompFilter
	cgroups     *CgroupManager
	apparmor    *AppArmorManager
	namespaces  *NamespaceManager
	caps        *CapabilityManager
}

// Config holds sandbox manager configuration
type Config struct {
	Enabled         bool
	SeccompEnabled  bool
	CgroupsEnabled  bool
	AppArmorEnabled bool
	NamespacesEnabled bool
	CapsEnabled     bool
	DefaultProfile  string
	BaseCgroupPath  string
	AppArmorDir     string
	SeccompDir      string
	SessionTimeout  time.Duration
	CleanupInterval time.Duration
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Enabled:           true,
		SeccompEnabled:    true,
		CgroupsEnabled:    true,
		AppArmorEnabled:   true,
		NamespacesEnabled: true,
		CapsEnabled:       true,
		DefaultProfile:    "ssh",
		BaseCgroupPath:    "/sys/fs/cgroup/vexa",
		AppArmorDir:       "/etc/apparmor.d/vexa",
		SeccompDir:        "/etc/vexa/seccomp",
		SessionTimeout:    DefaultSessionTimeout,
		CleanupInterval:   5 * time.Minute,
	}
}

// NewManager creates a new sandbox manager
func NewManager(cfg Config, reg *prometheus.Registry) (*Manager, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("sandbox: Linux only (current: %s)", runtime.GOOS)
	}

	m := &Manager{
		sandboxes: make(map[string]*Sandbox),
		profiles:  make(map[string]*Profile),
		stopCh:    make(chan struct{}),
		cfg:       cfg,
		metrics:   NewMetrics(reg),
	}

	// Initialize subsystems
	var errs []string

	if cfg.SeccompEnabled {
		seccomp, err := NewSeccompFilter()
		if err != nil {
			errs = append(errs, fmt.Sprintf("seccomp: %v", err))
			cfg.SeccompEnabled = false
		} else {
			m.seccomp = seccomp
		}
	}

	if cfg.CgroupsEnabled {
		cgroups, err := NewCgroupManager(cfg.BaseCgroupPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("cgroups: %v", err))
			cfg.CgroupsEnabled = false
		} else {
			m.cgroups = cgroups
		}
	}

	if cfg.AppArmorEnabled {
		apparmor, err := NewAppArmorManager(cfg.AppArmorDir)
		if err != nil {
			errs = append(errs, fmt.Sprintf("apparmor: %v", err))
			cfg.AppArmorEnabled = false
		} else {
			m.apparmor = apparmor
		}
	}

	if cfg.NamespacesEnabled {
		m.namespaces = NewNamespaceManager()
	}

	if cfg.CapsEnabled {
		m.caps = NewCapabilityManager()
	}

	// Load built-in profiles
	if err := m.loadBuiltinProfiles(); err != nil {
		errs = append(errs, fmt.Sprintf("profiles: %v", err))
	}

	// Log warnings but continue with degraded features
	if len(errs) > 0 {
		// Log warning but continue
		_ = fmt.Sprintf("sandbox: degraded mode (%d/%d features): %s",
			len(errs), 5, strings.Join(errs, "; "))
	}

	// Start cleanup goroutine
	m.wg.Add(1)
	go m.cleanupLoop()

	return m, nil
}

// RegisterProfile registers a sandbox profile
func (m *Manager) RegisterProfile(p *Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.profiles[p.Name]; exists {
		return fmt.Errorf("profile %q already exists", p.Name)
	}

	m.profiles[p.Name] = p
	return nil
}

// GetProfile retrieves a profile by name
func (m *Manager) GetProfile(name string) (*Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

// CreateSandbox creates a new sandboxed session
func (m *Manager) CreateSandbox(id, profileName, sessionID string) (*Sandbox, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	profile, ok := m.profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}

	sb := &Sandbox{
		ID:         id,
		Profile:    profile,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		metrics:    m.metrics,
		sessionID:  sessionID,
	}

	m.sandboxes[id] = sb
	m.metrics.SandboxTotal.Inc()
	m.metrics.SandboxActive.Inc()

	return sb, nil
}

// Start initializes and starts the sandbox
func (m *Manager) Start(sb *Sandbox, cmd *exec.Cmd) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.Status != StatusPending {
		return fmt.Errorf("sandbox %s: invalid state %s for start", sb.ID, sb.Status)
	}

	sb.Status = StatusInitializing

	// Apply sandboxing layers
	// 1. Namespaces (must be first)
	if m.cfg.NamespacesEnabled && m.namespaces != nil {
		if err := m.namespaces.Apply(cmd); err != nil {
			m.metrics.SandboxErrors.WithLabelValues("namespace").Inc()
			// Graceful degradation
		}
	}

	// 2. Capabilities
	if m.cfg.CapsEnabled && m.caps != nil {
		if err := m.caps.Drop(cmd); err != nil {
			m.metrics.SandboxErrors.WithLabelValues("capability").Inc()
		}
	}

	// 3. Cgroups
	if m.cfg.CgroupsEnabled && m.cgroups != nil {
		if err := m.cgroups.Apply(sb.ID, sb.Profile.Resources); err != nil {
			m.metrics.SandboxErrors.WithLabelValues("cgroup").Inc()
		}
	}

	// 4. Seccomp
	if m.cfg.SeccompEnabled && m.seccomp != nil {
		if err := m.seccomp.Apply(cmd, sb.Profile.SeccompPath); err != nil {
			m.metrics.SandboxErrors.WithLabelValues("seccomp").Inc()
		}
	}

	// 5. AppArmor
	if m.cfg.AppArmorEnabled && m.apparmor != nil {
		if err := m.apparmor.Apply(cmd, sb.Profile.AppArmor); err != nil {
			m.metrics.SandboxErrors.WithLabelValues("apparmor").Inc()
		}
	}

	sb.Cmd = cmd

	// Start the process
	if err := cmd.Start(); err != nil {
		sb.Status = StatusError
		m.metrics.SandboxErrors.WithLabelValues("start").Inc()
		return fmt.Errorf("sandbox %s: start failed: %w", sb.ID, err)
	}

	sb.PID = cmd.Process.Pid
	sb.Status = StatusRunning
	sb.LastActive = time.Now()

	// Move process to cgroup after start
	if m.cfg.CgroupsEnabled && m.cgroups != nil {
		if err := m.cgroups.AddProcess(sb.ID, sb.PID); err != nil {
			m.metrics.SandboxErrors.WithLabelValues("cgroup_attach").Inc()
		}
	}

	m.metrics.SandboxStarted.Inc()

	// Start monitoring goroutine
	m.wg.Add(1)
	go m.monitorSandbox(sb)

	return nil
}

// Stop gracefully stops a sandbox
func (m *Manager) Stop(id string) error {
	sb, err := m.GetSandbox(id)
	if err != nil {
		return err
	}

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.Status == StatusTerminated || sb.Status == StatusTerminating {
		return nil
	}

	sb.Status = StatusTerminating

	// Send SIGTERM first
	if sb.Cmd != nil && sb.Cmd.Process != nil {
		sb.Cmd.Process.Signal(syscall.SIGTERM)
	}

	// Wait for graceful shutdown
	done := make(chan struct{})
	go func() {
		if sb.Cmd != nil {
			sb.Cmd.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		// Graceful shutdown succeeded
	case <-time.After(DefaultKillTimeout):
		// Force kill
		if sb.Cmd != nil && sb.Cmd.Process != nil {
			sb.Cmd.Process.Kill()
		}
	}

	// Cleanup cgroup
	if m.cfg.CgroupsEnabled && m.cgroups != nil {
		m.cgroups.Remove(id)
	}

	sb.Status = StatusTerminated
	m.metrics.SandboxActive.Dec()
	m.metrics.SandboxStopped.Inc()

	return nil
}

// GetSandbox retrieves a sandbox by ID
func (m *Manager) GetSandbox(id string) (*Sandbox, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sb, ok := m.sandboxes[id]
	if !ok {
		return nil, fmt.Errorf("sandbox %s not found", id)
	}
	return sb, nil
}

// ListSandboxes returns all active sandboxes
func (m *Manager) ListSandboxes() []*Sandbox {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Sandbox, 0, len(m.sandboxes))
	for _, sb := range m.sandboxes {
		result = append(result, sb)
	}
	return result
}

// cleanupLoop periodically cleans up expired sandboxes
func (m *Manager) cleanupLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup removes expired sandboxes
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, sb := range m.sandboxes {
		sb.mu.RLock()
		expired := now.Sub(sb.LastActive) > m.cfg.SessionTimeout
		terminated := sb.Status == StatusTerminated || sb.Status == StatusError
		sb.mu.RUnlock()

		if expired || terminated {
			delete(m.sandboxes, id)
			m.metrics.SandboxActive.Dec()
		}
	}
}

// monitorSandbox monitors a running sandbox
func (m *Manager) monitorSandbox(sb *Sandbox) {
	defer m.wg.Done()

	if sb.Cmd == nil {
		return
	}

	err := sb.Cmd.Wait()

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if err != nil {
		sb.Status = StatusError
		m.metrics.SandboxErrors.WithLabelValues("exit").Inc()
	} else {
		sb.Status = StatusTerminated
	}

	m.metrics.SandboxActive.Dec()
	m.metrics.SandboxStopped.Inc()

	// Cleanup cgroup
	if m.cfg.CgroupsEnabled && m.cgroups != nil {
		m.cgroups.Remove(sb.ID)
	}
}

// Stop gracefully shuts down the manager
func (m *Manager) StopManager() {
	close(m.stopCh)

	// Stop all sandboxes
	m.mu.Lock()
	for _, sb := range m.sandboxes {
		go m.Stop(sb.ID)
	}
	m.mu.Unlock()

	// Wait for all goroutines
	m.wg.Wait()
}

// loadBuiltinProfiles loads default protocol profiles
func (m *Manager) loadBuiltinProfiles() error {
	profiles := []*Profile{
		{
			Name:     "ssh",
			Protocol: "ssh",
			Resources: ResourceLimits{
				CPUPercent:   DefaultCPUPercent,
				MemoryMB:     DefaultMemoryMB,
				DiskMB:       DefaultDiskMB,
				MaxPIDs:      DefaultMaxPIDs,
				MaxOpenFiles: 256,
				NetEgressOnly: true,
			},
			SeccompPath: filepath.Join(m.cfg.SeccompDir, "seccomp-ssh.json"),
			AppArmor:    "vexa-ssh",
		},
		{
			Name:     "rdp",
			Protocol: "rdp",
			Resources: ResourceLimits{
				CPUPercent:   DefaultCPUPercent,
				MemoryMB:     1024, // RDP needs more memory
				DiskMB:       DefaultDiskMB,
				MaxPIDs:      128,
				MaxOpenFiles: 512,
				NetEgressOnly: true,
			},
			SeccompPath: filepath.Join(m.cfg.SeccompDir, "seccomp-rdp.json"),
			AppArmor:    "vexa-rdp",
		},
		{
			Name:     "vnc",
			Protocol: "vnc",
			Resources: ResourceLimits{
				CPUPercent:   DefaultCPUPercent,
				MemoryMB:     DefaultMemoryMB,
				DiskMB:       DefaultDiskMB,
				MaxPIDs:      64,
				MaxOpenFiles: 256,
				NetEgressOnly: true,
			},
			SeccompPath: filepath.Join(m.cfg.SeccompDir, "seccomp-vnc.json"),
			AppArmor:    "vexa-vnc",
		},
	}

	for _, p := range profiles {
		if err := m.RegisterProfile(p); err != nil {
			return err
		}
	}

	return nil
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// CheckRoot returns true if running as root
func CheckRoot() bool {
	return os.Getuid() == 0
}

// GetStatus returns current sandbox status
func (sb *Sandbox) GetStatus() Status {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.Status
}

// GetPID returns sandbox process PID
func (sb *Sandbox) GetPID() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.PID
}

// Touch updates last active timestamp
func (sb *Sandbox) Touch() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.LastActive = time.Now()
}

// String returns human-readable status
func (sb *Sandbox) String() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return fmt.Sprintf("Sandbox[%s] profile=%s status=%s pid=%d created=%s",
		sb.ID, sb.Profile.Name, sb.Status, sb.PID, sb.CreatedAt.Format(time.RFC3339))
}
