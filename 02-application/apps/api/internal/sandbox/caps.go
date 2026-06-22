//go:build linux

package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// CapabilityManager manages Linux capabilities for sandbox processes
type CapabilityManager struct {
	enabled bool
}

// NewCapabilityManager creates a new capability manager
func NewCapabilityManager() *CapabilityManager {
	return &CapabilityManager{enabled: true}
}

// Drop drops all capabilities except essential ones
func (c *CapabilityManager) Drop(cmd *exec.Cmd) error {
	if !c.enabled {
		return nil
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Clear ambient caps
	cmd.SysProcAttr.AmbientCaps = nil
	// Set no_new_privs via prctl on current process; child inherits it.
	return unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)
}

// DropAll drops all capabilities from the current process
func (c *CapabilityManager) DropAll() error {
	if !c.enabled {
		return nil
	}
	if os.Getuid() != 0 {
		return nil
	}
	lastCap := GetLastCap()
	for cap := 0; cap <= lastCap; cap++ {
		unix.Prctl(unix.PR_CAPBSET_DROP, uintptr(cap), 0, 0, 0) //nolint:errcheck
	}
	return nil
}

// KeepOnly keeps only specified capabilities
func (c *CapabilityManager) KeepOnly(caps []uintptr) error {
	if !c.enabled {
		return nil
	}
	if os.Getuid() != 0 {
		return fmt.Errorf("requires root to modify capabilities")
	}
	lastCap := GetLastCap()
	for cap := 0; cap <= lastCap; cap++ {
		keep := false
		for _, keepCap := range caps {
			if uintptr(cap) == keepCap {
				keep = true
				break
			}
		}
		if !keep {
			unix.Prctl(unix.PR_CAPBSET_DROP, uintptr(cap), 0, 0, 0) //nolint:errcheck
		}
	}
	return nil
}

// HasCapability checks if the process has a capability in the bounding set
func (c *CapabilityManager) HasCapability(cap int) bool {
	if !c.enabled {
		return false
	}
	err := unix.Prctl(unix.PR_CAPBSET_READ, uintptr(cap), 0, 0, 0)
	return err == nil
}

// GetCurrentCapabilities returns the current capability set
func (c *CapabilityManager) GetCurrentCapabilities() ([]int, error) {
	if !c.enabled {
		return nil, nil
	}
	var caps []int
	for cap := 0; cap <= GetLastCap(); cap++ {
		if c.HasCapability(cap) {
			caps = append(caps, cap)
		}
	}
	return caps, nil
}

// SetNoNewPrivileges sets the no_new_privs flag on the current process.
// Children spawned after this call inherit the flag.
func (c *CapabilityManager) SetNoNewPrivileges(_ *exec.Cmd) error {
	if !c.enabled {
		return nil
	}
	return unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)
}

// SetSecurebits sets securebits for the process
func (c *CapabilityManager) SetSecurebits(bits int) error {
	if !c.enabled {
		return nil
	}
	return unix.Prctl(unix.PR_SET_SECUREBITS, uintptr(bits), 0, 0, 0)
}

// IsEnabled returns true if capability management is enabled
func (c *CapabilityManager) IsEnabled() bool { return c.enabled }

// Disable disables capability management
func (c *CapabilityManager) Disable() { c.enabled = false }

// GetLastCap returns the highest capability number supported by the kernel
func GetLastCap() int {
	data, err := os.ReadFile("/proc/sys/kernel/cap_last_cap")
	if err != nil {
		return 40
	}
	var lastCap int
	fmt.Sscanf(string(data), "%d", &lastCap)
	return lastCap
}

// EssentialCapabilities returns the minimal set of capabilities needed
func EssentialCapabilities() []uintptr {
	return []uintptr{0, 1, 2, 6, 7, 13}
}

// NetworkCapabilities returns capabilities needed for network operations
func NetworkCapabilities() []uintptr {
	return []uintptr{10, 11, 12, 13}
}

// DangerousCapabilities returns capabilities that should always be dropped
func DangerousCapabilities() []uintptr {
	return []uintptr{15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36}
}

// FormatCapability returns a human-readable name for a capability
func FormatCapability(cap int) string {
	names := map[int]string{
		0: "CAP_CHOWN", 1: "CAP_DAC_OVERRIDE", 2: "CAP_DAC_READ_SEARCH",
		3: "CAP_FOWNER", 4: "CAP_FSETID", 5: "CAP_KILL",
		6: "CAP_SETGID", 7: "CAP_SETUID", 8: "CAP_SETPCAP",
		9: "CAP_LINUX_IMMUTABLE", 10: "CAP_NET_BIND_SERVICE", 11: "CAP_NET_BROADCAST",
		12: "CAP_NET_ADMIN", 13: "CAP_NET_RAW", 14: "CAP_IPC_LOCK",
		15: "CAP_SYS_MODULE", 16: "CAP_SYS_RAWIO", 17: "CAP_SYS_CHROOT",
		18: "CAP_SYS_PTRACE", 19: "CAP_SYS_PACCT", 20: "CAP_SYS_ADMIN",
		21: "CAP_SYS_BOOT", 22: "CAP_SYS_NICE", 23: "CAP_SYS_RESOURCE",
		24: "CAP_SYS_TIME", 25: "CAP_SYS_TTY_CONFIG", 26: "CAP_MKNOD",
		27: "CAP_LEASE", 28: "CAP_AUDIT_WRITE", 29: "CAP_AUDIT_CONTROL",
		30: "CAP_SETFCAP", 31: "CAP_MAC_OVERRIDE", 32: "CAP_MAC_ADMIN",
		33: "CAP_SYSLOG", 34: "CAP_WAKE_ALARM", 35: "CAP_BLOCK_SUSPEND",
		36: "CAP_AUDIT_READ", 37: "CAP_PERFMON", 38: "CAP_BPF",
		39: "CAP_CHECKPOINT_RESTORE",
	}
	if name, ok := names[cap]; ok {
		return name
	}
	return fmt.Sprintf("CAP_UNKNOWN_%d", cap)
}
