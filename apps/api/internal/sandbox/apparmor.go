package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AppArmorManager manages AppArmor profiles for sandbox processes
type AppArmorManager struct {
	profileDir string
	enabled    bool
	mode       string // "enforce" or "complain"
}

// NewAppArmorManager creates a new AppArmor manager
func NewAppArmorManager(profileDir string) (*AppArmorManager, error) {
	// Check if AppArmor is available
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err != nil {
		return nil, fmt.Errorf("apparmor not available: %w", err)
	}

	// Determine mode based on environment
	mode := "complain"
	if os.Getenv("APPARMOR_MODE") == "enforce" {
		mode = "enforce"
	}

	// Ensure profile directory exists
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return nil, fmt.Errorf("create profile dir: %w", err)
	}

	return &AppArmorManager{
		profileDir: profileDir,
		enabled:    true,
		mode:       mode,
	}, nil
}

// Apply applies an AppArmor profile to a command
func (a *AppArmorManager) Apply(cmd *exec.Cmd, profileName string) error {
	if !a.enabled {
		return nil
	}

	// Check if profile exists, create if not
	profilePath := filepath.Join(a.profileDir, profileName)
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		if err := a.createDefaultProfile(profileName, profilePath); err != nil {
			return fmt.Errorf("create profile: %w", err)
		}
	}

	// Load the profile
	if err := a.loadProfile(profilePath); err != nil {
		return fmt.Errorf("load profile: %w", err)
	}

	// Apply via aa-exec wrapper
	if a.mode == "enforce" {
		cmd.Env = append(cmd.Env, "_APPARMOR_PROFILE="+profileName)
	}

	return nil
}

// loadProfile loads an AppArmor profile into the kernel
func (a *AppArmorManager) loadProfile(path string) error {
	// Check if apparmor_parser is available
	if _, err := exec.LookPath("apparmor_parser"); err != nil {
		return fmt.Errorf("apparmor_parser not found")
	}

	cmd := exec.Command("apparmor_parser", "-r", path)
	if err := cmd.Run(); err != nil {
		// Try with sudo
		cmd = exec.Command("sudo", "apparmor_parser", "-r", path)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("load profile: %w", err)
		}
	}

	return nil
}

// unloadProfile unloads an AppArmor profile
func (a *AppArmorManager) unloadProfile(profileName string) error {
	cmd := exec.Command("apparmor_parser", "-R", filepath.Join(a.profileDir, profileName))
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("sudo", "apparmor_parser", "-R", filepath.Join(a.profileDir, profileName))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("unload profile: %w", err)
		}
	}
	return nil
}

// createDefaultProfile creates a default AppArmor profile for a protocol
func (a *AppArmorManager) createDefaultProfile(name, path string) error {
	var profile string

	switch {
	case strings.Contains(name, "ssh"):
		profile = a.sshProfile(name)
	case strings.Contains(name, "rdp"):
		profile = a.rdpProfile(name)
	case strings.Contains(name, "vnc"):
		profile = a.vncProfile(name)
	default:
		profile = a.defaultProfile(name)
	}

	return os.WriteFile(path, []byte(profile), 0644)
}

// sshProfile returns an AppArmor profile for SSH sessions
func (a *AppArmorManager) sshProfile(name string) string {
	mode := a.mode
	return fmt.Sprintf(`#include <tunables/global>

profile %s flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>
  #include <abstractions/nameservice>
  #include <abstractions/openssl>

  capability net_bind_service,
  capability setuid,
  capability setgid,

  network inet stream,
  network inet6 stream,
  network inet dgram,
  network inet6 dgram,

  /usr/bin/ssh ix,
  /usr/bin/ssh-agent ix,
  /usr/bin/ssh-add ix,

  /etc/ssh/** r,
  /etc/ssl/** r,
  /etc/pki/** r,

  /home/*/.ssh/** r,
  /home/*/.ssh/known_hosts rw,

  /tmp/** rw,
  /var/tmp/** rw,

  /dev/null rw,
  /dev/zero rw,
  /dev/random r,
  /dev/urandom r,
  /dev/tty rw,
  /dev/pts/* rw,

  # Deny dangerous operations
  deny capability sys_admin,
  deny capability mknod,
  deny capability sys_rawio,
  deny capability sys_module,
  deny capability sys_ptrace,
  deny capability sys_pacct,
  deny capability sys_boot,
  deny capability kill,
  deny capability ipc_lock,
  deny capability ipc_owner,
  deny capability lease,
  deny capability linux_immutable,
  deny capability net_admin,
  deny capability net_broadcast,
  deny capability net_raw,
  deny capability setfcap,
  deny capability setpcap,
  deny capability sys_nice,
  deny capability sys_resource,
  deny capability sys_time,
  deny capability syslog,
  deny capability wake_alarm,
  deny capability block_suspend,
  deny capability audit_read,
  deny capability audit_write,
  deny capability audit_control,
  deny capability mac_admin,
  deny capability mac_override,

  # Deny access to sensitive files
  deny /etc/shadow r,
  deny /etc/shadow- r,
  deny /etc/gshadow r,
  deny /etc/gshadow- r,
  deny /etc/passwd w,
  deny /etc/group w,
  deny /etc/sudoers r,
  deny /etc/sudoers.d/ r,
  deny /root/** rwx,
  deny /proc/sys/** w,
  deny /sys/** w,
  deny /dev/mem rw,
  deny /dev/kmem rw,
  deny /dev/port rw,

  # Log mode for testing
  #audit %s /tmp/** rw,
}
`, name, mode)
}

// rdpProfile returns an AppArmor profile for RDP sessions
func (a *AppArmorManager) rdpProfile(name string) string {
	mode := a.mode
	return fmt.Sprintf(`#include <tunables/global>

profile %s flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>
  #include <abstractions/nameservice>
  #include <abstractions/openssl>
  #include <abstractions/X>

  capability net_bind_service,

  network inet stream,
  network inet6 stream,

  /usr/bin/xfreerdp ix,
  /usr/bin/xfreerdp3 ix,
  /usr/bin/remmina ix,

  /etc/ssl/** r,
  /etc/pki/** r,

  /tmp/** rw,
  /var/tmp/** rw,

  /dev/null rw,
  /dev/zero rw,
  /dev/random r,
  /dev/urandom r,
  /dev/tty rw,
  /dev/pts/* rw,

  # X11
  /tmp/.X11-unix/* rw,

  # Deny dangerous operations
  deny capability sys_admin,
  deny capability mknod,
  deny capability sys_rawio,
  deny capability sys_module,
  deny capability sys_ptrace,
  deny capability sys_pacct,
  deny capability sys_boot,
  deny capability kill,
  deny capability ipc_lock,
  deny capability ipc_owner,
  deny capability lease,
  deny capability linux_immutable,
  deny capability net_admin,
  deny capability net_broadcast,
  deny capability net_raw,
  deny capability setfcap,
  deny capability setpcap,
  deny capability sys_nice,
  deny capability sys_resource,
  deny capability sys_time,
  deny capability syslog,
  deny capability wake_alarm,
  deny capability block_suspend,
  deny capability audit_read,
  deny capability audit_write,
  deny capability audit_control,
  deny capability mac_admin,
  deny capability mac_override,

  deny /etc/shadow r,
  deny /etc/shadow- r,
  deny /etc/gshadow r,
  deny /etc/gshadow- r,
  deny /etc/passwd w,
  deny /etc/group w,
  deny /etc/sudoers r,
  deny /etc/sudoers.d/ r,
  deny /root/** rwx,
  deny /proc/sys/** w,
  deny /sys/** w,
  deny /dev/mem rw,
  deny /dev/kmem rw,
  deny /dev/port rw,

  #audit %s /tmp/** rw,
}
`, name, mode)
}

// vncProfile returns an AppArmor profile for VNC sessions
func (a *AppArmorManager) vncProfile(name string) string {
	mode := a.mode
	return fmt.Sprintf(`#include <tunables/global>

profile %s flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>
  #include <abstractions/nameservice>
  #include <abstractions/openssl>

  capability net_bind_service,

  network inet stream,
  network inet6 stream,

  /usr/bin/vncviewer ix,
  /usr/bin/tigervncviewer ix,
  /usr/bin/remmina ix,

  /etc/ssl/** r,
  /etc/pki/** r,

  /tmp/** rw,
  /var/tmp/** rw,

  /dev/null rw,
  /dev/zero rw,
  /dev/random r,
  /dev/urandom r,
  /dev/tty rw,
  /dev/pts/* rw,

  # Deny dangerous operations
  deny capability sys_admin,
  deny capability mknod,
  deny capability sys_rawio,
  deny capability sys_module,
  deny capability sys_ptrace,
  deny capability sys_pacct,
  deny capability sys_boot,
  deny capability kill,
  deny capability ipc_lock,
  deny capability ipc_owner,
  deny capability lease,
  deny capability linux_immutable,
  deny capability net_admin,
  deny capability net_broadcast,
  deny capability net_raw,
  deny capability setfcap,
  deny capability setpcap,
  deny capability sys_nice,
  deny capability sys_resource,
  deny capability sys_time,
  deny capability syslog,
  deny capability wake_alarm,
  deny capability block_suspend,
  deny capability audit_read,
  deny capability audit_write,
  deny capability audit_control,
  deny capability mac_admin,
  deny capability mac_override,

  deny /etc/shadow r,
  deny /etc/shadow- r,
  deny /etc/gshadow r,
  deny /etc/gshadow- r,
  deny /etc/passwd w,
  deny /etc/group w,
  deny /etc/sudoers r,
  deny /etc/sudoers.d/ r,
  deny /root/** rwx,
  deny /proc/sys/** w,
  deny /sys/** w,
  deny /dev/mem rw,
  deny /dev/kmem rw,
  deny /dev/port rw,

  #audit %s /tmp/** rw,
}
`, name, mode)
}

// defaultProfile returns a generic restrictive profile
func (a *AppArmorManager) defaultProfile(name string) string {
	return fmt.Sprintf(`#include <tunables/global>

profile %s flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>
  #include <abstractions/nameservice>

  capability net_bind_service,

  network inet stream,
  network inet6 stream,

  /tmp/** rw,
  /var/tmp/** rw,

  /dev/null rw,
  /dev/zero rw,
  /dev/random r,
  /dev/urandom r,
  /dev/tty rw,
  /dev/pts/* rw,

  deny capability sys_admin,
  deny capability mknod,
  deny capability sys_rawio,
  deny capability sys_module,
  deny capability sys_ptrace,
  deny capability sys_pacct,
  deny capability sys_boot,
  deny capability kill,
  deny capability setuid,
  deny capability setgid,
  deny capability setpcap,
  deny capability setfcap,
  deny capability net_admin,
  deny capability net_raw,
  deny capability net_broadcast,
  deny capability ipc_lock,
  deny capability ipc_owner,
  deny capability sys_nice,
  deny capability sys_resource,
  deny capability sys_time,
  deny capability syslog,
  deny capability lease,
  deny capability linux_immutable,
  deny capability wake_alarm,
  deny capability block_suspend,
  deny capability audit_read,
  deny capability audit_write,
  deny capability audit_control,
  deny capability mac_admin,
  deny capability mac_override,

  deny /etc/shadow r,
  deny /etc/shadow- r,
  deny /etc/gshadow r,
  deny /etc/gshadow- r,
  deny /etc/passwd w,
  deny /etc/group w,
  deny /etc/sudoers r,
  deny /root/** rwx,
  deny /proc/sys/** w,
  deny /sys/** w,
  deny /dev/mem rw,
  deny /dev/kmem rw,
  deny /dev/port rw,
}
`, name)
}

// GetMode returns the current AppArmor mode
func (a *AppArmorManager) GetMode() string {
	return a.mode
}

// SetMode sets the AppArmor mode (enforce or complain)
func (a *AppArmorManager) SetMode(mode string) {
	a.mode = mode
}

// IsEnabled returns true if AppArmor is enabled
func (a *AppArmorManager) IsEnabled() bool {
	return a.enabled
}

// GenerateProfile generates a custom AppArmor profile for a protocol
func (a *AppArmorManager) GenerateProfile(name, protocol string, extraRules []string) (string, error) {
	base := a.defaultProfile(name)

	// Add protocol-specific rules
	var rules strings.Builder
	for _, rule := range extraRules {
		rules.WriteString("  ")
		rules.WriteString(rule)
		rules.WriteString("\n")
	}

	// Insert rules before the closing brace
	base = strings.Replace(base, "}\n", rules.String()+"}\n", 1)

	return base, nil
}

// ValidateProfile validates an AppArmor profile syntax
func (a *AppArmorManager) ValidateProfile(path string) error {
	if _, err := exec.LookPath("apparmor_parser"); err != nil {
		return fmt.Errorf("apparmor_parser not found")
	}

	cmd := exec.Command("apparmor_parser", "-p", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	return nil
}

// ListProfiles returns all loaded AppArmor profiles
func (a *AppArmorManager) ListProfiles() ([]string, error) {
	data, err := os.ReadFile("/sys/kernel/security/apparmor/profiles")
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			profiles = append(profiles, line)
		}
	}

	return profiles, nil
}
