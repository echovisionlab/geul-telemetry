import { stableErrorType } from "./error-metadata.ts";
import { isForbiddenKey, normalizeKey } from "./log-attribute-policy.ts";

// Stable public entrypoint retained for package consumers. Policy ownership and
// error classification live in focused modules; only attribute normalization
// is implemented here.
export { stableErrorType } from "./error-metadata.ts";
export { isForbiddenKey, normalizeKey } from "./log-attribute-policy.ts";

export type NormalizedLogAttributes = Record<string, string | number | boolean>;

// Conservative log boundary shared by TypeScript services. Sensitive keys,
// raw errors, and untyped objects never reach stdout or OTLP exporters.
export function normalizeLogAttributes(
  attributes?: Readonly<Record<string, unknown>>,
): NormalizedLogAttributes {
  const normalized: NormalizedLogAttributes = {};
  if (attributes === undefined) return normalized;
  for (const [key, value] of Object.entries(attributes)) {
    const normalizedKey = normalizeKey(key);
    if (normalizedKey === "" || isForbiddenKey(normalizedKey)) continue;
    if (normalizedKey === "error" || normalizedKey === "err") {
      normalized.error_type = stableErrorType(value);
      continue;
    }
    if (
      typeof value === "string" ||
      typeof value === "number" ||
      typeof value === "boolean"
    ) {
      normalized[normalizedKey] = value;
    }
  }
  return normalized;
}
