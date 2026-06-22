package sftp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrPathTraversal    = errors.New("path traversal attempt detected")
	ErrPathNotAllowed = errors.New("path not allowed")
	ErrSymlinkEscape   = errors.New("symlink points outside chroot")
	ErrPathTooLong     = errors.New("path exceeds maximum length")
	ErrInvalidPath     = errors.New("invalid path")
)

const (
	MaxPathLength = 4096
	MaxSymlinkDepth = 32
)

type PathSanitizer struct {
	mu sync.RWMutex
}

func NewPathSanitizer() *PathSanitizer {
	return &PathSanitizer{}
}

// SanitizePath sanitizes a user-provided path against a base directory (chroot).
// It prevents path traversal, resolves symlinks, and ensures the final path stays within the chroot.
func (ps *PathSanitizer) SanitizePath(basePath, userPath string) (string, error) {
	if len(userPath) > MaxPathLength {
		return "", ErrPathTooLong
	}

	// Clean and resolve the base path to an absolute path
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}

	// Ensure base path exists and is a directory
	info, err := os.Stat(absBase)
	if err != nil {
		return "", fmt.Errorf("base path error: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("base path is not a directory")
	}

	// Clean the user path (removes ., .., duplicate slashes)
	cleanUserPath := filepath.Clean(userPath)

	// Reject absolute paths
	if filepath.IsAbs(userPath) {
		return "", ErrPathTraversal
	}

	// Reject paths that try to traverse upward
	if strings.Contains(userPath, "..") {
		return "", ErrPathTraversal
	}

	// Join base and user path
	candidate := filepath.Join(absBase, cleanUserPath)

	// Ensure the resolved path is within the base directory
	resolvedBase := filepath.Clean(absBase)
	if !strings.HasPrefix(filepath.Clean(candidate), resolvedBase) {
		return "", ErrPathTraversal
	}

	// Resolve symlinks with depth limit
	resolved, err := ps.resolveSymlinks(candidate, absBase, 0)
	if err != nil {
		return "", err
	}

	// Final check: ensure resolved path is within base path
	rel, err := filepath.Rel(absBase, resolved)
	if err != nil {
		return "", fmt.Errorf("path resolution error: %w", err)
	}

	// Check for path traversal after resolution
	if strings.HasPrefix(rel, "..") || rel == ".." {
		return "", ErrPathTraversal
	}

	// Ensure the path is actually under the base directory
	if !strings.HasPrefix(resolved, absBase) {
		return "", ErrPathTraversal
	}

	return resolved, nil
}

// resolveSymlinks resolves symlinks in a path while ensuring they don't escape the chroot.
// It has a maximum depth to prevent symlink loops.
func (ps *PathSanitizer) resolveSymlinks(path, basePath string, depth int) (string, error) {
	if depth > MaxSymlinkDepth {
		return "", errors.New("symlink depth exceeded")
	}

	// Check if the current path is a symlink
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist - check parent directories for symlinks
			return ps.resolveParentSymlinks(path, basePath, depth)
		}
		return "", err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		// Not a symlink, return as-is
		return path, nil
	}

	// It's a symlink, resolve it
	target, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink: %w", err)
	}

	// If target is relative, resolve against the symlink's directory
	if !filepath.IsAbs(target) {
		symlinkDir := filepath.Dir(path)
		target = filepath.Join(symlinkDir, target)
	}

	// Check if the symlink target escapes the chroot
	rel, err := filepath.Rel(basePath, target)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rel, "..") || rel == ".." {
		return "", ErrSymlinkEscape
	}

	// Recursively resolve the target
	return ps.resolveSymlinks(target, basePath, depth+1)
}

// resolveParentSymlinks checks parent directories for symlinks when a path doesn't exist yet
func (ps *PathSanitizer) resolveParentSymlinks(path, basePath string, depth int) (string, error) {
	dir := filepath.Dir(path)
	if dir == path || dir == basePath {
		return path, nil
	}

	resolvedDir, err := ps.resolveSymlinks(dir, basePath, depth)
	if err != nil {
		return "", err
	}

	base := filepath.Base(path)
	return filepath.Join(resolvedDir, base), nil
}

// ValidateFileName checks if a filename is safe (no path separators, no control characters)
func ValidateFileName(name string) error {
	if name == "" {
		return ErrInvalidPath
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return ErrPathTraversal
	}

	if strings.Contains(name, "..") {
		return ErrPathTraversal
	}

	// Reject control characters
	for _, r := range name {
		if r < 32 || r == 127 {
			return ErrInvalidPath
		}
	}

	// Reject special names
	switch name {
	case ".", "..", "", "~", "/":
		return ErrInvalidPath
	}

	return nil
}

// SafeJoin safely joins a base path with a relative path, preventing traversal
func SafeJoin(base, rel string) (string, error) {
	ps := NewPathSanitizer()
	return ps.SanitizePath(base, rel)
}
