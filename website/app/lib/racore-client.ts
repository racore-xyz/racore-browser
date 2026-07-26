import { desktopBridge, isDesktopApp } from "./desktop";

const LOCAL_DAEMON = "http://127.0.0.1:47831";

export type LocalAIStatus = {
  ready: boolean;
  state: "ready" | "offline" | "error";
  model: string;
  label: string;
  engine: string;
  parameters: number;
  quantization: string;
  localOnly: true;
  error?: string;
};

export async function daemonRequest<T>(
  path: string,
  options: { method?: string; body?: unknown } = {},
): Promise<T> {
  if (isDesktopApp()) {
    const result = await desktopBridge.api({
      path,
      method: options.method,
      body: options.body,
    });
    if (!result.ok)
      throw new Error(
        (result.data as { detail?: string })?.detail ||
          `Racore daemon returned ${result.status}`,
      );
    return result.data as T;
  }
  const response = await fetch(`${LOCAL_DAEMON}${path}`, {
    method: options.method || "GET",
    headers: options.body ? { "Content-Type": "application/json" } : undefined,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(
      error.detail || `Racore daemon returned ${response.status}`,
    );
  }
  return response.json();
}

export async function checkDaemon() {
  try {
    return await daemonRequest<Record<string, unknown>>("/health");
  } catch {
    return null;
  }
}

export async function localAIStatus(): Promise<LocalAIStatus> {
  return daemonRequest<LocalAIStatus>("/v1/local-ai/status");
}
