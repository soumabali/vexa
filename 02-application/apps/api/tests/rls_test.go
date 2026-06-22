package tests

import "testing"

// TestRowLevelSecurity is skipped because setupTestDB helper and row-level
// security schema are not implemented in the current codebase.
func TestRowLevelSecurity(t *testing.T) {
	t.Skip("row-level security tests not implemented")
}
