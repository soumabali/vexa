package discovery

import (
	"fmt"
	"strings"
)

// OSFingerprint contains fingerprinting data
type OSFingerprint struct {
	Name       string
	Confidence int      // 0-100
	Methods    []string // Detection methods used
}

// TCPWindowSizes for OS fingerprinting
var TCPWindowSizes = map[int]string{
	65535: "Windows/BSD",
	5840:  "Linux",
	5720:  "Linux (older)",
	4128:  "Cisco IOS",
	8760:  "Solaris",
}

// TTLSig contains TTL-based OS signatures
var TTLSig = map[int][]string{
	64:  {"Linux", "FreeBSD", "OpenBSD", "NetBSD", "macOS"},
	128: {"Windows XP/2003", "Windows 7/8/10/11", "Windows Server"},
	255: {"Cisco IOS", "Solaris", "AIX", "HP-UX"},
}

// FingerprintOS attempts to identify the OS from scan results
func FingerprintOS(openPorts []PortResult, hostname string) string {
	var clues []string
	var methods []string

	// 1. Service banner analysis
	for _, port := range openPorts {
		if port.Banner != "" {
			osFromBanner := fingerprintFromBanner(port.Banner)
			if osFromBanner != "" {
				clues = append(clues, osFromBanner)
				methods = append(methods, fmt.Sprintf("banner:%d", port.Port))
			}
		}
	}

	// 2. Port combination analysis
	osFromPorts := fingerprintFromPorts(openPorts)
	if osFromPorts != "" {
		clues = append(clues, osFromPorts)
		methods = append(methods, "port-pattern")
	}

	// 3. Hostname analysis
	if hostname != "" {
		osFromHostname := fingerprintFromHostname(hostname)
		if osFromHostname != "" {
			clues = append(clues, osFromHostname)
			methods = append(methods, "hostname")
		}
	}

	// 4. Service version analysis
	for _, port := range openPorts {
		osFromService := fingerprintFromService(port)
		if osFromService != "" {
			clues = append(clues, osFromService)
			methods = append(methods, fmt.Sprintf("service:%d", port.Port))
		}
	}

	// Return most common guess
	if len(clues) == 0 {
		return "Unknown"
	}

	return mostCommon(clues)
}

// fingerprintFromBanner extracts OS info from service banner
func fingerprintFromBanner(banner string) string {
	banner = strings.ToLower(banner)

	switch {
	case strings.Contains(banner, "openssh"):
		return "Linux/Unix"
	case strings.Contains(banner, "windows"):
		return "Windows"
	case strings.Contains(banner, "ubuntu"):
		return "Ubuntu Linux"
	case strings.Contains(banner, "debian"):
		return "Debian Linux"
	case strings.Contains(banner, "centos"):
		return "CentOS Linux"
	case strings.Contains(banner, "red hat") || strings.Contains(banner, "rhel"):
		return "RHEL Linux"
	case strings.Contains(banner, "freebsd"):
		return "FreeBSD"
	case strings.Contains(banner, "openbsd"):
		return "OpenBSD"
	case strings.Contains(banner, "macos") || strings.Contains(banner, "darwin"):
		return "macOS"
	default:
		return ""
	}
}

// fingerprintFromPorts guesses OS from open port patterns
func fingerprintFromPorts(ports []PortResult) string {
	// Look for OS-specific port combinations
	hasRDP := false
	hasSMB := false
	hasSSH := false
	hasVNC := false
	hasWinRM := false
	
	for _, p := range ports {
		switch p.Port {
		case 3389:
			hasRDP = true
		case 445:
			hasSMB = true
		case 22:
			hasSSH = true
		case 5900:
			hasVNC = true
		case 5985, 5986:
			hasWinRM = true
		}
	}

	switch {
	case hasRDP && hasSMB:
		return "Windows Server"
	case hasRDP && hasWinRM:
		return "Windows (WinRM)"
	case hasRDP:
		return "Windows"
	case hasWinRM:
		return "Windows (WinRM-only)"
	case hasSSH && hasSMB:
		return "Linux (Samba)"
	case hasSSH && !hasRDP:
		return "Linux/Unix"
	case hasVNC && !hasRDP:
		return "Linux/Unix with VNC"
	default:
		return ""
	}
}

// fingerprintFromHostname extracts OS from hostname patterns
func fingerprintFromHostname(hostname string) string {
	hostname = strings.ToLower(hostname)

	switch {
	case strings.Contains(hostname, "win") || strings.Contains(hostname, "windows"):
		return "Windows"
	case strings.Contains(hostname, "ubuntu"):
		return "Ubuntu Linux"
	case strings.Contains(hostname, "debian"):
		return "Debian Linux"
	case strings.Contains(hostname, "centos") || strings.Contains(hostname, "rhel"):
		return "RHEL/CentOS Linux"
	case strings.Contains(hostname, "mac") || strings.Contains(hostname, "darwin"):
		return "macOS"
	case strings.Contains(hostname, "srv") || strings.Contains(hostname, "server"):
		return "Server OS"
	default:
		return ""
	}
}

// fingerprintFromService extracts OS from service-specific signatures
func fingerprintFromService(port PortResult) string {
	switch port.Port {
	case 3389:
		if strings.Contains(strings.ToLower(port.Banner), "rdp") {
			return "Windows"
		}
	case 22:
		// SSH version often reveals OS
		if strings.Contains(strings.ToLower(port.Banner), "windows") {
			return "Windows (OpenSSH)"
		}
	case 5900:
		if strings.Contains(strings.ToLower(port.Banner), "apple") {
			return "macOS (Screen Sharing)"
		}
	}
	return ""
}

// mostCommon returns the most common string in a slice
func mostCommon(items []string) string {
	if len(items) == 0 {
		return ""
	}

	counts := make(map[string]int)
	for _, item := range items {
		counts[item]++
	}

	var maxItem string
	maxCount := 0
	for item, count := range counts {
		if count > maxCount {
			maxCount = count
			maxItem = item
		}
	}

	return maxItem
}
