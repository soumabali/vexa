import { execSync } from 'child_process';
import { mkdtempSync, writeFileSync, chmodSync } from 'fs';
import { tmpdir } from 'os';
import { join } from 'path';

/**
 * Test SSH Server Fixture using Docker
 * Spins up a lightweight SSH server container for integration testing
 */
export class TestSSHServer {
  private containerName: string;
  private hostKeyPath: string;
  private tempDir: string;
  private _port: number = 0;
  private _containerId: string = '';
  private _username: string = 'testuser';
  private _password: string = 'testpass123';
  private _host: string = 'localhost';

  constructor() {
    this.containerName = `ssh-test-server-${Date.now()}`;
    this.tempDir = mkdtempSync(join(tmpdir(), 'ssh-test-'));
    this.hostKeyPath = join(this.tempDir, 'ssh_host_rsa_key');
  }

  get port(): number { return this._port; }
  get host(): string { return this._host; }
  get username(): string { return this._username; }
  get password(): string { return this._password; }
  get containerId(): string { return this._containerId; }

  /**
   * Generate SSH host key and start container
   */
  async start(): Promise<void> {
    // Generate SSH host key
    execSync(`ssh-keygen -t rsa -b 2048 -f ${this.hostKeyPath} -N '' -C 'test-host-key'`, {
      stdio: 'pipe'
    });
    chmodSync(this.hostKeyPath, 0o600);

    // Create a custom entrypoint script that sets up user
    const entrypointScript = `#!/bin/sh
set -e

# Create test user
adduser -D -s /bin/sh ${this._username}
echo "${this._username}:${this._password}" | chpasswd

# Setup SSH directories
mkdir -p /home/${this._username}/.ssh
chmod 700 /home/${this._username}/.ssh

# Copy host key
cp ${this.hostKeyPath} /etc/ssh/ssh_host_rsa_key
chmod 600 /etc/ssh/ssh_host_rsa_key

# Configure SSH
sed -i 's/#PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config
sed -i 's/#ChallengeResponseAuthentication.*/ChallengeResponseAuthentication no/' /etc/ssh/sshd_config
sed -i 's/#UsePAM.*/UsePAM no/' /etc/ssh/sshd_config
sed -i 's/#HostKey \/etc\/ssh\/ssh_host_rsa_key/HostKey \/etc\/ssh\/ssh_host_rsa_key/' /etc/ssh/sshd_config
echo "AllowUsers ${this._username}" >> /etc/ssh/sshd_config

# Ensure proper ownership
chown -R ${this._username}:${this._username} /home/${this._username}

# Start sshd
exec /usr/sbin/sshd -D -e
`;

    const entrypointPath = join(this.tempDir, 'entrypoint.sh');
    writeFileSync(entrypointPath, entrypointScript, { mode: 0o755 });

    // Start container with alpine-based sshd
    const runCmd = `docker run -d \
      --name ${this.containerName} \
      --rm \
      -p 0:22 \
      -v ${this.tempDir}:/tmp/setup:ro \
      -v ${entrypointPath}:/entrypoint.sh:ro \
      --entrypoint /entrypoint.sh \
      alpine:3.19`;

    try {
      this._containerId = execSync(runCmd, { encoding: 'utf-8' }).trim();
    } catch (e: any) {
      // Fallback: use a pre-built image with sshd
      const fallbackCmd = `docker run -d \
        --name ${this.containerName} \
        --rm \
        -p 0:22 \
        -e SSH_USERS="${this._username}:${this._password}:1000:1000" \
        -e SSH_ENABLE_PASSWORD_AUTH=true \
        sickp/alpine-sshd:7.5-r2`;
      
      try {
        this._containerId = execSync(fallbackCmd, { encoding: 'utf-8' }).trim();
      } catch (e2: any) {
        // Final fallback: use linuxserver/openssh-server
        const finalCmd = `docker run -d \
          --name ${this.containerName} \
          --rm \
          -p 0:2222 \
          -e PUID=1000 \
          -e PGID=1000 \
          -e USER_NAME=${this._username} \
          -e USER_PASSWORD=${this._password} \
          -e PASSWORD_ACCESS=true \
          -e SUDO_ACCESS=false \
          linuxserver/openssh-server:latest`;
        
        this._containerId = execSync(finalCmd, { encoding: 'utf-8' }).trim();
        
        // Get the actual mapped port
        const portOutput = execSync(
          `docker port ${this.containerName} 2222`,
          { encoding: 'utf-8' }
        ).trim();
        this._port = parseInt(portOutput.split(':')[1], 10);
        this._host = portOutput.split(':')[0] || 'localhost';
        
        // Wait for SSH to be ready
        await this.waitForSSH(2222);
        return;
      }
    }

    // Get the actual mapped port
    const portOutput = execSync(
      `docker port ${this.containerName} 22`,
      { encoding: 'utf-8' }
    ).trim();
    this._port = parseInt(portOutput.split(':')[1], 10);
    this._host = portOutput.split(':')[0] || 'localhost';

    // Wait for SSH to be ready
    await this.waitForSSH(22);
  }

  /**
   * Wait for SSH server to be ready
   */
  private async waitForSSH(internalPort: number): Promise<void> {
    const maxAttempts = 30;
    const delayMs = 1000;

    for (let i = 0; i < maxAttempts; i++) {
      try {
        execSync(
          `docker exec ${this.containerName} sh -c "nc -z localhost ${internalPort}"`,
          { stdio: 'pipe' }
        );
        // Give it a moment more to fully initialize
        await new Promise(r => setTimeout(r, 500));
        return;
      } catch {
        await new Promise(r => setTimeout(r, delayMs));
      }
    }

    throw new Error(`SSH server failed to start after ${maxAttempts} attempts`);
  }

  /**
   * Check if SSH server is healthy
   */
  async healthCheck(): Promise<boolean> {
    try {
      execSync(
        `ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=5 -p ${this._port} ${this._username}@${this._host} echo "OK"`,
        { stdio: 'pipe' }
      );
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Execute a command on the SSH server
   */
  exec(command: string): string {
    return execSync(
      `ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p ${this._port} ${this._username}@${this._host} "${command}"`,
      { encoding: 'utf-8' }
    );
  }

  /**
   * Stop and remove the container
   */
  async stop(): Promise<void> {
    if (this._containerId) {
      try {
        execSync(`docker stop ${this.containerName} 2>/dev/null || true`, { stdio: 'pipe' });
        execSync(`docker rm ${this.containerName} 2>/dev/null || true`, { stdio: 'pipe' });
      } catch {
        // Ignore cleanup errors
      }
    }
  }
}
