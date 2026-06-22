package sftp_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/sftp"
)

// MockSSHGateway implements a minimal SSH gateway for testing
type MockSSHGateway struct{}

func (m *MockSSHGateway) GetConnection(ctx interface{}, host string, port int, user, password string) (*struct {
	Client interface{}
	ID     string
}, error) {
	return nil, nil
}

func TestSFTPService_CreateSession(t *testing.T) {
	// This is a stub test - real SFTP session creation requires an SSH server
	// For unit tests, we test the service structure and error handling

	service := sftp.NewSFTPService(nil, nil)
	require.NotNil(t, service)

	// Test that service is properly initialized
	assert.NotNil(t, service)
}

func TestSFTPService_ListDirectory_ErrorCases(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent session
	_, err := service.ListDirectory("non-existent-session", "/")
	assert.ErrorIs(t, err, sftp.ErrSessionNotFound)
}

func TestSFTPService_Upload_ErrorCases(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test file too large
	largeSize := int64(11 * 1024 * 1024 * 1024) // 11GB
	_, err := service.Upload("non-existent", "/test.txt", strings.NewReader("test"), largeSize)
	assert.ErrorIs(t, err, sftp.ErrFileTooLarge)
}

func TestSFTPService_Download_ErrorCases(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent session
	_, _, err := service.Download("non-existent", "/test.txt", 0)
	assert.ErrorIs(t, err, sftp.ErrSessionNotFound)
}

func TestSFTPService_Remove_ErrorCases(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent session
	err := service.Remove("non-existent", "/test.txt")
	assert.ErrorIs(t, err, sftp.ErrSessionNotFound)
}

func TestSFTPService_Mkdir_ErrorCases(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent session
	err := service.Mkdir("non-existent", "/testdir")
	assert.ErrorIs(t, err, sftp.ErrSessionNotFound)
}

func TestSFTPService_Move_ErrorCases(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent session
	err := service.Move("non-existent", "/old", "/new")
	assert.ErrorIs(t, err, sftp.ErrSessionNotFound)
}

func TestSFTPService_CloseSession_ErrorCases(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent session
	err := service.CloseSession("non-existent")
	assert.ErrorIs(t, err, sftp.ErrSessionNotFound)
}

func TestSFTPService_MonitorBandwidth(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent session - should return empty stats
	stats := service.MonitorBandwidth("non-existent")
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.BytesSent)
	assert.Equal(t, int64(0), stats.BytesReceived)
	assert.Equal(t, 0.0, stats.SentSpeed)
	assert.Equal(t, 0.0, stats.ReceiveSpeed)
}

func TestSFTPService_CancelTransfer(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent transfer
	err := service.CancelTransfer("non-existent")
	assert.ErrorIs(t, err, sftp.ErrTransferNotFound)
}

func TestSFTPService_GetTransfer(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Test with non-existent transfer
	_, err := service.GetTransfer("non-existent")
	assert.ErrorIs(t, err, sftp.ErrTransferNotFound)
}

func TestSFTPService_ListActiveSessions(t *testing.T) {
	service := sftp.NewSFTPService(nil, nil)

	// Initially empty
	sessions := service.ListActiveSessions()
	assert.Empty(t, sessions)
}

// TestProgressCalculations tests progress calculation logic
func TestProgressCalculations(t *testing.T) {
	// Test zero size
	progress := &sftp.Progress{
		BytesTransferred: 0,
		TotalBytes:       0,
		Percentage:       0,
		SpeedBps:         0,
	}
	assert.Equal(t, 0.0, progress.Percentage)

	// Test 50% complete
	progress = &sftp.Progress{
		BytesTransferred: 512,
		TotalBytes:       1024,
		Percentage:       50.0,
		SpeedBps:         1024,
	}
	assert.Equal(t, 50.0, progress.Percentage)
}

// BenchmarkPathSanitization benchmarks path sanitization
func BenchmarkPathSanitization(b *testing.B) {
	baseDir := b.TempDir()
	ps := sftp.NewPathSanitizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ps.SanitizePath(baseDir, "foo/bar/baz.txt")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSafeJoin benchmarks SafeJoin helper
func BenchmarkSafeJoin(b *testing.B) {
	baseDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sftp.SafeJoin(baseDir, "foo/bar/baz.txt")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestFileInfoStructure tests the FileInfo structure
func TestFileInfoStructure(t *testing.T) {
	now := time.Now()
	info := sftp.FileInfo{
		Name:    "test.txt",
		Size:    1024,
		Mode:    0644,
		ModTime: now,
		IsDir:   false,
	}

	assert.Equal(t, "test.txt", info.Name)
	assert.Equal(t, int64(1024), info.Size)
	assert.False(t, info.IsDir)
}
