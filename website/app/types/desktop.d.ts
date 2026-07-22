import type {
  DaemonApiRequest,
  DaemonApiResponse,
  PlatformInfo,
} from "../lib/desktop";

export {};

declare global {
  interface Window {
    /** Compatibility facade for UI code created before the typed Tauri adapter. */
    racoreDesktop?: {
      isDesktop: boolean;
      daemonUrl: string;
      status(): Promise<Record<string, unknown>>;
      api(request: DaemonApiRequest): Promise<DaemonApiResponse>;
      platform(): Promise<PlatformInfo>;
      openBrowser(url: string): Promise<boolean>;
      openExternal(url: string): Promise<void>;
      onDaemonExit(callback: (code: number | null) => void): Promise<() => void>;
    };
  }
}
