declare module 'asciinema-player' {
  export function create(src: string | { url: string }, el: HTMLElement | null, opts?: Record<string, unknown>): {
    play: () => void;
    pause: () => void;
    dispose: () => void;
    getDuration: () => Promise<number>;
  };
}
