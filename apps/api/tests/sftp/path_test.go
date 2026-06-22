package sftp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/soumabali/vexa/internal/sftp"
)

func TestSanitizePath_ValidPaths(t *testing.T) {
	baseDir := t.TempDir()
	ps := sftp.NewPathSanitizer()

	tests := []struct {
		name     string
		userPath string
		wantRel  string
	}{
		{"simple file", "file.txt", "file.txt"},
		{"nested path", "dir/file.txt", filepath.Join("dir", "file.txt")},
		{"deep nested", "a/b/c/d.txt", filepath.Join("a", "b", "c", "d.txt")},
		{"single dot", "./file.txt", "file.txt"},
		{"trailing slash", "dir/", filepath.Join("dir", "")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ps.SanitizePath(baseDir, tt.userPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := filepath.Join(baseDir, tt.wantRel)
			if got != want {
				t.Errorf("SanitizePath() = %v, want %v", got, want)
			}
		})
	}
}

func TestSanitizePath_PathTraversal(t *testing.T) {
	baseDir := t.TempDir()
	ps := sftp.NewPathSanitizer()

	tests := []struct {
		name     string
		userPath string
	}{
		{"dotdot prefix", "../etc/passwd"},
		{"dotdot in middle", "foo/../../etc/passwd"},
		{"dotdot suffix", "foo/.."},
		{"absolute path", "/etc/passwd"},
		{"null byte", "foo\x00bar"},
		{"double dotdot", "../../.."},
		{"dotdot with valid prefix", "valid/../../etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ps.SanitizePath(baseDir, tt.userPath)
			if err == nil {
				t.Errorf("SanitizePath() expected error for path traversal, got nil")
			}
		})
	}
}

func TestSanitizePath_Symlinks(t *testing.T) {
	baseDir := t.TempDir()
	ps := sftp.NewPathSanitizer()

	// Create a file inside base
	insideFile := filepath.Join(baseDir, "inside.txt")
	if err := os.WriteFile(insideFile, []byte("inside"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file outside base
	outsideFile := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create symlink inside base pointing to inside file (allowed)
	goodLink := filepath.Join(baseDir, "good_link")
	if err := os.Symlink(insideFile, goodLink); err != nil {
		t.Fatal(err)
	}

	// Create symlink inside base pointing to outside file (blocked)
	badLink := filepath.Join(baseDir, "bad_link")
	if err := os.Symlink(outsideFile, badLink); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		userPath  string
		wantError bool
	}{
		{"good symlink", "good_link", false},
		{"bad symlink", "bad_link", true},
		{"path through bad symlink", "bad_link/anything", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ps.SanitizePath(baseDir, tt.userPath)
			if tt.wantError && err == nil {
				t.Errorf("SanitizePath() expected error for symlink escape, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("SanitizePath() unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizePath_SymlinkLoop(t *testing.T) {
	baseDir := t.TempDir()
	ps := sftp.NewPathSanitizer()

	// Create symlink loop
	linkA := filepath.Join(baseDir, "loop_a")
	linkB := filepath.Join(baseDir, "loop_b")
	if err := os.Symlink(linkB, linkA); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(linkA, linkB); err != nil {
		t.Fatal(err)
	}

	_, err := ps.SanitizePath(baseDir, "loop_a")
	if err == nil {
		t.Error("SanitizePath() expected error for symlink loop, got nil")
	}
}

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name    string
		fileName string
		wantErr bool
	}{
		{"valid", "file.txt", false},
		{"empty", "", true},
		{"dot", ".", true},
		{"dotdot", "..", true},
		{"slash", "dir/file", true},
		{"backslash", "dir\\file", true},
		{"null byte", "foo\x00bar", true},
		{"control char", "foo\x01bar", true},
		{"tilde", "~", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sftp.ValidateFileName(tt.fileName)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateFileName(%q) expected error, got nil", tt.fileName)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateFileName(%q) unexpected error: %v", tt.fileName, err)
			}
		})
	}
}

func TestSafeJoin(t *testing.T) {
	baseDir := t.TempDir()

	got, err := sftp.SafeJoin(baseDir, "foo/bar.txt")
	if err != nil {
		t.Fatalf("SafeJoin() unexpected error: %v", err)
	}

	want := filepath.Join(baseDir, "foo", "bar.txt")
	if got != want {
		t.Errorf("SafeJoin() = %v, want %v", got, want)
	}

	_, err = sftp.SafeJoin(baseDir, "../outside")
	if err == nil {
		t.Error("SafeJoin() expected error for traversal, got nil")
	}
}
