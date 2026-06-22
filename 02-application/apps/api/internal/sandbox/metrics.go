package sandbox

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds Prometheus metrics for sandbox monitoring
type Metrics struct {
	// Sandbox lifecycle
	SandboxTotal    prometheus.Counter
	SandboxActive   prometheus.Gauge
	SandboxStarted  prometheus.Counter
	SandboxStopped  prometheus.Counter
	SandboxErrors   *prometheus.CounterVec

	// Resource usage
	CPUPercent    prometheus.Gauge
	MemoryBytes   prometheus.Gauge
	MemoryPercent prometheus.Gauge
	PIDs          prometheus.Gauge
	OpenFiles     prometheus.Gauge

	// Security events
	SeccompViolations   prometheus.Counter
	AppArmorViolations  prometheus.Counter
	CapDrops            prometheus.Counter
	NamespaceErrors     prometheus.Counter

	// Protocol distribution
	ProtocolSSH prometheus.Gauge
	ProtocolRDP prometheus.Gauge
	ProtocolVNC prometheus.Gauge
}

// NewMetrics creates new sandbox metrics
func NewMetrics(reg *prometheus.Registry) *Metrics {
	m := &Metrics{
		SandboxTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sandbox_total",
			Help: "Total number of sandboxes created",
		}),
		SandboxActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_active",
			Help: "Number of currently active sandboxes",
		}),
		SandboxStarted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sandbox_started_total",
			Help: "Total number of sandboxes started",
		}),
		SandboxStopped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sandbox_stopped_total",
			Help: "Total number of sandboxes stopped",
		}),
		SandboxErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sandbox_errors_total",
			Help: "Total number of sandbox errors",
		}, []string{"type"}),
		CPUPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_cpu_percent",
			Help: "CPU usage percentage",
		}),
		MemoryBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_memory_bytes",
			Help: "Memory usage in bytes",
		}),
		MemoryPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_memory_percent",
			Help: "Memory usage percentage of limit",
		}),
		PIDs: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_pids",
			Help: "Number of PIDs in sandbox",
		}),
		OpenFiles: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_open_files",
			Help: "Number of open files",
		}),
		SeccompViolations: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sandbox_seccomp_violations_total",
			Help: "Total seccomp violations",
		}),
		AppArmorViolations: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sandbox_apparmor_violations_total",
			Help: "Total AppArmor violations",
		}),
		CapDrops: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sandbox_cap_drops_total",
			Help: "Total capability drops",
		}),
		NamespaceErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sandbox_namespace_errors_total",
			Help: "Total namespace errors",
		}),
		ProtocolSSH: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_protocol_ssh",
			Help: "Active SSH sandboxes",
		}),
		ProtocolRDP: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_protocol_rdp",
			Help: "Active RDP sandboxes",
		}),
		ProtocolVNC: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sandbox_protocol_vnc",
			Help: "Active VNC sandboxes",
		}),
	}

	// Register all metrics
	reg.MustRegister(
		m.SandboxTotal,
		m.SandboxActive,
		m.SandboxStarted,
		m.SandboxStopped,
		m.SandboxErrors,
		m.CPUPercent,
		m.MemoryBytes,
		m.MemoryPercent,
		m.PIDs,
		m.OpenFiles,
		m.SeccompViolations,
		m.AppArmorViolations,
		m.CapDrops,
		m.NamespaceErrors,
		m.ProtocolSSH,
		m.ProtocolRDP,
		m.ProtocolVNC,
	)

	return m
}
