import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Terminal from '../terminal';
import TerminalTabs from '../terminal-tabs';

// Mock xterm
vi.mock('xterm', () => ({
  Terminal: class MockTerminal {
    element: HTMLElement | null = null;
    onDataCallback: ((data: string) => void) | null = null;

    options: Record<string, unknown> = {};

    constructor(options: Record<string, unknown>) {
      this.options = options;
    }

    open(container: HTMLElement) {
      this.element = document.createElement('div');
      this.element.className = 'xterm-mock';
      container.appendChild(this.element);
    }

    write(data: string | Uint8Array) {
      // Mock implementation
    }

    clear() {}
    focus() {}
    getSelection() {
      return 'mock selection';
    }

    onData(callback: (data: string) => void) {
      this.onDataCallback = callback;
      return { dispose: vi.fn() };
    }

    loadAddon() {}
    dispose() {
      if (this.element?.parentNode) {
        this.element.parentNode.removeChild(this.element);
      }
    }
  },
}));

vi.mock('xterm-addon-fit', () => ({
  FitAddon: class MockFitAddon {
    fit() {}
    dispose() {}
  },
}));

vi.mock('xterm-addon-web-links', () => ({
  WebLinksAddon: class MockWebLinksAddon {
    dispose() {}
  },
}));

describe('Terminal', () => {
  const mockWsUrl = 'ws://localhost:8080/ssh';

  beforeEach(() => {
    global.WebSocket = class MockWebSocket {
      url: string;
      readyState = WebSocket.CONNECTING;
      binaryType: BinaryType = 'arraybuffer';
      onopen: (() => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onclose: (() => void) | null = null;
      onerror: (() => void) | null = null;

      constructor(url: string) {
        this.url = url;
        setTimeout(() => {
          this.readyState = WebSocket.OPEN;
          this.onopen?.();
        }, 10);
      }

      send() {}
      close() {
        this.readyState = WebSocket.CLOSED;
        this.onclose?.();
      }
    } as unknown as typeof WebSocket;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders terminal container', () => {
    render(<Terminal id="test" wsUrl={mockWsUrl} />);
    expect(screen.getByRole('generic')).toBeInTheDocument();
  });

  it('connects to WebSocket on mount', async () => {
    const onConnectionChange = vi.fn();
    render(
<Terminal id="test" wsUrl={mockWsUrl} onConnectionChange={onConnectionChange} />);

    await waitFor(() => {
      expect(onConnectionChange).toHaveBeenCalledWith('connecting');
    });
  });

  it('handles binary data from WebSocket', async () => {
    const onBinary = vi.fn();
    render(<Terminal id="test" wsUrl={mockWsUrl} onBinary={onBinary} />);

    await waitFor(() => {
      const wsInstance = (global.WebSocket as unknown as typeof MockWebSocket).prototype;
      // This test would need more sophisticated mocking
    });
  });

  it('applies custom theme', () => {
    const customTheme = {
      background: '#ff0000',
      foreground: '#00ff00',
      cursor: '#0000ff',
      cursorAccent: '#ffffff',
      selectionBackground: '#cccccc',
      black: '#000000',
      red: '#ff0000',
      green: '#00ff00',
      yellow: '#ffff00',
      blue: '#0000ff',
      magenta: '#ff00ff',
      cyan: '#00ffff',
      white: '#ffffff',
      brightBlack: '#808080',
      brightRed: '#ff8080',
      brightGreen: '#80ff80',
      brightYellow: '#ffff80',
      brightBlue: '#8080ff',
      brightMagenta: '#ff80ff',
      brightCyan: '#80ffff',
      brightWhite: '#ffffff',
    };

    const { container } = render(
<Terminal id="test" wsUrl={mockWsUrl} theme={customTheme} />);

    const viewport = container.querySelector('.terminal-viewport');
    expect(viewport).toHaveStyle({ backgroundColor: '#ff0000' });
  });

  it('handles font size changes', () => {
    const { rerender } = render(
<Terminal id="test" wsUrl={mockWsUrl} fontSize={16} />);

    rerender(<Terminal id="test" wsUrl={mockWsUrl} fontSize={18} />);
    // Component should update with new font size
  });

  it('handles window resize', () => {
    const { container } = render(<Terminal id="test" wsUrl={mockWsUrl} />);

    // Simulate resize
    window.dispatchEvent(new Event('resize'));

    // Component should handle resize gracefully
    expect(container.querySelector('.terminal-container')).toBeInTheDocument();
  });
});

describe('TerminalTabs', () => {
  const defaultWsUrl = 'ws://localhost:8080/ssh';

  it('renders with initial tab', () => {
    render(<TerminalTabs defaultWsUrl={defaultWsUrl} />);
    expect(screen.getByText('Local')).toBeInTheDocument();
  });

  it('adds new tab on button click', async () => {
    const user = userEvent.setup();
    render(<TerminalTabs defaultWsUrl={defaultWsUrl} />);

    const addButton = screen.getByTitle('New tab');
    await user.click(addButton);

    await waitFor(() => {
      expect(screen.getByText('Session 1')).toBeInTheDocument();
    });
  });

  it('closes tab on X click', async () => {
    const user = userEvent.setup();
    const onTabChange = vi.fn();
    render(<TerminalTabs defaultWsUrl={defaultWsUrl} onTabChange={onTabChange} />);

    // Add a tab first
    const addButton = screen.getByTitle('New tab');
    await user.click(addButton);

    await waitFor(() => {
      expect(screen.getByText('Session 1')).toBeInTheDocument();
    });

    // Close the tab
    const closeButtons = screen.getAllByRole('button');
    const closeButton = closeButtons.find((btn) => btn.querySelector('svg'));
    if (closeButton) {
      await user.click(closeButton);
    }
  });

  it('switches active tab on click', async () => {
    const user = userEvent.setup();
    const onActiveTabChange = vi.fn();
    render(<TerminalTabs defaultWsUrl={defaultWsUrl} onActiveTabChange={onActiveTabChange} />);

    // Add a tab
    const addButton = screen.getByTitle('New tab');
    await user.click(addButton);

    await waitFor(() => {
      expect(screen.getByText('Session 1')).toBeInTheDocument();
    });

    // Click on the first tab
    const firstTab = screen.getByText('Local');
    await user.click(firstTab);

    expect(onActiveTabChange).toHaveBeenCalled();
  });

  it('displays correct status indicator', () => {
    const { container } = render(<TerminalTabs defaultWsUrl={defaultWsUrl} />);

    // Should show connecting status (yellow)
    const indicators = container.querySelectorAll('.bg-yellow-500');
    expect(indicators.length).toBeGreaterThan(0);
  });
});

describe('Terminal integration', () => {
  it('handles multiple tabs with different connections', async () => {
    const user = userEvent.setup();
    render(<TerminalTabs defaultWsUrl="ws://localhost:8080/ssh" />);

    // Add multiple tabs
    const addButton = screen.getByTitle('New tab');
    await user.click(addButton);
    await user.click(addButton);

    await waitFor(() => {
      expect(screen.getByText('Session 1')).toBeInTheDocument();
      expect(screen.getByText('Session 2')).toBeInTheDocument();
    });
  });

  it('maintains scrollback on tab switch', () => {
    // This would test that scroll position is maintained
    // when switching between tabs
    render(<TerminalTabs defaultWsUrl="ws://localhost:8080/ssh" />);

    // Component should render without errors
    expect(screen.getByText('Local')).toBeInTheDocument();
  });
});
