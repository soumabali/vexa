package validators

import (
	"fmt"
	"math"
	"net"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Validator holds all validation functions
type Validator struct {
	errors []string
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{errors: make([]string, 0)}
}

// Valid returns true if no validation errors
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// Errors returns all validation errors
func (v *Validator) Errors() []string {
	return v.errors
}

// Error returns combined validation errors
func (v *Validator) Error() string {
	return strings.Join(v.errors, "; ")
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, fmt.Sprintf("%s: %s", field, message))
}

// ============================================
// EMAIL VALIDATION
// ============================================

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) bool {
	if email == "" {
		v.AddError("email", "email is required")
		return false
	}

	if len(email) > 254 {
		v.AddError("email", "email must be less than 254 characters")
		return false
	}

	if !emailRegex.MatchString(email) {
		v.AddError("email", "invalid email format")
		return false
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		v.AddError("email", "invalid email address")
		return false
	}

	return true
}

// ============================================
// IP ADDRESS VALIDATION
// ============================================

// ValidateIP validates IP address (IPv4 or IPv6)
func (v *Validator) ValidateIP(ip string) bool {
	if ip == "" {
		v.AddError("ip", "IP address is required")
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		v.AddError("ip", "invalid IP address format")
		return false
	}

	// Reject private IPs in production (optional)
	// if isPrivateIP(parsedIP) {
	//     v.AddError("ip", "private IP addresses not allowed")
	//     return false
	// }

	return true
}

// ValidateIPv4 validates IPv4 address
func (v *Validator) ValidateIPv4(ip string) bool {
	if ip == "" {
		v.AddError("ip", "IPv4 address is required")
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.To4() == nil {
		v.AddError("ip", "invalid IPv4 address")
		return false
	}

	return true
}

// ValidateIPv6 validates IPv6 address
func (v *Validator) ValidateIPv6(ip string) bool {
	if ip == "" {
		v.AddError("ip", "IPv6 address is required")
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.To4() != nil {
		v.AddError("ip", "invalid IPv6 address")
		return false
	}

	return true
}

// ============================================
// PORT VALIDATION
// ============================================

// ValidatePort validates TCP/UDP port number
func (v *Validator) ValidatePort(port int) bool {
	if port < 1 || port > 65535 {
		v.AddError("port", "port must be between 1 and 65535")
		return false
	}

	// Warn about well-known ports (optional)
	if port < 1024 {
		// v.AddError("port", "well-known ports require elevated privileges")
		// return false
	}

	return true
}

// ValidatePortString validates port from string
func (v *Validator) ValidatePortString(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		v.AddError("port", "port must be a valid integer")
		return false
	}
	return v.ValidatePort(port)
}

// ============================================
// HOSTNAME VALIDATION
// ============================================

var hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)
var fqdnRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// ValidateHostname validates hostname
func (v *Validator) ValidateHostname(hostname string) bool {
	if hostname == "" {
		v.AddError("hostname", "hostname is required")
		return false
	}

	if len(hostname) > 253 {
		v.AddError("hostname", "hostname must be less than 253 characters")
		return false
	}

	// Allow IP addresses
	if net.ParseIP(hostname) != nil {
		return true
	}

	// Validate FQDN
	if !fqdnRegex.MatchString(hostname) && !hostnameRegex.MatchString(hostname) {
		v.AddError("hostname", "invalid hostname format")
		return false
	}

	return true
}

// ============================================
// SSH KEY VALIDATION
// ============================================

var sshKeyPrefixes = []string{
	"ssh-rsa",
	"ssh-ed25519",
	"ecdsa-sha2-nistp256",
	"ecdsa-sha2-nistp384",
	"ecdsa-sha2-nistp521",
	"sk-ssh-ed25519@openssh.com",
	"sk-ecdsa-sha2-nistp256@openssh.com",
}

// ValidateSSHKey validates SSH public key format
func (v *Validator) ValidateSSHKey(key string) bool {
	if key == "" {
		v.AddError("ssh_key", "SSH key is required")
		return false
	}

	parts := strings.Fields(key)
	if len(parts) < 2 {
		v.AddError("ssh_key", "invalid SSH key format")
		return false
	}

	prefix := parts[0]
	validPrefix := false
	for _, p := range sshKeyPrefixes {
		if prefix == p {
			validPrefix = true
			break
		}
	}

	if !validPrefix {
		v.AddError("ssh_key", fmt.Sprintf("unsupported SSH key type: %s", prefix))
		return false
	}

	// Base64 validation (simplified)
	if len(parts[1]) < 20 {
		v.AddError("ssh_key", "SSH key appears too short")
		return false
	}

	return true
}

// ============================================
// PASSWORD STRENGTH VALIDATION
// ============================================

type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	MinEntropy     float64
}

// DefaultPasswordPolicy returns default password policy
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      10,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		MinEntropy:     45.0,
	}
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(password string, policy PasswordPolicy) bool {
	valid := true

	if len(password) < policy.MinLength {
		v.AddError("password", fmt.Sprintf("password must be at least %d characters", policy.MinLength))
		valid = false
	}

	if len(password) > policy.MaxLength {
		v.AddError("password", fmt.Sprintf("password must be less than %d characters", policy.MaxLength))
		valid = false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUpper && !hasUpper {
		v.AddError("password", "password must contain at least one uppercase letter")
		valid = false
	}

	if policy.RequireLower && !hasLower {
		v.AddError("password", "password must contain at least one lowercase letter")
		valid = false
	}

	if policy.RequireDigit && !hasDigit {
		v.AddError("password", "password must contain at least one digit")
		valid = false
	}

	if policy.RequireSpecial && !hasSpecial {
		v.AddError("password", "password must contain at least one special character")
		valid = false
	}

	// Entropy check (simplified)
	entropy := calculateEntropy(password)
	if policy.MinEntropy > 0 && entropy < policy.MinEntropy {
		v.AddError("password", fmt.Sprintf("password entropy too low (%.1f, minimum %.1f)", entropy, policy.MinEntropy))
		valid = false
	}

	// Reject common predictable patterns regardless of entropy.
	lower := strings.ToLower(password)
	commonPasswords := []string{"password", "qwerty", "admin", "login", "letmein", "welcome", "123456"}
	for _, bad := range commonPasswords {
		if lower == bad || strings.HasPrefix(lower, bad+" ") || strings.HasSuffix(lower, " "+bad) || strings.Contains(lower, " "+bad+" ") || strings.HasPrefix(lower, bad) && len(lower) < len(bad)+3 {
			v.AddError("password", "password contains a common predictable pattern")
			valid = false
			break
		}
	}

	return valid
}

// calculateEntropy estimates password entropy
func calculateEntropy(password string) float64 {
	if len(password) == 0 {
		return 0
	}

	poolSize := 0
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		poolSize += 26
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		poolSize += 26
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		poolSize += 10
	}
	if regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		poolSize += 32
	}

	if poolSize == 0 {
		return 0
	}

	// Entropy = length * log2(poolSize)
	importEnt := 0.0
	if poolSize > 0 {
		importEnt = float64(len(password)) * math.Log2(float64(poolSize))
	}
	return importEnt
}

// ============================================
// INPUT SANITIZATION
// ============================================

// SanitizeInput removes potentially dangerous characters
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except common whitespace
	result := make([]rune, 0, len([]rune(input)))
	for _, r := range input {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' || (r >= 32 && r < 127) {
			result = append(result, r)
		}
	}

	return string(result)
}

// SanitizeSQLParameter sanitizes SQL parameter
func SanitizeSQLParameter(param string) string {
	// Remove SQL comment markers
	param = strings.ReplaceAll(param, "--", "")
	param = strings.ReplaceAll(param, "/*", "")
	param = strings.ReplaceAll(param, "*/", "")

	// Remove SQL operators that could be used for injection
	param = strings.ReplaceAll(param, ";", "")

	return param
}

// EscapeXSS escapes XSS payloads
func EscapeXSS(input string) string {
	replacer := strings.NewReplacer(
		"\u003c", "\u0026lt;",
		">", "\u0026gt;",
		"\"", "\u0026quot;",
		"'", "\u0026#39;",
		"\u0026", "\u0026amp;",
	)
	return replacer.Replace(input)
}

// ValidateNoCommandInjection checks for command injection patterns
func (v *Validator) ValidateNoCommandInjection(input string, field string) bool {
	dangerousPatterns := []string{
		";", "\u0026\u0026", "||", "|", "`", "$(", "${", "$",
		"\u003c(", "\u003e(", "\n", "\r", "\x00",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(input, pattern) {
			v.AddError(field, fmt.Sprintf("potentially dangerous character detected: %s", pattern))
			return false
		}
	}

	return true
}

// ============================================
// UUID VALIDATION
// ============================================

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidateUUID validates UUID format
func (v *Validator) ValidateUUID(id string, field string) bool {
	if id == "" {
		v.AddError(field, "UUID is required")
		return false
	}

	if !uuidRegex.MatchString(strings.ToLower(id)) {
		v.AddError(field, "invalid UUID format")
		return false
	}

	return true
}

// ============================================
// TAG VALIDATION
// ============================================

var tagRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,30}[a-zA-Z0-9]$`)

// ValidateTag validates host tag
func (v *Validator) ValidateTag(tag string) bool {
	if tag == "" {
		v.AddError("tag", "tag is required")
		return false
	}

	if len(tag) > 32 {
		v.AddError("tag", "tag must be less than 32 characters")
		return false
	}

	if !tagRegex.MatchString(tag) {
		v.AddError("tag", "invalid tag format (alphanumeric, hyphens, underscores)")
		return false
	}

	return true
}

// ValidateTags validates multiple tags
func (v *Validator) ValidateTags(tags []string) bool {
	if len(tags) > 20 {
		v.AddError("tags", "maximum 20 tags allowed")
		return false
	}

	seen := make(map[string]bool)
	for _, tag := range tags {
		if seen[tag] {
			v.AddError("tags", fmt.Sprintf("duplicate tag: %s", tag))
			return false
		}
		seen[tag] = true

		if !v.ValidateTag(tag) {
			return false
		}
	}

	return true
}
