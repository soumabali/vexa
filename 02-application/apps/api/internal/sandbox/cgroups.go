package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CgroupManager manages cgroups v2 for sandbox processes
type CgroupManager struct {
	basePath string
	enabled  bool
}

// NewCgroupManager creates a new cgroup manager
func NewCgroupManager(basePath string) (*CgroupManager, error) {
	// Check if cgroups v2 is available
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		return nil, fmt.Errorf("cgroups v2 not available: %w", err)
	}

	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("create cgroup base: %w", err)
	}

	return &CgroupManager{
		basePath: basePath,
		enabled:  true,
	}, nil
}

// Apply creates a cgroup for a sandbox with resource limits
func (c *CgroupManager) Apply(sandboxID string, limits ResourceLimits) error {
	if !c.enabled {
		return nil
	}

	cgroupPath := filepath.Join(c.basePath, sandboxID)

	// Create cgroup directory
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return fmt.Errorf("create cgroup: %w", err)
	}

	// Enable controllers
	if err := c.enableControllers(cgroupPath); err != nil {
		// Log warning but continue
		_ = err
	}

	// Set CPU limit (weight-based: 1-10000)
	if limits.CPUPercent > 0 {
		cpuWeight := (limits.CPUPercent * 100) // Scale to cgroup weight
		if cpuWeight < 1 {
			cpuWeight = 1
		}
		if cpuWeight > 10000 {
			cpuWeight = 10000
		}
		if err := writeCgroupFile(cgroupPath, "cpu.weight", strconv.Itoa(cpuWeight)); err != nil {
			_ = err
		}

		// Also set max CPU time (quota)
		quota := fmt.Sprintf("%d %d", limits.CPUPercent*1000, 100000)
		if err := writeCgroupFile(cgroupPath, "cpu.max", quota); err != nil {
			_ = err
		}
	}

	// Set memory limit
	if limits.MemoryMB > 0 {
		memoryBytes := int64(limits.MemoryMB) * 1024 * 1024
		if err := writeCgroupFile(cgroupPath, "memory.max", strconv.FormatInt(memoryBytes, 10)); err != nil {
			_ = err
		}

		// Set memory high (soft limit at 80%)
		highBytes := memoryBytes * 8 / 10
		if err := writeCgroupFile(cgroupPath, "memory.high", strconv.FormatInt(highBytes, 10)); err != nil {
			_ = err
		}

		// Disable OOM killer
		if err := writeCgroupFile(cgroupPath, "memory.oom.group", "0"); err != nil {
			_ = err
		}
	}

	// Set PID limit
	if limits.MaxPIDs > 0 {
		if err := writeCgroupFile(cgroupPath, "pids.max", strconv.Itoa(limits.MaxPIDs)); err != nil {
			_ = err
		}
	}

	// Set file descriptor limit (via pids for now)
	// cgroups v2 doesn't have direct fd limit, use rlimit in process

	// Set IO limits (throttle)
	if limits.DiskMB > 0 {
		// Set IO weight
		if err := writeCgroupFile(cgroupPath, "io.weight", "100"); err != nil {
			_ = err
		}

		// Set max IO bandwidth (read/write)
		// Format: "rbps=N wbps=N riops=N wiops=N"
		ioMax := fmt.Sprintf("rbps=%d wbps=%d", limits.DiskMB*1024*1024, limits.DiskMB*1024*1024)
		if err := writeCgroupFile(cgroupPath, "io.max", ioMax); err != nil {
			_ = err
		}
	}

	return nil
}

// AddProcess adds a process to a cgroup
func (c *CgroupManager) AddProcess(sandboxID string, pid int) error {
	if !c.enabled {
		return nil
	}

	cgroupPath := filepath.Join(c.basePath, sandboxID)
	return writeCgroupFile(cgroupPath, "cgroup.procs", strconv.Itoa(pid))
}

// Remove removes a cgroup
func (c *CgroupManager) Remove(sandboxID string) error {
	if !c.enabled {
		return nil
	}

	cgroupPath := filepath.Join(c.basePath, sandboxID)

	// Kill all processes in the cgroup first
	if err := c.killAll(cgroupPath); err != nil {
		_ = err
	}

	// Wait a bit for processes to exit
	time.Sleep(100 * time.Millisecond)

	// Remove cgroup directory
	if err := os.RemoveAll(cgroupPath); err != nil {
		return fmt.Errorf("remove cgroup: %w", err)
	}

	return nil
}

// GetStats returns cgroup statistics for a sandbox
func (c *CgroupManager) GetStats(sandboxID string) (map[string]string, error) {
	if !c.enabled {
		return nil, fmt.Errorf("cgroups disabled")
	}

	cgroupPath := filepath.Join(c.basePath, sandboxID)
	stats := make(map[string]string)

	// Read CPU stats
	if data, err := readCgroupFile(cgroupPath, "cpu.stat"); err == nil {
		stats["cpu.stat"] = data
	}

	// Read memory stats
	if data, err := readCgroupFile(cgroupPath, "memory.current"); err == nil {
		stats["memory.current"] = data
	}

	if data, err := readCgroupFile(cgroupPath, "memory.peak"); err == nil {
		stats["memory.peak"] = data
	}

	// Read PID stats
	if data, err := readCgroupFile(cgroupPath, "pids.current"); err == nil {
		stats["pids.current"] = data
	}

	// Read IO stats
	if data, err := readCgroupFile(cgroupPath, "io.stat"); err == nil {
		stats["io.stat"] = data
	}

	return stats, nil
}

// enableControllers enables all available controllers for a cgroup
func (c *CgroupManager) enableControllers(cgroupPath string) error {
	// Read available controllers
	controllers, err := readCgroupFile(c.basePath, "cgroup.controllers")
	if err != nil {
		return err
	}

	// Enable all controllers for subtree
	if err := writeCgroupFile(c.basePath, "cgroup.subtree_control", "+"+strings.ReplaceAll(controllers, " ", " +")); err != nil {
		_ = err
	}

	return nil
}

// killAll kills all processes in a cgroup
func (c *CgroupManager) killAll(cgroupPath string) error {
	procs, err := readCgroupFile(cgroupPath, "cgroup.procs")
	if err != nil {
		return err
	}

	for _, line := range strings.Split(procs, "\n") {
		pid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			continue
		}

		// Send SIGKILL
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		proc.Kill()
	}

	return nil
}

// writeCgroupFile writes a value to a cgroup file
func writeCgroupFile(cgroupPath, file, value string) error {
	path := filepath.Join(cgroupPath, file)
	return os.WriteFile(path, []byte(value+"\n"), 0644)
}

// readCgroupFile reads a cgroup file
func readCgroupFile(cgroupPath, file string) (string, error) {
	path := filepath.Join(cgroupPath, file)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SetOOMScore sets OOM score for a process
func SetOOMScore(pid, score int) error {
	path := fmt.Sprintf("/proc/%d/oom_score_adj", pid)
	return os.WriteFile(path, []byte(strconv.Itoa(score)), 0644)
}

// GetCgroupPath returns the cgroup path for a sandbox
func (c *CgroupManager) GetCgroupPath(sandboxID string) string {
	return filepath.Join(c.basePath, sandboxID)
}

// IsEnabled returns true if cgroups is enabled
func (c *CgroupManager) IsEnabled() bool {
	return c.enabled
}
