import { describe, it, expect, beforeEach } from 'vitest';
import { TerminalSanitizer } from '../terminal-sanitize';

describe('TerminalSanitizer', () => {
  let sanitizer: TerminalSanitizer;

  beforeEach(() => {
    sanitizer = new TerminalSanitizer();
  });

  // ── Null byte stripping ────────────────────────────────────────────────────
  describe('null byte stripping', () => {
    it('strips null bytes from output', () => {
      expect(sanitizer.sanitize('hel\x00lo')).toBe('hello');
    });

    it('strips multiple null bytes', () => {
      expect(sanitizer.sanitize('\x00\x00\x00')).toBe('');
    });

    it('passes normal ASCII through unchanged', () => {
      const input = 'ls -la /home\r\n';
      expect(sanitizer.sanitize(input)).toBe(input);
    });
  });

  // ── OSC 8 hyperlink injection (XSS vector) ────────────────────────────────
  describe('OSC 8 hyperlink injection', () => {
    it('strips OSC 8 hyperlink sequence with BEL terminator', () => {
      const osc8 = '\x1b]8;;https://evil.com\x07click me\x1b]8;;\x07';
      const result = sanitizer.sanitize(osc8);
      expect(result).not.toContain('\x1b]8;');
      expect(result).not.toContain('evil.com');
    });

    it('strips OSC 8 hyperlink sequence with ST terminator (ESC \\)', () => {
      const osc8 = '\x1b]8;;https://evil.com\x1b\\click me\x1b]8;;\x1b\\';
      const result = sanitizer.sanitize(osc8);
      expect(result).not.toContain('\x1b]8;');
      expect(result).not.toContain('evil.com');
    });

    it('strips OSC 8 with params', () => {
      const osc8 = '\x1b]8;id=1;https://phishing.com\x07link\x1b]8;;\x07';
      const result = sanitizer.sanitize(osc8);
      expect(result).not.toContain('phishing.com');
    });
  });

  // ── Allowed OSC codes ──────────────────────────────────────────────────────
  describe('allowed OSC codes pass through', () => {
    it('allows OSC 0 (window title)', () => {
      const osc0 = '\x1b]0;My Terminal\x07';
      const result = sanitizer.sanitize(osc0);
      expect(result).toBe(osc0);
    });

    it('allows OSC 2 (icon name)', () => {
      const osc2 = '\x1b]2;icon\x07';
      expect(sanitizer.sanitize(osc2)).toBe(osc2);
    });

    it('allows OSC 4 (color)', () => {
      const osc4 = '\x1b]4;1;?\x07';
      expect(sanitizer.sanitize(osc4)).toBe(osc4);
    });
  });

  // ── Clipboard write stripping (OSC 52) ────────────────────────────────────
  describe('OSC 52 clipboard', () => {
    it('strips OSC 52 clipboard write', () => {
      const write = '\x1b]52;c;SGVsbG8gV29ybGQ=\x07'; // "Hello World" base64
      const result = sanitizer.sanitize(write);
      expect(result).not.toContain('SGVsbG8gV29ybGQ=');
    });

    it('allows OSC 52 clipboard read query', () => {
      const query = '\x1b]52;c;?\x07';
      const result = sanitizer.sanitize(query);
      expect(result).toBe(query);
    });
  });

  // ── Unknown OSC codes stripped ────────────────────────────────────────────
  describe('unknown OSC codes stripped', () => {
    it('strips arbitrary unknown OSC', () => {
      const unknown = '\x1b]9999;payload\x07';
      const result = sanitizer.sanitize(unknown);
      expect(result).not.toContain('\x1b]9999');
    });
  });

  // ── sanitizeTitle ──────────────────────────────────────────────────────────
  describe('sanitizeTitle()', () => {
    it('strips HTML meta chars', () => {
      expect(sanitizer.sanitizeTitle('<script>alert(1)</script>')).not.toContain('<');
      expect(sanitizer.sanitizeTitle('<script>alert(1)</script>')).not.toContain('>');
    });

    it('strips C0 control characters', () => {
      expect(sanitizer.sanitizeTitle('title\x1b[31m')).toBe('title[31m');
    });

    it('truncates title to MAX_TITLE_LENGTH (200)', () => {
      const long = 'A'.repeat(300);
      expect(sanitizer.sanitizeTitle(long).length).toBe(200);
    });

    it('strips quotes and ampersand', () => {
      const title = 'a"b\'c&d';
      expect(sanitizer.sanitizeTitle(title)).toBe('abcd');
    });

    it('returns empty string for empty input', () => {
      expect(sanitizer.sanitizeTitle('')).toBe('');
    });
  });

  // ── Rate limiting ──────────────────────────────────────────────────────────
  describe('rate limiting', () => {
    it('truncates frames larger than MAX_BYTES_PER_FRAME (10KB)', () => {
      const big = 'A'.repeat(15 * 1024);
      const result = sanitizer.sanitize(big);
      expect(result.length).toBeLessThanOrEqual(10 * 1024);
    });

    it('returns exceeded message after total bytes > 100MB', () => {
      // Exceed 100MB: 100MB / 10KB = 10240 chunks
      const chunk = 'A'.repeat(10 * 1024);
      let result = '';
      for (let i = 0; i < 11000; i++) {
        result = sanitizer.sanitize(chunk);
        if (result.includes('Output limit exceeded')) break;
      }
      expect(result).toContain('Output limit exceeded');
    });

    it('reset() clears exceeded state', () => {
      const chunk = 'A'.repeat(10 * 1024);
      for (let i = 0; i < 10001; i++) {
        const r = sanitizer.sanitize(chunk);
        if (r.includes('Output limit exceeded')) break;
      }
      sanitizer.reset();
      expect(sanitizer.sanitize('hello')).toBe('hello');
    });
  });
});
