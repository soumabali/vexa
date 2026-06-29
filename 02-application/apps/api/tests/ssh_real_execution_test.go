package tests

// Real SSH command execution tests against an ephemeral sshd container.
// Mirrors tests/integration/fixtures/ssh-server.ts (alpine-based sshd,
// password auth). Skipped when -short flag set or docker unavailable.
//
// Pattern: spin container -> dial via golang.org/x/crypto/ssh ->
// execute command -> verify output on stream.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// sshTestServer is a self-contained Go port of fixtures/ssh-server.ts.
// Each instance owns a unique container, temp dir, host key, and creds.
type sshTestServer struct {
	containerName string
	tempDir       string
	hostKeyPath   string
	username      string
	password      string
	host          string
	port          int
	containerID   string
}

func newSSHTestServer() *sshTestServer {
	suffix := make([]byte, 4)
	_, _ = rand.Read(suffix)
	return &sshTestServer{
		containerName: fmt.Sprintf("ssh-test-server-go-%d-%s", time.Now().UnixNano(), hex.EncodeToString(suffix)),
		username:      "testuser",
		password:      "testpass123",
		host:          "127.0.0.1",
	}
}

// dockerAvailable returns false on machines without docker CLI or daemon.
func dockerAvailable() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		return false
	}
	return true
}

func (s *sshTestServer) start(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipped in -short mode (requires Docker)")
	}
	if !dockerAvailable() {
		t.Skip("docker not available; skipping real SSH integration test")
	}

	dir, err := os.MkdirTemp("", "ssh-go-test-")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	s.tempDir = dir
	s.hostKeyPath = filepath.Join(dir, "ssh_host_rsa_key")

	// Generate host key on host so we control it.
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen not on PATH; skipping real SSH integration test")
	}
	genKey := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048",
		"-f", s.hostKeyPath, "-N", "", "-C", "test-host-key")
	if out, err := genKey.CombinedOutput(); err != nil {
		t.Skipf("ssh-keygen failed (skip): %v\n%s", err, out)
	}
	_ = os.Chmod(s.hostKeyPath, 0o600)

	// Build entrypoint script. alpine:3.19 base image has NO openssh,
	// so apk add before configuring.
	script := fmt.Sprintf(`#!/bin/sh
set -e
apk add --no-cache openssh
adduser -D -s /bin/sh %[1]s
echo "%[1]s:%[2]s" | chpasswd
mkdir -p /home/%[1]s/.ssh
chmod 700 /home/%[1]s/.ssh
cp /tmp/setup/ssh_host_rsa_key /etc/ssh/ssh_host_rsa_key
chmod 600 /etc/ssh/ssh_host_rsa_key
sed -i 's/#PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config
sed -i 's/#ChallengeResponseAuthentication.*/ChallengeResponseAuthentication no/' /etc/ssh/sshd_config
sed -i 's/#UsePAM.*/UsePAM no/' /etc/ssh/sshd_config
sed -i 's/#HostKey \/etc\/ssh\/ssh_host_rsa_key/HostKey \/etc\/ssh\/ssh_host_rsa_key/' /etc/ssh/sshd_config
echo "AllowUsers %[1]s" >> /etc/ssh/sshd_config
chown -R %[1]s:%[1]s /home/%[1]s
exec /usr/sbin/sshd -D -e
`, s.username, s.password)

	entryPath := filepath.Join(dir, "entrypoint.sh")
	if err := os.WriteFile(entryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write entrypoint: %v", err)
	}

	run := exec.Command("docker", "run", "-d",
		"--name", s.containerName,
		"--rm",
		"-p", "0:22",
		"-v", dir+":/tmp/setup:ro",
		"-v", entryPath+":/entrypoint.sh:ro",
		"--entrypoint", "/entrypoint.sh",
		"alpine:3.19",
	)
	out, err := run.CombinedOutput()
	if err == nil {
		s.containerID = strings.TrimSpace(string(out))
	} else {
		// Fallback: pre-built image with sshd (no host key sharing).
		fb := exec.Command("docker", "run", "-d",
			"--name", s.containerName,
			"--rm",
			"-p", "0:22",
			"-e", fmt.Sprintf("SSH_USERS=%s:%s:1000:1000", s.username, s.password),
			"-e", "SSH_ENABLE_PASSWORD_AUTH=true",
			"sickp/alpine-sshd:7.5-r2",
		)
		out, err = fb.CombinedOutput()
		if err != nil {
			_ = os.RemoveAll(dir)
			t.Skipf("docker run failed (skip): %v\n%s", err, out)
		}
		s.containerID = strings.TrimSpace(string(out))
	}

	portOut, err := exec.Command("docker", "port", s.containerName, "22").CombinedOutput()
	if err != nil {
		s.stop()
		t.Fatalf("docker port: %v\n%s", err, portOut)
	}
	// Format may include multiple bindings, one per line, e.g.:
	//   0.0.0.0:32768
	//   [::]:32768
	firstLine := strings.TrimSpace(strings.SplitN(string(portOut), "\n", 2)[0])
	parts := strings.Split(firstLine, ":")
	if len(parts) != 2 {
		s.stop()
		t.Fatalf("unexpected docker port output: %q", string(portOut))
	}
	fmt.Sscanf(parts[1], "%d", &s.port)

	// Wait for SSH TCP listener. Cold alpine + `apk add openssh` may
	// take 60+ s on first run.
	deadline := time.Now().Add(180 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.host, s.port), 500*time.Millisecond)
		if err == nil {
			_ = c.Close()
			// Give sshd a moment to actually handshake after install.
			time.Sleep(1500 * time.Millisecond)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.stop()
	t.Fatalf("SSH server never became reachable on %s:%d", s.host, s.port)
}

func (s *sshTestServer) stop() {
	if s.containerID == "" {
		_ = os.RemoveAll(s.tempDir)
		return
	}
	_ = exec.Command("docker", "stop", s.containerName).Run()
	_ = exec.Command("docker", "rm", s.containerName).Run()
	_ = os.RemoveAll(s.tempDir)
}

func (s *sshTestServer) client(t *testing.T) *ssh.Client {
	t.Helper()
	cfg := &ssh.ClientConfig{
		User:            s.username,
		Auth:            []ssh.AuthMethod{ssh.Password(s.password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // acceptable for tests
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	c, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		t.Fatalf("ssh dial %s: %v", addr, err)
	}
	return c
}

// runCmd executes cmd via a new session, returning combined output.
func runCmd(t *testing.T, c *ssh.Client, cmd string) (string, error) {
	t.Helper()
	sess, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	return string(out), err
}

// ---- TESTS ----

func TestSSHRealExecution(t *testing.T) {
	srv := newSSHTestServer()
	srv.start(t)
	defer srv.stop()

	cl := srv.client(t)
	defer cl.Close()

	out, err := runCmd(t, cl, "echo 'hello-vexa'")
	if err != nil {
		t.Fatalf("echo failed: %v", err)
	}
	if !strings.Contains(out, "hello-vexa") {
		t.Fatalf("expected output to contain %q, got %q", "hello-vexa", out)
	}
}

func TestSSHRealExecution_MultipleCommands(t *testing.T) {
	srv := newSSHTestServer()
	srv.start(t)
	defer srv.stop()

	cl := srv.client(t)
	defer cl.Close()

	cmds := map[string]string{
		"pwd":      "/home/testuser",
		"whoami":   "testuser",
		"ls /tmp":  "", // empty allowed; just verify no error
	}
	for cmd, wantSubstr := range cmds {
		out, err := runCmd(t, cl, cmd)
		if err != nil {
			t.Fatalf("cmd %q failed: %v", cmd, err)
		}
		if wantSubstr != "" && !strings.Contains(out, wantSubstr) {
			t.Fatalf("cmd %q: expected substring %q, got %q", cmd, wantSubstr, out)
		}
	}
}

func TestSSHRealExecution_StderrRedirect(t *testing.T) {
	srv := newSSHTestServer()
	srv.start(t)
	defer srv.stop()

	cl := srv.client(t)
	defer cl.Close()

	// Redirect stderr to stdout, then assert mixed output.
	out, err := runCmd(t, cl, "echo stdout-msg; echo stderr-msg 1>&2")
	if err != nil {
		t.Fatalf("redirect cmd failed: %v", err)
	}
	if !strings.Contains(out, "stdout-msg") || !strings.Contains(out, "stderr-msg") {
		t.Fatalf("expected both stdout and stderr in combined output, got %q", out)
	}
}

func TestSSHRealExecution_DisconnectReconnect(t *testing.T) {
	srv := newSSHTestServer()
	srv.start(t)
	defer srv.stop()

	// First session: run a command, then close.
	cl1 := srv.client(t)
	if _, err := runCmd(t, cl1, "echo first"); err != nil {
		t.Fatalf("first session cmd: %v", err)
	}
	cl1.Close()

	// Second session against same server should still work.
	cl2 := srv.client(t)
	defer cl2.Close()
	out, err := runCmd(t, cl2, "echo second")
	if err != nil {
		t.Fatalf("second session cmd: %v", err)
	}
	if !strings.Contains(out, "second") {
		t.Fatalf("expected %q in second session output, got %q", "second", out)
	}

	// Open many short-lived sessions to stress concurrent disconnects.
	for i := 0; i < 5; i++ {
		c := srv.client(t)
		out, err := runCmd(t, c, fmt.Sprintf("echo cycle-%d", i))
		_ = c.Close()
		if err != nil {
			t.Fatalf("cycle %d: %v", i, err)
		}
		if !strings.Contains(out, fmt.Sprintf("cycle-%d", i)) {
			t.Fatalf("cycle %d output mismatch: %q", i, out)
		}
	}
}

func TestSSHRealExecution_NonZeroExit(t *testing.T) {
	srv := newSSHTestServer()
	srv.start(t)
	defer srv.stop()

	cl := srv.client(t)
	defer cl.Close()

	sess, err := cl.NewSession()
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	defer sess.Close()
	err = sess.Run("false")
	if err == nil {
		t.Fatal("expected non-zero exit error from `false`")
	}
	if exitErr, ok := err.(*ssh.ExitError); !ok || exitErr.ExitStatus() != 1 {
		t.Fatalf("expected ExitStatus=1, got %v", err)
	}
}
