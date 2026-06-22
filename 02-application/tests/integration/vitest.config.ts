import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    testTimeout: 60000,     // 60s for tests that need infrastructure
    hookTimeout: 120000,    // 120s for beforeAll/afterAll hooks
    reporters: ['verbose', 'json'],
    outputFile: {
      json: './test-report.json'
    },
    coverage: {
      reporter: ['text', 'json', 'html'],
      reportsDirectory: './coverage',
      include: [
        '../../apps/api/internal/**/*.go',
        '../../apps/desktop/src-frontend/**/*.ts',
        '../../apps/desktop/src-frontend/**/*.tsx'
      ]
    },
    // Sequence configuration
    sequence: {
      hooks: 'list'
    }
  },
  resolve: {
    alias: {
      '@': './'
    }
  }
});