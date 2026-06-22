package api

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$`)
	ipRegex       = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email) && len(email) <= 255
}

func IsValidHostname(hostname string) bool {
	return hostnameRegex.MatchString(hostname) && len(hostname) <= 253
}

func IsValidIP(ip string) bool {
	return ipRegex.MatchString(ip)
}

func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

func IsValidPassword(password string) bool {
	if len(password) < 12 || len(password) > 128 {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^*()_+-=[]{}|;:,.?", c):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func IsValidTOTP(code string) bool {
	if len(code) != 6 {
		return false
	}
	_, err := strconv.Atoi(code)
	return err == nil
}

func SanitizeString(s string) string {
	if len(s) > 10000 {
		return s[:10000]
	}
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}
