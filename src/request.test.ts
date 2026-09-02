import { readFile } from "node:fs/promises";

import { describe, expect, it } from "vitest";

import {
  buildHTTPRequestRecord,
  buildRPCRequestRecord,
  classifyHTTPResult,
  type RequestResult,
} from "./request.ts";

const resultFixturePath = new URL(
  "../fixtures/http-request-results.json",
  import.meta.url,
);

describe("request builders", () => {
  const metadata = {
    occurred_at: "2026-08-09T03:04:05Z",
    request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
    actor_kind: "anonymous",
  } as const;
  const result = {
    status_code: 200,
    duration_ms: 4,
    outcome: "succeeded",
  } as const;

  it("fixes the event and exact HTTP or RPC boundary", () => {
    expect(
      buildHTTPRequestRecord(metadata, "GET", "/members/{id}", result),
    ).toMatchObject({
      event: "request.completed",
      http_method: "GET",
      http_route: "/members/{id}",
    });
    expect(
      buildRPCRequestRecord(
        metadata,
        "POST",
        "geul.v1.MemberService",
        "GetMember",
        result,
      ),
    ).toMatchObject({
      event: "request.completed",
      rpc_service: "geul.v1.MemberService",
      rpc_method: "GetMember",
    });
    expect(() => buildHTTPRequestRecord(metadata, "", "", result)).toThrow();
  });

  it("matches the cross-language HTTP status classifier fixture", async () => {
    const fixture = JSON.parse(
      await readFile(resultFixturePath, "utf8"),
    ) as ReadonlyArray<Omit<RequestResult, "duration_ms">>;
    for (const expected of fixture) {
      expect(classifyHTTPResult(expected.status_code, 7)).toEqual({
        ...expected,
        duration_ms: 7,
      });
    }
    for (const [statusCode, durationMs] of [
      [99, 0],
      [600, 0],
      [200, -1],
      [200.5, 0],
      [200, 0.5],
    ]) {
      expect(() => classifyHTTPResult(statusCode, durationMs)).toThrow(
        TypeError,
      );
    }
  });
});
