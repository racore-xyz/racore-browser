const LOCAL_DAEMON = "http://127.0.0.1:47831";

export type ProviderInfo = {
  id: string;
  name: string;
  kind: string;
  default_model: string;
  free: boolean;
  local: boolean;
  connected: boolean;
  maskedKey?: string | null;
};

export async function daemonRequest<T>(
  path: string,
  options: { method?: string; body?: unknown } = {},
): Promise<T> {
  if (typeof window !== "undefined" && window.racoreDesktop?.api) {
    const result = await window.racoreDesktop.api({
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

export async function listProviders(): Promise<ProviderInfo[]> {
  return daemonRequest<ProviderInfo[]>("/v1/providers");
}

export async function connectProvider(provider: string, apiKey: string) {
  return daemonRequest(`/v1/providers/${provider}/connect`, {
    method: "PUT",
    body: { api_key: apiKey },
  });
}
