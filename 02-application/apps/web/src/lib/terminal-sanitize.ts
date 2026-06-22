const MAX_BYTES_PER_FRAME = 10 * 1024;
const MAX_TOTAL_BYTES = 100 * 1024 * 1024;
const MAX_TITLE_LENGTH = 200;

/**
 * XSS / injection mitigations for terminal output:
 *
 * 1. Strip null bytes (terminal confusion)
 * 2. Block OSC 8 hyperlink sequences (OSC 8 ; params ; uri BEL/ST)
 *    — prevents link injection / open redirect via terminal output
 * 3. Allow-list safe OSC codes: 0 (title), 2 (icon name), 52 (clipboard read
 *    only — clipboard write is stripped), 4/10/11/12 (color queries)
 * 4. Strip all other OSC sequences not in the allow-list
 * 5. Sanitize window title separately (strip control chars + HTML meta chars)
 * 6. Rate-limit total bytes per session
 */

// OSC sequences: ESC ] <number> ; <data> (BEL | ESC \)
const OSC_PATTERN = /\x1b\]([\d;]+);(.*?)(?:\x07|\x1b\\)/g;

// OSC codes we allow to pass through
// 0 = set window title, 2 = set icon name
// 4/10/11/12 = color palette/foreground/background/cursor queries
// 52 without clipboard-write (r/p only)
const ALLOWED_OSC = new Set(['0', '2', '4', '10', '11', '12']);

// Strip OSC 8 (hyperlinks) and other unsafe OSC codes
function filterOsc(data: string): string {
  return data.replace(OSC_PATTERN, (match, code, payload) => {
    const primaryCode = code.split(';')[0];

    // OSC 8 = hyperlink — always strip (link injection vector)
    if (primaryCode === '8') return '';

    // OSC 52 = clipboard — strip write operations, allow read query
    // Format: OSC 52 ; <Pc> ; <data> — payload here is "<Pc>;<data>"
    if (primaryCode === '52') {
      // data field is after the last semicolon in payload
      const lastSemi = payload.lastIndexOf(';');
      const dataField = lastSemi >= 0 ? payload.slice(lastSemi + 1).trim() : payload.trim();
      if (dataField === '?') return match;
      // Anything else is clipboard write — strip
      return '';
    }

    // Allow-listed codes pass through unchanged
    if (ALLOWED_OSC.has(primaryCode)) return match;

    // All other OSC codes stripped
    return '';
  });
}

export class TerminalSanitizer {
  private totalBytes = 0;
  private exceeded = false;

  reset(): void {
    this.totalBytes = 0;
    this.exceeded = false;
  }

  sanitize(data: string): string {
    if (!data || this.exceeded) return '';

    if (data.length > MAX_BYTES_PER_FRAME) {
      data = data.slice(0, MAX_BYTES_PER_FRAME);
    }

    this.totalBytes += data.length;
    if (this.totalBytes > MAX_TOTAL_BYTES) {
      this.exceeded = true;
      return '\r\n\x1b[31m[Output limit exceeded]\x1b[0m';
    }

    // Strip null bytes
    data = data.replace(/\0/g, '');

    // Filter unsafe OSC sequences (XSS / link injection)
    data = filterOsc(data);

    return data;
  }

  sanitizeTitle(title: string): string {
    if (!title) return '';
    let result = title.slice(0, MAX_TITLE_LENGTH);
    // Strip C0/C1 control characters
    result = result.replace(/[\x00-\x1F\x7F-\x9F]/g, '');
    // Strip HTML meta chars (XSS in title bar rendering)
    result = result.replace(/[<>"'&]/g, '');
    return result;
  }
}
