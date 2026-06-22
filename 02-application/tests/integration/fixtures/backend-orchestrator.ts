/**
 * BackendOrchestrator
 *
 * Manages test infrastructure for integration tests:
 * - PostgreSQL + Redis via docker-compose.dev.yml
 * - Go API server (native or Docker)
 * - Database seeding / reset
 */

import { execSync, spawn } from 'child_process';
import { setTimeout } from 'timers/promises';

export interface BackendServices {
  apiUrl: string;
  wsUrl: string;
  postgresHost: string;
  postgresPort: number;
  redisHost: string;
  redisPort: number;
}

export class BackendOrchestrator {
  private composeFile: string;
  private projectName: string;
  private services: BackendServices | null = null;
  private apiProcess: ReturnType<typeof spawn> | null = null;

  constructor() {
    this.composeFile = './docker-compose.dev.yml';
    this.projectName = 'vexa-test';
  }

  /**
   * Start PostgreSQL and Redis containers
   */
  async startInfrastructure(): Promise<void> {
    console.log('Starting test infrastructure (PostgreSQL + Redis)...');

    try {
      execSync(
        `docker compose -f "${this.composeFile}" -p ${this.projectName} up -d postgres redis`,
        { stdio: 'inherit' }
      );
    } catch {
      // Containers might already be running; continue
    }

    // Wait for PostgreSQL using Docker healthcheck or port probe
    let pgReady = false;
    for (let i = 0; i < 30; i++) {
      try {
        // Use docker exec to check PostgreSQL readiness
        execSync(
          `docker exec ${this.projectName}-postgres-1 pg_isready -U vexa`,
          { stdio: 'ignore' }
        );
        pgReady = true;
        break;
      } catch {
        await setTimeout(1000);
      }
    }
    if (!pgReady) throw new Error('PostgreSQL failed to start');
    console.log('PostgreSQL is ready');

    // Wait for Redis
    let redisReady = false;
    for (let i = 0; i < 30; i++) {
      try {
        execSync(
          `docker exec ${this.projectName}-redis-1 redis-cli ping`,
          { stdio: 'ignore' }
        );
        redisReady = true;
        break;
      } catch {
        await setTimeout(1000);
      }
    }
    if (!redisReady) throw new Error('Redis failed to start');
    console.log('Redis is ready');

    this.services = {
      apiUrl: 'http://localhost:18080',
      wsUrl: 'ws://localhost:18080',
      postgresHost: 'localhost',
      postgresPort: 15432,
      redisHost: 'localhost',
      redisPort: 16379,
    };

    console.log('Test infrastructure ready');
  }

  /**
   * Stop and remove PostgreSQL + Redis containers
   */
  async stopInfrastructure(): Promise<void> {
    console.log('Stopping test infrastructure...');
    try {
      execSync(
        `docker compose -f "${this.composeFile}" -p ${this.projectName} down -v`,
        { stdio: 'inherit' }
      );
    } catch {
      // Ignore errors during cleanup
    }
  }

  /**
   * Start the API server in a Docker container using pre-built image.
   * Falls back to native mode if Go binary is available.
   */
  async startAPIServer(mode: 'docker' | 'native' = 'docker'): Promise<void> {
    if (mode === 'native') {
      try {
        execSync('which go', { stdio: 'ignore' });
      } catch {
        console.log('Go binary not found, falling back to Docker');
        mode = 'docker';
      }
    }

    if (mode === 'docker') {
      const composeFile = this.composeFile;
      execSync(`docker compose -f "${composeFile}" -p ${this.projectName} up -d`, {
        stdio: 'inherit',
      });

      let attempts = 0;
      const maxAttempts = 60;
      while (attempts < maxAttempts) {
        try {
          const response = await fetch('http://localhost:18080/health');
          if (response.status === 200) {
            console.log('API server is ready');
            return;
          }
        } catch {
          // not ready yet
        }
        await setTimeout(1000);
        attempts++;
      }
      throw new Error('API server failed to start within 60 seconds');
    } else {
      console.log('Starting API server natively...');
      try {
        execSync('go mod download', { cwd: process.cwd(), stdio: 'inherit' });
      } catch {
        console.log('go mod download failed, continuing...');
      }

      this.apiProcess = spawn('go', ['run', 'cmd/server/main.go'], {
        env: { ...process.env, PORT: '18080' },
        stdio: 'pipe',
        detached: true,
      });

      this.apiProcess.stdout?.on('data', (data) => {
        console.log('[API]', data.toString().trim());
      });

      this.apiProcess.stderr?.on('data', (data) => {
        console.error('[API]', data.toString().trim());
      });

      let attempts = 0;
      const maxAttempts = 30;
      while (attempts < maxAttempts) {
        try {
          const response = await fetch('http://localhost:18080/health');
          if (response.status === 200) {
            console.log('API server is ready');
            return;
          }
        } catch {
          // not ready yet
        }
        await setTimeout(1000);
        attempts++;
      }
      throw new Error('API server failed to start within 30 seconds');
    }
  }

  /**
   * Stop API server
   */
  async stopAPIServer(): Promise<void> {
    if (this.apiProcess) {
      process.kill(-this.apiProcess.pid!, 'SIGTERM');
      this.apiProcess = null;
    }
    // Also remove docker-compose API container if present
    try {
      execSync(`docker compose -f "${this.composeFile}" -p ${this.projectName} stop api`, {
        stdio: 'ignore',
      });
    } catch {
      // ignore
    }
  }

  /**
   * Seed minimal test data (user, host) and return JWT token
   */
  async seedTestData(): Promise<{ token: string; userId: string; hostId: string }> {
    // For integration testing we use a static test secret and generate JWT directly
    const secret = 'test-jwt-secret-minimum-32-bytes-long!!';

    // Use openssl to generate HMAC-SHA256 JWT
    const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
    const now = Math.floor(Date.now() / 1000);
    const payload = Buffer.from(JSON.stringify({
      user_id: '550e8400-e29b-41d4-a716-446655440000',
      email: 'test@example.com',
      role: 'user',
      mfa_enabled: false,
      mfa_verified: true,
      token_type: 'access',
      exp: now + 3600,
      iat: now,
      nbf: now,
      iss: 'vexa',
      sub: '550e8400-e29b-41d4-a716-446655440000',
      jti: `test-${Date.now()}`,
    })).toString('base64url');

    const signature = execSync(
      `echo -n "${header}.${payload}" | openssl dgst -sha256 -hmac "${secret}" -binary | base64 | tr '+/' '-_' | tr -d '='`,
      { encoding: 'utf-8' }
    ).trim();

    const token = `${header}.${payload}.${signature}`;

    return {
      token,
      userId: '550e8400-e29b-41d4-a716-446655440000',
      hostId: '550e8400-e29b-41d4-a716-446655440001',
    };
  }

  /**
   * Reset database (truncate tables)
   */
  async resetDatabase(): Promise<void> {
    // Placeholder: can be implemented with psql or migrations reset
    console.log('Resetting database...');
  }

  /**
   * Stop everything
   */
  async stopAll(): Promise<void> {
    await this.stopAPIServer();
    await this.stopInfrastructure();
    console.log('All test infrastructure stopped');
  }

  getServices(): BackendServices {
    if (!this.services) {
      throw new Error('Infrastructure not started');
    }
    return this.services;
  }
}
