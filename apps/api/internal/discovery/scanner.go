package discovery

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ScanResult represents a single host discovery result
type ScanResult struct {
	ID           uuid.UUID `json:"id"`
	IPAddress    string    `json:"ip_address"`
	Hostname     string    `json:"hostname,omitempty"`
	OpenPorts    []PortResult `json:"open_ports"`
	OSGuess      string    `json:"os_guess,omitempty"`
	ResponseTime int64     `json:"response_time_ms"`
	Status       string    `json:"status"` // "up", "down", "filtered"
	ScannedAt    time.Time `json:"scanned_at"`
}

// PortResult represents an open port finding
type PortResult struct {
	Port        int    `json:"port"`
	Service     string `json:"service,omitempty"`
	Banner      string `json:"banner,omitempty"`
	Protocol    string `json:"protocol"` // "tcp", "udp"
	TLS         bool   `json:"tls,omitempty"`
}

// ScanJob represents an active scan job
type ScanJob struct {
	ID          uuid.UUID     `json:"id"`
	Network     string        `json:"network"`      // CIDR, e.g., "192.168.1.0/24"
	Ports       []int         `json:"ports"`
	Status      string        `json:"status"`       // "pending", "running", "completed", "cancelled"
	Progress    int           `json:"progress"`     // 0-100
	Results     []ScanResult  `json:"results"`
	TotalHosts  int           `json:"total_hosts"`
	ScannedHosts int          `json:"scanned_hosts"`
	CreatedAt   time.Time     `json:"created_at"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Error       string        `json:"error,omitempty"`
	Options     ScanOptions   `json:"options"`
}

// ScanOptions configures scan behavior
type ScanOptions struct {
	TimeoutMs        int  `json:"timeout_ms"`        // Default: 1000
	RateLimitMs      int  `json:"rate_limit_ms"`     // Default: 100
	Concurrency      int  `json:"concurrency"`       // Default: 50
	ICMPFirst        bool `json:"icmp_first"`        // ICMP ping before TCP scan
	OSScan           bool `json:"os_scan"`           // Enable OS fingerprinting
	ResolveHostname  bool `json:"resolve_hostname"`  // Reverse DNS lookup
}

// DefaultScanOptions returns default options
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		TimeoutMs:       1000,
		RateLimitMs:     100,
		Concurrency:     50,
		ICMPFirst:       true,
		OSScan:          true,
		ResolveHostname: true,
	}
}

// KnownServices maps common ports to service names
var KnownServices = map[int]string{
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	143:   "IMAP",
	443:   "HTTPS",
	445:   "SMB",
	993:   "IMAPS",
	995:   "POP3S",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	5900:  "VNC",
	6379:  "Redis",
	8080:  "HTTP-Alt",
	8443:  "HTTPS-Alt",
	9200:  "Elasticsearch",
	27017: "MongoDB",
}

// DefaultPorts to scan
var DefaultPorts = []int{22, 80, 443, 3389, 5900, 5432, 3306, 6379, 8080, 8443}

// Scanner manages network discovery scans
type Scanner struct {
	jobs      map[uuid.UUID]*ScanJob
	jobsMu    sync.RWMutex
	tcpProber *TCPProber
	icmpProber *ICMPProber
}

// NewScanner creates a new discovery scanner
func NewScanner() *Scanner {
	return &Scanner{
		jobs:       make(map[uuid.UUID]*ScanJob),
		tcpProber:  NewTCPProber(),
		icmpProber: NewICMPProber(),
	}
}

// CreateJob creates a new scan job
func (s *Scanner) CreateJob(network string, ports []int, options ScanOptions) (*ScanJob, error) {
	if network == "" {
		return nil, fmt.Errorf("network range is required")
	}

	// Validate CIDR
	ip, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		return nil, fmt.Errorf("invalid network CIDR: %w", err)
	}

	// Skip private/reserved IPs check (configurable)
	if isPrivateOrReserved(ip) {
		// Allow but warn - scanning private networks is expected use case
		// Skip loopback and link-local only
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return nil, fmt.Errorf("cannot scan loopback or link-local addresses: %s", network)
		}
	}

	// Count total hosts
	var totalHosts int
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		totalHosts++
		if totalHosts >= 65536 { // Max /16
			return nil, fmt.Errorf("network too large, max /16 supported")
		}
	}

	if totalHosts == 0 {
		return nil, fmt.Errorf("no hosts in network range")
	}

	// Use default ports if none specified
	if len(ports) == 0 {
		ports = DefaultPorts
	}

	job := &ScanJob{
		ID:          uuid.New(),
		Network:     network,
		Ports:       ports,
		Status:      "pending",
		Progress:    0,
		Results:     []ScanResult{},
		TotalHosts:  totalHosts,
		CreatedAt:   time.Now(),
		Options:     options,
	}

	s.jobsMu.Lock()
	s.jobs[job.ID] = job
	s.jobsMu.Unlock()

	return job, nil
}

// StartJob begins scanning
func (s *Scanner) StartJob(ctx context.Context, jobID uuid.UUID) error {
	s.jobsMu.Lock()
	job, exists := s.jobs[jobID]
	if !exists {
		s.jobsMu.Unlock()
		return fmt.Errorf("job not found")
	}
	if job.Status != "pending" {
		s.jobsMu.Unlock()
		return fmt.Errorf("job is not pending (status: %s)", job.Status)
	}

	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	s.jobsMu.Unlock()

	go s.executeScan(ctx, job)
	return nil
}

// GetJob retrieves a scan job
func (s *Scanner) GetJob(jobID uuid.UUID) (*ScanJob, bool) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	job, exists := s.jobs[jobID]
	if !exists {
		return nil, false
	}
	// Return copy to avoid race conditions
	jobCopy := *job
	resultsCopy := make([]ScanResult, len(job.Results))
	copy(resultsCopy, job.Results)
	jobCopy.Results = resultsCopy
	return &jobCopy, true
}

// CancelJob stops an active scan
func (s *Scanner) CancelJob(jobID uuid.UUID) error {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found")
	}
	if job.Status != "running" {
		return fmt.Errorf("job is not running")
	}
	job.Status = "cancelled"
	now := time.Now()
	job.CompletedAt = &now
	return nil
}

// ListJobs returns all jobs (sorted by created_at desc)
func (s *Scanner) ListJobs() []*ScanJob {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	jobs := make([]*ScanJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobCopy := *job
		resultsCopy := make([]ScanResult, len(job.Results))
		copy(resultsCopy, job.Results)
		jobCopy.Results = resultsCopy
		jobs = append(jobs, &jobCopy)
	}
	return jobs
}

// executeScan performs the actual scanning
func (s *Scanner) executeScan(ctx context.Context, job *ScanJob) {
	defer func() {
		now := time.Now()
		job.CompletedAt = &now
		if job.Status == "running" {
			job.Status = "completed"
		}
	}()

	// Parse network
	_, ipnet, err := net.ParseCIDR(job.Network)
	if err != nil {
		job.Error = err.Error()
		job.Status = "failed"
		return
	}

	// Prepare IP list
	var ips []net.IP
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ips = append(ips, dupIP(ip))
	}

	// Progress tracking
	total := len(ips)
	scanned := 0
	var resultsMu sync.Mutex

	// Semaphore for concurrency
	sem := make(chan struct{}, job.Options.Concurrency)
	var wg sync.WaitGroup

	for _, ip := range ips {
		// Check cancellation
		s.jobsMu.RLock()
		status := job.Status
		s.jobsMu.RUnlock()
		if status != "running" {
			return
		}

		sem <- struct{}{}
		wg.Add(1)

		go func(targetIP net.IP) {
			defer wg.Done()
			defer func() { <-sem }()

			// Rate limiting
			if job.Options.RateLimitMs > 0 {
				time.Sleep(time.Duration(job.Options.RateLimitMs) * time.Millisecond)
			}

			result := s.scanHost(ctx, targetIP, job.Ports, job.Options)

			resultsMu.Lock()
			if result.Status == "up" {
				job.Results = append(job.Results, result)
			}
			scanned++
			job.ScannedHosts = scanned
			job.Progress = (scanned * 100) / total
			resultsMu.Unlock()
		}(ip)
	}

	wg.Wait()
}

// scanHost scans a single host
func (s *Scanner) scanHost(ctx context.Context, ip net.IP, ports []int, opts ScanOptions) ScanResult {
	result := ScanResult{
		ID:        uuid.New(),
		IPAddress: ip.String(),
		Status:    "down",
		ScannedAt: time.Now(),
	}

	start := time.Now()

	// ICMP probe first (if enabled)
	if opts.ICMPFirst {
		alive, rtt := s.icmpProber.Probe(ip.String(), opts.TimeoutMs)
		if alive {
			result.Status = "up"
			result.ResponseTime = rtt
		} else {
			// Host might be up but blocking ICMP, try TCP anyway
			result.Status = "filtered"
		}
	}

	// TCP port scanning
	var openPorts []PortResult
	for _, port := range ports {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		if s.tcpProber.IsOpen(ip.String(), port, opts.TimeoutMs) {
			result.Status = "up"
			portResult := PortResult{
				Port:     port,
				Service:  KnownServices[port],
				Protocol: "tcp",
			}

			// Banner grab for SSH/RDP/VNC
			if port == 22 || port == 3389 || port == 5900 {
				banner := s.tcpProber.GrabBanner(ip.String(), port, opts.TimeoutMs)
				if banner != "" {
					portResult.Banner = banner
				}
			}

			// TLS detection for HTTPS
			if port == 443 || port == 8443 {
				portResult.TLS = true
			}

			openPorts = append(openPorts, portResult)
		}
	}

	result.OpenPorts = openPorts
	result.ResponseTime = time.Since(start).Milliseconds()

	// Hostname resolution
	if opts.ResolveHostname {
		names, err := net.LookupAddr(ip.String())
		if err == nil && len(names) > 0 {
			result.Hostname = strings.TrimSuffix(names[0], ".")
		}
	}

	// OS fingerprinting
	if opts.OSScan && len(openPorts) > 0 {
		result.OSGuess = FingerprintOS(openPorts, result.Hostname)
	}

	return result
}

// incrementIP increments an IP address
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// dupIP creates a copy of an IP
func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

// isPrivateOrReserved checks if IP is private/reserved
func isPrivateOrReserved(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsMulticast()
}

// ImportResults imports scan results into hosts table
func (s *Scanner) ImportResults(jobID uuid.UUID, userID string, importAll bool, selectedIPs []string) ([]string, error) {
	s.jobsMu.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("job not found")
	}
	if job.Status != "completed" {
		return nil, fmt.Errorf("job not completed")
	}

	var imported []string
	for _, result := range job.Results {
		if result.Status != "up" {
			continue
		}

		if !importAll {
			found := false
			for _, ip := range selectedIPs {
				if ip == result.IPAddress {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Determine protocol from open ports
		protocol := "ssh"
		for _, port := range result.OpenPorts {
			if port.Port == 3389 {
				protocol = "rdp"
				break
			} else if port.Port == 5900 {
				protocol = "vnc"
				break
			}
		}

		imported = append(imported, result.IPAddress)
		_ = userID // Will be used for actual DB insert
		_ = protocol
		// TODO: Insert into hosts table via host service
	}

	return imported, nil
}
