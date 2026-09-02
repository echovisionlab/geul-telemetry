import { normalizeKey } from "./log-attribute-policy.ts";

export function stableErrorType(value: unknown): string {
  if (value instanceof Error) {
    const errorName = normalizeKey(value.name);
    if (errorName !== "" && errorName !== "error") return errorName;
    const constructorName = normalizeKey(value.constructor.name);
    if (constructorName !== "") return constructorName;
  }
  return "reported_error";
}
