export const RACORE_SUFFIXES = [".racore", ".rac", ".core", ".ra"] as const;

export function racoreDomainFromInput(value: string): string | null {
  const candidate = value.trim().toLowerCase();
  if (!candidate) return null;
  try {
    const url = new URL(
      /^[a-z][a-z\d+.-]*:\/\//i.test(candidate)
        ? candidate
        : `racore://${candidate}`,
    );
    const domain = url.hostname.replace(/\.$/, "");
    const supported = RACORE_SUFFIXES.some(
      (suffix) => domain.endsWith(suffix) && domain.length > suffix.length,
    );
    return supported && /^[a-z\d](?:[a-z\d.-]*[a-z\d])?$/.test(domain)
      ? domain
      : null;
  } catch {
    return null;
  }
}
