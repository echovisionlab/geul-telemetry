export type Actor = AnonymousActor | MemberActor | SystemActor;

export const SERVICE_BACKEND = "geul-backend";
export const SERVICE_WEB = "geul-web";
export const SERVICE_IDENTITY = "geul-identity";
export const SERVICE_EDITOR_COLLAB = "geul-collab";
export const SERVICE_CDN = "geul-cdn";
export const SERVICE_OG = "geul-og";
export const SERVICE_ASSET_OPTIMIZER = "geul-asset-optimizer";
export const SERVICE_TRANSCODER = "geul-transcoder";
export const SERVICE_WAVEFORM_PROCESSOR = "geul-waveform-processor";

export const CANONICAL_SERVICE_NAMES = [
  SERVICE_BACKEND,
  SERVICE_WEB,
  SERVICE_IDENTITY,
  SERVICE_EDITOR_COLLAB,
  SERVICE_CDN,
  SERVICE_OG,
  SERVICE_ASSET_OPTIMIZER,
  SERVICE_TRANSCODER,
  SERVICE_WAVEFORM_PROCESSOR,
] as const;

export type ServiceName = (typeof CANONICAL_SERVICE_NAMES)[number];

const canonicalServiceNames = new Set<string>(CANONICAL_SERVICE_NAMES);

export function parseServiceName(value: string): ServiceName {
  if (!canonicalServiceNames.has(value)) {
    throw new TypeError(
      `unknown canonical service name ${JSON.stringify(value)}`,
    );
  }
  return value as ServiceName;
}

export function instrumentationName(
  serviceName: ServiceName,
  component: string,
): string {
  return `${serviceName}/${component}`;
}

export interface AnonymousActor {
  readonly kind: "anonymous";
}

export interface MemberActor {
  readonly kind: "member";
  readonly sessionId: string;
  readonly identityId: string;
  readonly memberId: string;
}

export interface SystemActor {
  readonly kind: "system";
  readonly serviceName: ServiceName;
}

export type RecordActor =
  | { readonly actor_kind: "anonymous" }
  | { readonly actor_kind: "member"; readonly actor_member_id: string }
  | { readonly actor_kind: "system"; readonly actor_service: string };

export function actorForRecord(actor: Actor): RecordActor {
  switch (actor.kind) {
    case "anonymous":
      return { actor_kind: "anonymous" };
    case "member":
      if (actor.memberId === "") {
        throw new TypeError("member actor requires memberId");
      }
      return { actor_kind: "member", actor_member_id: actor.memberId };
    case "system":
      return {
        actor_kind: "system",
        actor_service: parseServiceName(actor.serviceName),
      };
  }
}
