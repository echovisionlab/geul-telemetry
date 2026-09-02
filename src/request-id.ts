export const REQUEST_ID_HEADER = "X-Request-ID";
export const MESSAGE_REQUEST_ID_HEADER = "x-request-id";

const canonicalUUIDv4 =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

export function isRequestId(value: string): boolean {
  return canonicalUUIDv4.test(value);
}

export function createRequestId(): string {
  return globalThis.crypto.randomUUID();
}
