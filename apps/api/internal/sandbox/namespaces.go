package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// NamespaceManager manages Linux namespaces for sandboxing
type NamespaceManager struct {
	enabled bool
}

// NewNamespaceManager creates a new namespace manager
func NewNamespaceManager() *NamespaceManager {
	return &NamespaceManager{
		enabled: true,
	}
}

// Apply applies namespace isolation to a command
func (n *NamespaceManager) Apply(cmd *exec.Cmd) error {
	if !n.enabled {
		return nil
	}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// Clone flags for namespaces
	cmd.SysProcAttr.Cloneflags = syscall.CLONE_NEWIPC |
		syscall.CLONE_NEWNET |
		syscall.CLONE_NEWNS |
		syscall.CLONE_NEWPID |
		syscall.CLONE_NEWUSER |
		syscall.CLONE_NEWUTS

	// Setup UID/GID mapping for user namespace
	cmd.SysProcAttr.UidMappings = []syscall.SysProcIDMap{
		{
			ContainerID: 0,
			HostID:      os.Getuid(),
			Size:        1,
		},
	}

	cmd.SysProcAttr.GidMappings = []syscall.SysProcIDMap{
		{
			ContainerID: 0,
			HostID:      os.Getgid(),
			Size:        1,
		},
	}

	// Unshare additional namespaces
	cmd.SysProcAttr.Unshareflags = syscall.CLONE_NEWIPC |
		syscall.CLONE_NEWNET |
		syscall.CLONE_NEWNS |
		syscall.CLONE_NEWPID |
		syscall.CLONE_NEWUSER |
		syscall.CLONE_NEWUTS

	return nil
}

// SetupRootFS sets up a minimal root filesystem for the namespace
func (n *NamespaceManager) SetupRootFS(root string) error {
	if !n.enabled {
		return nil
	}

	// Create minimal rootfs structure
	dirs := []string{
		"proc", "sys", "dev", "tmp", "etc",
		"bin", "sbin", "lib", "lib64",
		"usr", "usr/bin", "usr/sbin", "usr/lib", "usr/lib64",
		"run", "var", "var/tmp",
	}

	for _, dir := range dirs {
		path := fmt.Sprintf("%s/%s", root, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	// Mount proc
	if err := syscall.Mount("proc", fmt.Sprintf("%s/proc", root), "proc", 0, ""); err != nil {
		_ = err // May fail without root, log warning
	}

	// Mount sysfs (read-only)
	if err := syscall.Mount("sysfs", fmt.Sprintf("%s/sys", root), "sysfs", syscall.MS_RDONLY, ""); err != nil {
		_ = err
	}

	// Mount tmpfs for /dev and /tmp
	if err := syscall.Mount("tmpfs", fmt.Sprintf("%s/dev", root), "tmpfs", 0, "size=10M"); err != nil {
		_ = err
	}

	if err := syscall.Mount("tmpfs", fmt.Sprintf("%s/tmp", root), "tmpfs", 0, "size=100M"); err != nil {
		_ = err
	}

	// Create essential device nodes
	nodes := []struct {
		path string
		mode uint32
		dev  int
	}{
		{"/dev/null", syscall.S_IFCHR | 0666, 0x0103},
		{"/dev/zero", syscall.S_IFCHR | 0666, 0x0105},
		{"/dev/random", syscall.S_IFCHR | 0666, 0x0108},
		{"/dev/urandom", syscall.S_IFCHR | 0666, 0x0109},
		{"/dev/tty", syscall.S_IFCHR | 0666, 0x0500},
	}

	for _, node := range nodes {
		path := fmt.Sprintf("%s%s", root, node.path)
		if err := syscall.Mknod(path, node.mode, node.dev); err != nil {
			_ = err
		}
	}

	return nil
}

// CleanupRootFS unmounts the root filesystem
func (n *NamespaceManager) CleanupRootFS(root string) error {
	if !n.enabled {
		return nil
	}

	// Unmount in reverse order
	mounts := []string{"/tmp", "/dev", "/sys", "/proc"}
	for _, mount := range mounts {
		path := fmt.Sprintf("%s%s", root, mount)
		if err := syscall.Unmount(path, 0); err != nil {
			_ = err
		}
	}

	return nil
}

// SetupNetwork sets up network namespace with limited connectivity
func (n *NamespaceManager) SetupNetwork() error {
	if !n.enabled {
		return nil
	}

	// In a full implementation, this would:
	// 1. Create veth pair
	// 2. Attach to bridge
	// 3. Set up NAT
	// 4. Configure iptables for egress-only

	// For now, just log that network namespace was created
	return nil
}

// PivotRoot pivots to a new root filesystem (requires root)
func (n *NamespaceManager) PivotRoot(newroot string) error {
	if !n.enabled {
		return nil
	}

	// Ensure newroot is a mount point
	if err := syscall.Mount(newroot, newroot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount: %w", err)
	}

	// Create oldroot directory
	oldroot := fmt.Sprintf("%s/.oldroot", newroot)
	if err := os.MkdirAll(oldroot, 0700); err != nil {
		return fmt.Errorf("create oldroot: %w", err)
	}

	// Pivot root
	if err := syscall.PivotRoot(newroot, oldroot); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	// Change to new root
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir: %w", err)
	}

	// Unmount old root
	if err := syscall.Unmount("/.oldroot", syscall.MNT_DETACH); err != nil {
		_ = err
	}

	return nil
}

// IsEnabled returns true if namespaces are enabled
func (n *NamespaceManager) IsEnabled() bool {
	return n.enabled
}

// Disable disables namespace management
func (n *NamespaceManager) Disable() {
	n.enabled = false
}

// GetSupportedNamespaces returns the namespaces supported by the kernel
func GetSupportedNamespaces() []string {
	var namespaces []string

	// Check /proc/self/ns for supported namespaces
	nsDir := "/proc/self/ns"
	entries, err := os.ReadDir(nsDir)
	if err != nil {
		return namespaces
	}

	for _, entry := range entries {
		if entry.Type() == os.ModeSymlink {
			namespaces = append(namespaces, entry.Name())
		}
	}

	return namespaces
}
