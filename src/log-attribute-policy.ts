import policy from "../policy/log-redaction.json" with { type: "json" };

const forbiddenKeys = new Set<string>(policy.exact);

export function normalizeKey(key: string): string {
  return key
    .trim()
    .replace(/([a-z\d])([A-Z])/g, "$1_$2")
    .replace(/[^\p{L}\p{N}]+/gu, "_")
    .replace(/^_+|_+$/g, "")
    .toLowerCase();
}

export function isForbiddenKey(key: string): boolean {
  const normalized = normalizeKey(key);
  return (
    forbiddenKeys.has(normalized) ||
    policy.suffixes.some((suffix) => normalized.endsWith(suffix)) ||
    policy.prefixes.some((prefix) => normalized.startsWith(prefix))
  );
}
