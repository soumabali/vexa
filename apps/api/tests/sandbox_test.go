package tests

import "testing"

// TestSandbox is skipped because the sandbox package (IsLinux, DefaultConfig,
// NewManager, Manager methods, Profile, ResourceLimits) is not implemented in
// the current codebase.
func TestSandbox(t *testing.T) {
	t.Skip("sandbox package not implemented")
}
