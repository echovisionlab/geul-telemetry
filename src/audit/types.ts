import type { RecordActor } from "../actor.ts";
import type { Correlation } from "../records.ts";

/** Metadata shared by every domain-audit record. */
export type AuditMetadata = Correlation &
  RecordActor & {
    readonly audit_id: string;
    readonly occurred_at: string;
  };
