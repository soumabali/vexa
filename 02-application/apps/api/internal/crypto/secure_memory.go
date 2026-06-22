package crypto

import (
	"crypto/subtle"
)

// SecureZero overwrites the contents of a slice in constant time to help
// clear cryptographic material from memory.
func SecureZero(b []byte) {
	if len(b) == 0 {
		return
	}
	subtle.ConstantTimeCopy(0, b, make([]byte, len(b)))
}

// SecureZeroString clears a string's underlying character data.
func SecureZeroString(s string) {
	SecureZero([]byte(s))
}

// SecureCompare compares two byte slices in constant time.
func SecureCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}
