# Sandboxing & Security Hardening

## Overview

Linux security sandboxing untuk vexa menggunakan multi-layer approach:

| Layer | Technology | Purpose |
|-------|-----------|---------|
| 1 | **Seccomp-bpf** | Syscall filtering |
| 2 | **Cgroups v2** | Resource limits |
| 3 | **Namespaces** | Process/network isolation |
| 4 | **Capabilities** | Privilege dropping |
| 5 | **AppArmor** | Mandatory Access Control |

## Architecture

```
┌─────────────────────────────────────────┐
│           vexa Process           │
├─────────────────────────────────────────┤
│  5. AppArmor Profile (MAC)              │
│  4. Capability Bounding Set (dropped)   │
│  3. Namespaces (pid, net, mount, user)  │
│  2. Cgroups v2 (cpu, mem, pids, io)    │
│  1. Seccomp-bpf (syscall whitelist)     │
├─────────────────────────────────────────┤
│        Protocol Handler (ssh/rdp/vnc)   │
└─────────────────────────────────────────┘
```

## Components

### 1. Seccomp-bpf (`seccomp.go`)

Syscall filtering menggunakan Berkeley Packet Filter.

**Default SSH Whitelist (180+ syscalls):**
- File operations: `read`, `write`, `open`, `close`, `stat`
- Memory: `mmap`, `mprotect`, `munmap`, `brk`
- Network: `socket`, `connect`, `bind`, `listen`, `accept`
- Process: `clone`, `fork`, `execve`, `exit`, `wait4`
- Time: `gettimeofday`, `clock_gettime`, `nanosleep`
- IPC: `pipe`, `pipe2`, `epoll_*`, `futex`

**Kill List (blacklisted):**
- `ptrace`, `process_vm_readv/writev`
- `mount`, `umount2`, `pivot_root`, `chroot`
- `init_module`, `finit_module`, `delete_module`
- `bpf`, `perf_event_open`
- `kexec_load`, `kexec_file_load`

**Default Action:** `SCMP_ACT_KILL` (kill on violation)

### 2. Cgroups v2 (`cgroups.go`)

Resource limits per sandbox:

| Resource | Default SSH | Default RDP | Default VNC |
|----------|-------------|-------------|-------------|
| CPU %    | 100%        | 100%        | 100%        |
| Memory   | 512MB       | 1024MB      | 512MB       |
| Disk IO  | 1GB/s       | 1GB/s       | 1GB/s       |
| Max PIDs | 64          | 128         | 64          |
| Max FDs  | 256         | 512         | 256         |

**Controllers:**
- `cpu.weight` + `cpu.max` (quota)
- `memory.max` + `memory.high` (soft limit)
- `pids.max`
- `io.weight` + `io.max` (throttle)

### 3. Namespaces (`namespaces.go`)

Isolation layers:
- **PID namespace**: Process ID isolation
- **Network namespace**: Network stack isolation
- **Mount namespace**: Filesystem mount isolation
- **IPC namespace**: Inter-process communication isolation
- **User namespace**: UID/GID mapping
- **UTS namespace**: Hostname isolation

**UID/GID Mapping:**
```
Container UID 0 → Host UID (current)
Container GID 0 → Host GID (current)
```

### 4. Capabilities (`caps.go`)

All dangerous capabilities dropped:
- `CAP_SYS_ADMIN` (root-like privileges)
- `CAP_SYS_MODULE` (kernel module loading)
- `CAP_SYS_PTRACE` (process tracing)
- `CAP_SYS_RAWIO` (raw I/O)
- `CAP_MKNOD` (device node creation)
- `CAP_NET_ADMIN` (network admin)
- `CAP_NET_RAW` (raw sockets - dropped unless needed)
- `CAP_SETUID/CAP_SETGID` (dropped after use)

**Kept (minimal):**
- `CAP_NET_BIND_SERVICE` (bind low ports if needed)
- `CAP_CHOWN` (file ownership)

### 5. AppArmor (`apparmor.go`)

Profiles per protocol:

**SSH Profile:**
- `/usr/bin/ssh`, `/usr/bin/ssh-agent`, `/usr/bin/ssh-add`
- `/etc/ssh/**` read
- `/home/*/.ssh/**` read
- `/tmp/**` read/write
- Deny: `/etc/shadow`, `/proc/sys/**`, `/sys/**`, `/dev/mem`

**RDP Profile:**
- `/usr/bin/xfreerdp`, `/usr/bin/remmina`
- X11: `/tmp/.X11-unix/*`
- Similar deny rules

**Mode:** `complain` (testing) → `enforce` (production)

## Integration

### Auto-Sandbox on Session Create

```go
// In session handler
sandboxID := generateID()
profile := "ssh" // or "rdp", "vnc"

sb, err := sandboxMgr.CreateSandbox(sandboxID, profile, sessionID)
if err != nil {
    return err
}

cmd := exec.Command("ssh", args...)
err = sandboxMgr.Start(sb, cmd)
```

### Resource Limits Enforcement

```go
limits := sandbox.ResourceLimits{
    CPUPercent:    100,
    MemoryMB:      512,
    DiskMB:        1024,
    MaxPIDs:       64,
    MaxOpenFiles:  256,
    NetEgressOnly: true,
}
```

### Timeout Handling

- **Session timeout:** 24 hours (auto-kill)
- **Kill timeout:** 5 seconds (SIGTERM → SIGKILL)

## Metrics

Prometheus metrics exported:

| Metric | Type | Description |
|--------|------|-------------|
| `sandbox_total` | Counter | Total sandboxes created |
| `sandbox_active` | Gauge | Active sandboxes |
| `sandbox_errors_total{type}` | Counter | Errors by type |
| `sandbox_cpu_percent` | Gauge | CPU usage |
| `sandbox_memory_bytes` | Gauge | Memory usage |
| `sandbox_seccomp_violations_total` | Counter | Seccomp violations |
| `sandbox_apparmor_violations_total` | Counter | AppArmor violations |

## Alerts

```yaml
- alert: SandboxEscapeAttempt
  expr: rate(sandbox_seccomp_violations_total[5m]) > 0
  for: 1m
  severity: critical

- alert: SandboxResourceExhaustion
  expr: sandbox_memory_percent > 95
  for: 2m
  severity: warning

- alert: TooManySandboxes
  expr: sandbox_active > 1000
  for: 5m
  severity: warning
```

## Testing

```bash
# Run tests
cd apps/api
go test ./internal/sandbox/... -v

# Run with coverage
go test ./internal/sandbox/... -cover -coverprofile=sandbox.out

# Benchmark
go test ./internal/sandbox/... -bench=BenchmarkCreateSandbox
```

## Graceful Degradation

If a sandbox layer fails:
1. Log warning
2. Continue without that layer
3. Never fail the entire session

Example:
```go
if err := m.seccomp.Apply(cmd, profile); err != nil {
    m.metrics.SandboxErrors.WithLabelValues("seccomp").Inc()
    // Log and continue
}
```

## Production Checklist

- [ ] AppArmor mode: `enforce`
- [ ] Seccomp profile: verified on target kernel
- [ ] Cgroups: controllers enabled (`cpu`, `memory`, `pids`, `io`)
- [ ] Namespaces: user namespace enabled
- [ ] Capabilities: verify bounding set
- [ ] Monitoring: Prometheus alerts configured
- [ ] Testing: penetration test sandbox escape attempts
- [ ] Documentation: incident response runbook

## Security Notes

- **No root required** for basic sandbox (user namespace)
- **seccomp**: default action = kill (whitelist approach)
- **cgroups**: OOM killer disabled for graceful handling
- **AppArmor**: deny all by default, explicit allow
- **All layers** are defense in depth — bypassing one doesn't compromise others
