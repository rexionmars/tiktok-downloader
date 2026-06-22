/// <reference types="vite/client" />

declare const __APP_VERSION__: string;

interface Window {
  go: Record<string, Record<string, Record<string, (...args: unknown[]) => Promise<unknown>>>>;
  runtime: {
    EventsOn: (event: string, cb: (...data: unknown[]) => void) => () => void;
    EventsOff: (event: string, ...additional: string[]) => void;
    EventsEmit: (event: string, ...data: unknown[]) => void;
    BrowserOpenURL: (url: string) => void;
    WindowMinimise: () => void;
    Quit: () => void;
  };
}
