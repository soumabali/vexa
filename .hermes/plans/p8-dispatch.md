# P8 Dispatch — SSH Terminal E2E Real

**Tanggal:** 2026-06-28
**Branch:** `feat/p8-ssh-terminal-e2e-real`
**Dispatcher:** Ame (Hermes)

---

## 📜 Konteks

Audit P8:
- ✅ Mock SSH server fixture sudah ada: `tests/integration/fixtures/ssh-server.ts` (204 LOC, Docker-based)
- ✅ E2E test `tests/e2e/specs/terminal.spec.ts` ada (40 baris) — tapi cuma smoke test (render, status, tabs)
- ❌ Tidak ada test case real SSH execution (ngetik command + verify output)

**Real gap:** Tambah integration test yang benar-benar eksekusi SSH command via test server + verify output di terminal.

---

## 🎯 3 Task

### Task 1 — Integration Test Real SSH Execution

**File baru:** `tests/integration/ssh_real_execution_test.go`

Test cases:
1. **Connect to test SSH server** dengan TestSSHServer fixture
2. **Execute command** "echo 'hello-vexa'" via SSH protocol
3. **Verify output** di stream buffer: mengandung "hello-vexa"
4. **Test multiple commands**: `pwd`, `whoami`, `ls /tmp`
5. **Test command with stderr**: redirect stderr dan verify mixed output
6. **Test disconnect/reconnect** scenario

Pakai `golang.org/x/crypto/ssh` untuk client. Pattern:
```go
func TestSSHRealExecution(t *testing.T) {
    server := fixtures.NewTestSSHServer()
    require.NoError(t, server.Start())
    defer server.Stop()
    
    config := &ssh.ClientConfig{
        User: server.Username(),
        Auth: []ssh.AuthMethod{ssh.Password(server.Password())},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    
    client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Host(), server.Port()), config)
    require.NoError(t, err)
    defer client.Close()
    
    session, err := client.NewSession()
    require.NoError(t, err)
    defer session.Close()
    
    output, err := session.CombinedOutput("echo 'hello-vexa'")
    require.NoError(t, err)
    assert.Contains(t, string(output), "hello-vexa")
}
```

### Task 2 — Extend E2E Spec (Real Command in UI)

**File:** `tests/e2e/specs/terminal.spec.ts`

Tambah test case:
```typescript
test('user dapat mengetik command dan lihat output', async ({ page }) => {
  // Assume mock SSH backend running via docker-compose
  // Setup test host di fixtures
  await page.goto('/terminal');
  // Wait terminal ready
  await expect(page.locator('.xterm')).toBeVisible({ timeout: 15000 });
  // Type command
  await page.locator('.xterm').click();
  await page.keyboard.type('echo test-vexa\n');
  // Verify output appears
  await expect(page.locator('.xterm').getByText('test-vexa')).toBeVisible({ timeout: 10000 });
});
```

### Task 3 — Documentation

**File baru:** `tests/e2e/README.md` (jika belum ada)

Document how to run E2E + integration tests dengan mock SSH server.

---

## 🚨 Pitfalls

1. HANYA edit `02-application/`. No commit/push/.git.
2. JANGAN ubah TestSSHServer fixture (sudah 204 LOC real impl).
3. SSH key verification pakai `InsecureIgnoreHostKey` di test (acceptable untuk test).
4. Test harus skip gracefully jika Docker tidak tersedia (server.Start() return error).
5. Integration test terpisah dari unit test — bisa di-skip dengan `-short` flag.

## ✅ Verification

```bash
cd apps/api && go test -run TestSSHRealExecution -v ./tests/integration/
cd apps/api && go test -short ./tests/...
```

## 📦 Commit (2-3 commit)

```
test(integration): real SSH command execution tests via fixture
test(e2e): terminal command input + output verification
docs(test): E2E + integration test README
```

Print summary dengan test results.
