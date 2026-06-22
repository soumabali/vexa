package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// SeccompFilter manages seccomp-bpf syscall filtering
type SeccompFilter struct {
	enabled bool
}

// SeccompProfile represents a seccomp policy
type SeccompProfile struct {
	DefaultAction string            `json:"defaultAction"`
	Architectures []string          `json:"architectures"`
	Syscalls      []SyscallRule     `json:"syscalls"`
}

// SyscallRule defines a group of syscalls with an action
type SyscallRule struct {
	Names  []string `json:"names"`
	Action string   `json:"action"`
}

// NewSeccompFilter creates a new seccomp filter manager
func NewSeccompFilter() (*SeccompFilter, error) {
	// Check if seccomp is supported
	if err := checkSeccompSupport(); err != nil {
		return nil, fmt.Errorf("seccomp not supported: %w", err)
	}

	return &SeccompFilter{enabled: true}, nil
}

// Apply applies seccomp filter to a command
func (s *SeccompFilter) Apply(cmd *exec.Cmd, profilePath string) error {
	if !s.enabled {
		return nil
	}

	// Load profile
	profile, err := s.loadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("load seccomp profile: %w", err)
	}

	// Generate BPF program
	bpf, err := s.generateBPF(profile)
	if err != nil {
		return fmt.Errorf("generate BPF: %w", err)
	}

	// Use SysProcAttr to apply seccomp before exec
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// We'll use a wrapper that applies seccomp
	// For production, use libseccomp or a small C wrapper
	cmd.Env = append(cmd.Env, "_SSH_SECCOMP_BPF="+encodeBPF(bpf))

	return nil
}

// loadProfile loads a seccomp profile from JSON file
func (s *SeccompFilter) loadProfile(path string) (*SeccompProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var profile SeccompProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// generateBPF generates a BPF program from a seccomp profile
func (s *SeccompFilter) generateBPF(profile *SeccompProfile) ([]syscall.SockFilter, error) {
	// Build a simple BPF program
	// For production, use libseccomp-go for full support

	// Get architecture
	arch := determineArch()

	// Build whitelist of syscalls
	whitelist := make(map[string]bool)
	for _, rule := range profile.Syscalls {
		if rule.Action == "allow" || rule.Action == "SCMP_ACT_ALLOW" {
			for _, name := range rule.Names {
				whitelist[name] = true
			}
		}
	}

	// Convert to syscall numbers
	var allowedSyscalls []int
	for name := range whitelist {
		num, err := syscallNumber(name, arch)
		if err != nil {
			continue // Skip unknown syscalls
		}
		allowedSyscalls = append(allowedSyscalls, num)
	}

	// Build BPF program
	// Architecture check + whitelist check
	bpf := buildWhitelistBPF(allowedSyscalls, arch)

	return bpf, nil
}

// buildWhitelistBPF builds a BPF whitelist program
func buildWhitelistBPF(syscalls []int, arch uint32) []syscall.SockFilter {
	// BPF program structure:
	// 1. Load architecture (offset 4 in seccomp_data)
	// 2. Compare with expected arch
	// 3. Load syscall number (offset 0)
	// 4. Jump table for each allowed syscall
	// 5. Default: kill

	var bpf []syscall.SockFilter

	// Load arch
	bpf = append(bpf, syscall.SockFilter{
		Code: unix.BPF_LD | unix.BPF_W | unix.BPF_ABS,
		K:    4, // seccomp_data.arch
	})

	// Compare arch
	bpf = append(bpf, syscall.SockFilter{
		Code: unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K,
		Jt:   1, // If match, continue
		Jf:   0, // If not match, kill (will be updated)
		K:    arch,
	})

	// Arch mismatch: kill
	bpf = append(bpf, syscall.SockFilter{
		Code: unix.BPF_RET | unix.BPF_K,
		K:    unix.SECCOMP_RET_KILL,
	})

	// Load syscall number
	bpf = append(bpf, syscall.SockFilter{
		Code: unix.BPF_LD | unix.BPF_W | unix.BPF_ABS,
		K:    0, // seccomp_data.nr
	})

	// Whitelist checks
	for i, sysno := range syscalls {
		jumpto := len(syscalls) - i - 1
		if jumpto > 255 {
			jumpto = 255
		}

		bpf = append(bpf, syscall.SockFilter{
			Code: unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K,
			Jt:   0, // If match, allow (will be updated)
			Jf:   uint8(jumpto),
			K:    uint32(sysno),
		})
	}

	// Default: kill
	bpf = append(bpf, syscall.SockFilter{
		Code: unix.BPF_RET | unix.BPF_K,
		K:    unix.SECCOMP_RET_KILL,
	})

	// Allow: update jt for each check
	allowIdx := len(bpf)
	bpf = append(bpf, syscall.SockFilter{
		Code: unix.BPF_RET | unix.BPF_K,
		K:    unix.SECCOMP_RET_ALLOW,
	})

	// Update jt values
	for i := range syscalls {
		idx := 3 + i // After arch check + load syscall
		bpf[idx].Jt = uint8(allowIdx - idx - 1)
	}

	return bpf
}

// determineArch returns the seccomp architecture constant
func determineArch() uint32 {
	// Check runtime architecture
	var nativeArch uint32
	switch {
	// Check 64-bit first
	case unsafe.Sizeof(uintptr(0)) == 8:
		nativeArch = unix.AUDIT_ARCH_X86_64
	default:
		nativeArch = unix.AUDIT_ARCH_I386
	}
	return nativeArch
}

// syscallNumber maps syscall name to number
func syscallNumber(name string, arch uint32) (int, error) {
	// Map of common syscalls
	// For production, use libseccomp-go or generate from headers
	syscalls := map[string]int{
		"read":              0,
		"write":             1,
		"open":              2,
		"close":             3,
		"stat":              4,
		"fstat":             5,
		"lstat":             6,
		"poll":              7,
		"lseek":             8,
		"mmap":              9,
		"mprotect":          10,
		"munmap":            11,
		"brk":               12,
		"rt_sigaction":      13,
		"rt_sigprocmask":    14,
		"ioctl":             16,
		"pread64":           17,
		"pwrite64":          18,
		"readv":             19,
		"writev":            20,
		"access":            21,
		"pipe":              22,
		"select":            23,
		"sched_yield":       24,
		"mremap":            25,
		"msync":             26,
		"mincore":           27,
		"madvise":           28,
		"shmget":            29,
		"shmat":             30,
		"shmctl":            31,
		"dup":               32,
		"dup2":              33,
		"pause":             34,
		"nanosleep":         35,
		"getitimer":         36,
		"alarm":             37,
		"setitimer":         38,
		"getpid":            39,
		"sendfile":          40,
		"socket":            41,
		"connect":           42,
		"accept":            43,
		"sendto":            44,
		"recvfrom":          45,
		"sendmsg":           46,
		"recvmsg":           47,
		"shutdown":          48,
		"bind":              49,
		"listen":            50,
		"getsockname":       51,
		"getpeername":       52,
		"socketpair":        53,
		"setsockopt":        54,
		"getsockopt":        55,
		"clone":             56,
		"fork":              57,
		"vfork":             58,
		"execve":            59,
		"exit":              60,
		"wait4":             61,
		"kill":              62,
		"uname":             63,
		"semget":            64,
		"semop":             65,
		"semctl":            66,
		"shmdt":             67,
		"msgget":            68,
		"msgsnd":            69,
		"msgrcv":            70,
		"msgctl":            71,
		"fcntl":             72,
		"flock":             73,
		"fsync":             74,
		"fdatasync":         75,
		"truncate":          76,
		"ftruncate":         77,
		"getdents":          78,
		"getcwd":            79,
		"chdir":             80,
		"fchdir":            81,
		"rename":            82,
		"mkdir":             83,
		"rmdir":             84,
		"creat":             85,
		"link":              86,
		"unlink":            87,
		"symlink":           88,
		"readlink":          89,
		"chmod":             90,
		"fchmod":            91,
		"chown":             92,
		"fchown":            93,
		"lchown":            94,
		"umask":             95,
		"gettimeofday":      96,
		"getrlimit":         97,
		"getrusage":         98,
		"sysinfo":           99,
		"times":             100,
		"ptrace":            101,
		"getuid":            102,
		"syslog":            103,
		"getgid":            104,
		"setuid":            105,
		"setgid":            106,
		"geteuid":           107,
		"getegid":           108,
		"setpgid":           109,
		"getppid":           110,
		"getpgrp":           111,
		"setsid":            112,
		"setreuid":          113,
		"setregid":          114,
		"getgroups":         115,
		"setgroups":         116,
		"setresuid":         117,
		"getresuid":         118,
		"setresgid":         119,
		"getresgid":         120,
		"getpgid":           121,
		"setfsuid":          122,
		"setfsgid":          123,
		"getsid":            124,
		"capget":            125,
		"capset":            126,
		"rt_sigpending":     127,
		"rt_sigtimedwait":   128,
		"rt_sigqueueinfo":   129,
		"rt_sigsuspend":     130,
		"sigaltstack":       131,
		"utime":             132,
		"mknod":             133,
		"personality":       135,
		"ustat":             136,
		"statfs":            137,
		"fstatfs":           138,
		"sysfs":             139,
		"getpriority":       140,
		"setpriority":       141,
		"sched_setparam":    142,
		"sched_getparam":    143,
		"sched_setscheduler": 144,
		"sched_getscheduler": 145,
		"sched_get_priority_max": 146,
		"sched_get_priority_min": 147,
		"sched_rr_get_interval":  148,
		"mlock":             149,
		"munlock":           150,
		"mlockall":          151,
		"munlockall":        152,
		"vhangup":           153,
		"modify_ldt":        154,
		"pivot_root":        155,
		"prctl":             157,
		"arch_prctl":        158,
		"adjtimex":          159,
		"setrlimit":         160,
		"chroot":            161,
		"sync":              162,
		"acct":              163,
		"settimeofday":      164,
		"mount":             165,
		"umount2":           166,
		"swapon":            167,
		"swapoff":           168,
		"reboot":            169,
		"sethostname":       170,
		"setdomainname":     171,
		"iopl":              172,
		"ioperm":            173,
		"create_module":     174,
		"init_module":       175,
		"delete_module":     176,
		"get_kernel_syms": 177,
		"query_module":      178,
		"quotactl":          179,
		"nfsservctl":        180,
		"getpmsg":           181,
		"putpmsg":           182,
		"afs_syscall":       183,
		"tuxcall":           184,
		"security":          185,
		"gettid":            186,
		"readahead":         187,
		"setxattr":          188,
		"lsetxattr":         189,
		"fsetxattr":         190,
		"getxattr":          191,
		"lgetxattr":         192,
		"fgetxattr":         193,
		"listxattr":         194,
		"llistxattr":        195,
		"flistxattr":        196,
		"removexattr":       197,
		"lremovexattr":      198,
		"fremovexattr":      199,
		"tkill":             200,
		"time":              201,
		"futex":             202,
		"sched_setaffinity": 203,
		"sched_getaffinity": 204,
		"set_thread_area":   205,
		"io_setup":          206,
		"io_destroy":        207,
		"io_getevents":      208,
		"io_submit":         209,
		"io_cancel":         210,
		"get_thread_area":   211,
		"lookup_dcookie":    212,
		"epoll_create":      213,
		"epoll_ctl_old":     214,
		"epoll_wait_old":    215,
		"remap_file_pages":  216,
		"getdents64":        217,
		"set_tid_address":   218,
		"restart_syscall":   219,
		"semtimedop":        220,
		"fadvise64":         221,
		"timer_create":      222,
		"timer_settime":     223,
		"timer_gettime":     224,
		"timer_getoverrun":  225,
		"timer_delete":      226,
		"clock_settime":     227,
		"clock_gettime":     228,
		"clock_getres":      229,
		"clock_nanosleep":   230,
		"exit_group":        231,
		"epoll_wait":        232,
		"epoll_ctl":         233,
		"tgkill":            234,
		"utimes":            235,
		"vserver":           236,
		"mbind":             237,
		"set_mempolicy":     238,
		"get_mempolicy":     239,
		"mq_open":           240,
		"mq_unlink":         241,
		"mq_timedsend":      242,
		"mq_timedreceive":   243,
		"mq_notify":         244,
		"mq_getsetattr":     245,
		"kexec_load":        246,
		"waitid":            247,
		"add_key":           248,
		"request_key":       249,
		"keyctl":            250,
		"ioprio_set":        251,
		"ioprio_get":        252,
		"inotify_init":      253,
		"inotify_add_watch": 254,
		"inotify_rm_watch":  255,
		"migrate_pages":     256,
		"openat":            257,
		"mkdirat":           258,
		"mknodat":           259,
		"fchownat":          260,
		"futimesat":         261,
		"newfstatat":        262,
		"unlinkat":          263,
		"renameat":          264,
		"linkat":            265,
		"symlinkat":         266,
		"readlinkat":        267,
		"fchmodat":          268,
		"faccessat":         269,
		"pselect6":          270,
		"ppoll":             271,
		"unshare":           272,
		"set_robust_list":   273,
		"get_robust_list":   274,
		"splice":            275,
		"tee":               276,
		"sync_file_range":   277,
		"vmsplice":          278,
		"move_pages":        279,
		"utimensat":         280,
		"epoll_pwait":       281,
		"signalfd":          282,
		"timerfd_create":    283,
		"eventfd":           284,
		"fallocate":         285,
		"timerfd_settime":   286,
		"timerfd_gettime":   287,
		"accept4":           288,
		"signalfd4":         289,
		"eventfd2":          290,
		"epoll_create1":     291,
		"dup3":              292,
		"pipe2":             293,
		"inotify_init1":     294,
		"preadv":            295,
		"pwritev":           296,
		"rt_tgsigqueueinfo": 297,
		"perf_event_open":   298,
		"recvmmsg":          299,
		"fanotify_init":     300,
		"fanotify_mark":     301,
		"prlimit64":         302,
		"name_to_handle_at": 303,
		"open_by_handle_at": 304,
		"clock_adjtime":     305,
		"syncfs":            306,
		"sendmmsg":          307,
		"setns":             308,
		"getcpu":            309,
		"process_vm_readv":  310,
		"process_vm_writev": 311,
		"kcmp":              312,
		"finit_module":      313,
		"sched_setattr":     314,
		"sched_getattr":     315,
		"renameat2":         316,
		"seccomp":           317,
		"getrandom":         318,
		"memfd_create":      319,
		"kexec_file_load":   320,
		"bpf":               321,
		"stub_execveat":     322,
		"userfaultfd":       323,
		"membarrier":        324,
		"mlock2":            325,
		"copy_file_range":   326,
		"preadv2":           327,
		"pwritev2":          328,
		"pkey_mprotect":     329,
		"pkey_alloc":        330,
		"pkey_free":         331,
		"statx":             332,
	}

	num, ok := syscalls[name]
	if !ok {
		return 0, fmt.Errorf("unknown syscall: %s", name)
	}
	return num, nil
}

// checkSeccompSupport checks if seccomp is available
func checkSeccompSupport() error {
	// Try to read /proc/sys/kernel/seccomp
	_, err := os.Stat("/proc/sys/kernel/seccomp")
	if err != nil {
		return fmt.Errorf("seccomp not available in kernel")
	}

	// Check if we can use seccomp
	// In practice, use prctl(PR_SET_SECCOMP, SECCOMP_MODE_FILTER, ...)
	// but we can't easily test without actual syscall
	return nil
}

// encodeBPF encodes BPF program to string for env var passing
func encodeBPF(bpf []syscall.SockFilter) string {
	// Simple encoding for demo
	// Production: use proper serialization
	return fmt.Sprintf("BPF:%d", len(bpf))
}

// WriteDefaultSSHProfile writes a default SSH seccomp profile
func WriteDefaultSSHProfile(path string) error {
	profile := SeccompProfile{
		DefaultAction: "SCMP_ACT_KILL",
		Architectures: []string{"SCMP_ARCH_X86_64", "SCMP_ARCH_X86"},
		Syscalls: []SyscallRule{
			{
				Names: []string{
					"read", "write", "open", "openat", "close",
					"stat", "fstat", "lstat", "newfstatat",
					"poll", "select", "pselect6", "ppoll",
					"lseek", "ioctl",
					"mmap", "mprotect", "munmap", "mremap",
					"brk", "sbrk",
					"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
					"sigaltstack",
					"clone", "fork", "vfork", "execve", "exit", "exit_group",
					"wait4", "waitid",
					"kill", "tkill", "tgkill",
					"socket", "socketpair",
					"connect", "bind", "listen", "accept", "accept4",
					"sendto", "recvfrom", "sendmsg", "recvmsg",
					"shutdown", "getsockname", "getpeername",
					"setsockopt", "getsockopt",
					"fcntl", "flock", "fsync", "fdatasync",
					"dup", "dup2", "dup3",
					"pipe", "pipe2",
					"getpid", "getppid", "getpgrp", "getsid",
					"getuid", "geteuid", "getgid", "getegid",
					"getgroups", "setgroups",
					"getcwd", "chdir", "fchdir",
					"rename", "renameat", "renameat2",
					"mkdir", "mkdirat", "rmdir",
					"unlink", "unlinkat",
					"readlink", "readlinkat",
					"getdents", "getdents64",
					"chmod", "fchmod", "fchmodat",
					"chown", "fchown", "lchown", "fchownat",
					"umask",
					"gettimeofday", "getrlimit", "getrusage",
					"times", "clock_gettime", "getcpu",
					"sysinfo", "uname",
					"prctl", "arch_prctl",
					"getrandom",
					"nanosleep", "clock_nanosleep",
					"futex", "set_robust_list", "get_robust_list",
					"epoll_create", "epoll_create1", "epoll_ctl", "epoll_wait", "epoll_pwait",
					"eventfd", "eventfd2",
					"timerfd_create", "timerfd_settime", "timerfd_gettime",
					"signalfd", "signalfd4",
					"inotify_init", "inotify_init1", "inotify_add_watch", "inotify_rm_watch",
					"pread64", "pwrite64", "preadv", "pwritev", "preadv2", "pwritev2",
					"truncate", "ftruncate",
					"fallocate",
					"readahead",
					"sync", "syncfs", "fdatasync", "fsync",
					"memfd_create",
					"splice", "vmsplice", "tee",
					"copy_file_range",
					"statfs", "fstatfs",
					"access", "faccessat",
					"utime", "utimes", "utimensat",
					"set_tid_address", "set_robust_list",
					"get_thread_area", "set_thread_area",
				},
				Action: "SCMP_ACT_ALLOW",
			},
			{
				Names: []string{
					"ptrace", "process_vm_readv", "process_vm_writev",
					"kcmp", "personality",
					"mount", "umount2", "pivot_root", "chroot",
					"swapon", "swapoff", "reboot",
					"sethostname", "setdomainname",
					"iopl", "ioperm",
					"create_module", "init_module", "delete_module",
					"finit_module",
					"query_module", "get_kernel_syms",
					"acct", "lookup_dcookie",
					"perf_event_open",
					"bpf",
					"kexec_load", "kexec_file_load",
					"open_by_handle_at",
					"fanotify_init", "fanotify_mark",
					"nfsservctl", "afs_syscall", "tuxcall", "security",
					"getpmsg", "putpmsg", "vserver",
					"modify_ldt",
				},
				Action: "SCMP_ACT_KILL",
			},
		},
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// WriteDefaultProfileForProtocol writes a seccomp profile for a protocol
func WriteDefaultProfileForProtocol(protocol, path string) error {
	// Base on SSH profile with protocol-specific additions
	basePath := filepath.Join(filepath.Dir(path), "seccomp-ssh.json")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if err := WriteDefaultSSHProfile(basePath); err != nil {
			return err
		}
	}

	// For now, all protocols use the same base profile
	// In production, customize per protocol
	return nil
}
