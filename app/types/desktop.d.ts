export {};

declare global {
  interface Window {
    racoreDesktop?: {
      isDesktop: boolean;
      daemonUrl: string;
      status(): Promise<Record<string, unknown>>;
      api(request: { path: string; method?: string; body?: unknown }): Promise<{ ok: boolean; status: number; data: unknown }>;
      platform(): Promise<{ platform: string; version: string; packaged: boolean }>;
      openBrowser(url: string): Promise<boolean>;
      openExternal(url: string): Promise<void>;
      onDaemonExit(callback: (code: number) => void): void;
    };
  }
}
