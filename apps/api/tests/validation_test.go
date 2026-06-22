package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/soumabali/vexa/internal/api/validators"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"ValidEmail", "user@example.com", false},
		{"ValidEmailWithPlus", "user+tag@example.com", false},
		{"ValidEmailWithSubdomain", "user@sub.example.com", false},
		{"EmptyEmail", "", true},
		{"NoAtSymbol", "userexample.com", true},
		{"NoDomain", "user@", true},
		{"NoLocalPart", "@example.com", true},
		{"InvalidChars", "user name@example.com", true},
		{"TooLong", "a" + string(make([]byte, 250)) + "@example.com", true},
		{"MultipleAt", "user@@example.com", true},
		{"ValidUppercase", "User@Example.COM", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidateEmail(tt.email)
			assert.Equal(t, tt.wantErr, !v.Valid(), "Email validation failed for %s", tt.email)
		})
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{"ValidIPv4", "192.168.1.1", false},
		{"ValidIPv4Local", "10.0.0.1", false},
		{"ValidIPv6", "::1", false},
		{"ValidIPv6Full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"EmptyIP", "", true},
		{"InvalidIP", "256.256.256.256", true},
		{"InvalidFormat", "not-an-ip", true},
		{"TooManyOctets", "192.168.1.1.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidateIP(tt.ip)
			assert.Equal(t, tt.wantErr, !v.Valid(), "IP validation failed for %s", tt.ip)
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"ValidSSH", 22, false},
		{"ValidHTTPS", 443, false},
		{"ValidMin", 1, false},
		{"ValidMax", 65535, false},
		{"InvalidZero", 0, true},
		{"InvalidNegative", -1, true},
		{"InvalidTooHigh", 65536, true},
		{"InvalidTooHigh2", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidatePort(tt.port)
			assert.Equal(t, tt.wantErr, !v.Valid(), "Port validation failed for %d", tt.port)
		})
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		wantErr  bool
	}{
		{"ValidFQDN", "server.example.com", false},
		{"ValidHostname", "localhost", false},
		{"ValidIP", "192.168.1.1", false},
		{"ValidWithDash", "my-server.example.com", false},
		{"ValidLong", "a.b.c.d.e.f.g.h.i.j.k.l.m.n.o.p.q.r.s.t.u.v.w.x.y.z.com", false},
		{"Empty", "", true},
		{"TooLong", "a" + string(make([]byte, 260)), true},
		{"InvalidChars", "server@example.com", true},
		{"DoubleDot", "server..example.com", true},
		{"LeadingDash", "-server.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidateHostname(tt.hostname)
			assert.Equal(t, tt.wantErr, !v.Valid(), "Hostname validation failed for %s", tt.hostname)
		})
	}
}

func TestValidateSSHKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"ValidRSA", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC... user@example", false},
		{"ValidEd25519", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... user@example", false},
		{"ValidECDSA", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABB... user@example", false},
		{"Empty", "", true},
		{"NoPrefix", "not-a-key", true},
		{"TooShort", "ssh-rsa ab", true},
		{"InvalidPrefix", "ssh-invalid AAAAB3... user", true},
		{"OnlyPrefix", "ssh-rsa", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidateSSHKey(tt.key)
			assert.Equal(t, tt.wantErr, !v.Valid(), "SSH key validation failed for %s", tt.name)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	policy := validators.DefaultPasswordPolicy()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"StrongPassword", "MyStr0ng!P@ssw0rd", false},
		{"ValidMinimum", "Pass1234!@", false},
		{"ValidLong", "ThisIsAVeryLongPassword123!", false},
		{"Empty", "", true},
		{"TooShort", "Short1!", true},
		{"NoUppercase", "password123!", true},
		{"NoLowercase", "PASSWORD123!", true},
		{"NoDigit", "Password!@#", true},
		{"NoSpecial", "Password123", true},
		{"TooLong", string(make([]byte, 200)), true},
		{"CommonPattern", "Password1!", true}, // Low entropy
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidatePassword(tt.password, policy)
			assert.Equal(t, tt.wantErr, !v.Valid(), "Password validation failed for %s", tt.name)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"NormalText", "Hello World", "Hello World"},
		{"WithNewlines", "Hello\nWorld", "Hello\nWorld"},
		{"WithTabs", "Hello\tWorld", "Hello\tWorld"},
		{"WithNull", "Hello\x00World", "HelloWorld"},
		{"WithControlChars", "Hello\x01World\x02", "HelloWorld"},
		{"Mixed", "Hello\n\x00\tWorld", "Hello\n\tWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validators.SanitizeInput(tt.input)
			assert.Equal(t, tt.want, got, "Sanitization failed for %s", tt.name)
		})
	}
}

func TestEscapeXSS(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"NoXSS", "Hello World", "Hello World"},
		{"ScriptTag", "\u003cscript\u003ealert(1)\u003c/script\u003e", "\u0026lt;script\u0026gt;alert(1)\u0026lt;/script\u0026gt;"},
		{"HTMLAttribute", "\u003cimg src=x onerror=alert(1)\u003e", "\u0026lt;img src=x onerror=alert(1)\u0026gt;"},
		{"Quote", `" onclick="alert(1)`, "\u0026quot; onclick=\u0026quot;alert(1)"},
		{"Ampersand", "a & b", "a \u0026amp; b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validators.EscapeXSS(tt.input)
			assert.Equal(t, tt.want, got, "XSS escape failed for %s", tt.name)
		})
	}
}

func TestValidateNoCommandInjection(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"NormalPath", "/path/to/file", false},
		{"WithSpaces", "/path to/file", false},
		{"WithSemicolon", "cmd; rm -rf /", true},
		{"WithPipe", "cmd | cat", true},
		{"WithBacktick", "cmd `whoami`", true},
		{"WithDollar", "cmd $(whoami)", true},
		{"WithNewline", "cmd\nwhoami", true},
		{"WithNull", "cmd\x00whoami", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidateNoCommandInjection(tt.input, "command")
			assert.Equal(t, tt.wantErr, !v.Valid(), "Command injection validation failed for %s", tt.name)
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"Valid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"ValidUpper", "550E8400-E29B-41D4-A716-446655440000", false},
		{"Empty", "", true},
		{"TooShort", "550e8400-e29b-41d4-a716", true},
		{"InvalidFormat", "not-a-uuid", true},
		{"MissingDashes", "550e8400e29b41d4a716446655440000", true},
		{"TooLong", "550e8400-e29b-41d4-a716-446655440000-extra", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidateUUID(tt.uuid, "id")
			assert.Equal(t, tt.wantErr, !v.Valid(), "UUID validation failed for %s", tt.name)
		})
	}
}

func TestValidateTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantErr bool
	}{
		{"ValidSimple", "production", false},
		{"ValidWithDash", "web-server", false},
		{"ValidWithUnderscore", "db_server", false},
		{"ValidNumeric", "server123", false},
		{"Empty", "", true},
		{"TooLong", "this-is-a-very-long-tag-name-that-exceeds", true},
		{"InvalidStart", "-server", true},
		{"InvalidEnd", "server-", true},
		{"InvalidChars", "server@prod", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validators.New()
			v.ValidateTag(tt.tag)
			assert.Equal(t, tt.wantErr, !v.Valid(), "Tag validation failed for %s", tt.name)
		})
	}
}
