import { invoke as tauriInvoke } from "@tauri-apps/api/core";
import {
  listen as tauriListen,
  type EventCallback,
  type UnlistenFn,
} from "@tauri-apps/api/event";

export type DaemonApiRequest = {
  path: string;
  method?: string;
  body?: unknown;
};

export type DaemonApiResponse = {
  ok: boolean;
  status: number;
  data: unknown;
};

export type PlatformInfo = {
  platform: string;
  version: string;
  packaged: boolean;
};

export type DaemonExitPayload = {
  code: number | null;
  success: boolean;
};

type Invoke = <T>(command: string, args?: Record<string, unknown>) => Promise<T>;
type Listen = <T>(event: string, handler: EventCallback<T>) => Promise<UnlistenFn>;

export type DesktopInterop = {
  invoke: Invoke;
  listen: Listen;
};

export type DesktopBridge = ReturnType<typeof createDesktopBridge>;

export function createDesktopBridge(interop: DesktopInterop) {
  return {
    status: () => interop.invoke<Record<string, unknown>>("daemon_status"),
    api: (request: DaemonApiRequest) =>
      interop.invoke<DaemonApiResponse>("daemon_request", { request }),
    platform: () => interop.invoke<PlatformInfo>("platform_info"),
    openBrowser: (url: string) =>
      interop.invoke<boolean>("open_browser", { url }),
    openExternal: (url: string) =>
      interop.invoke<void>("open_external", { url }),
    onDaemonExit: (callback: (payload: DaemonExitPayload) => void) =>
      interop.listen<DaemonExitPayload>("racore://daemon-exit", (event) =>
        callback(event.payload),
      ),
  };
}

export const desktopBridge = createDesktopBridge({
  invoke: tauriInvoke,
  listen: tauriListen,
});

export function isDesktopApp(): boolean {
  return (
    typeof document !== "undefined" &&
    document.documentElement.dataset.desktop === "tauri"
  );
}

export function installLegacyDesktopBridge(): void {
  if (typeof window === "undefined" || window.racoreDesktop) return;
  window.racoreDesktop = {
    isDesktop: true,
    daemonUrl: "http://127.0.0.1:47831",
    ...desktopBridge,
    onDaemonExit(callback) {
      return desktopBridge.onDaemonExit(({ code }) => callback(code));
    },
  };
}
