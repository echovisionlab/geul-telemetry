export const JOB_KIND_MESH_OPTIMIZATION = "mesh_optimization";
export const JOB_KIND_OG_GENERATION = "og_generation";

export const CANONICAL_JOB_KINDS = [
  JOB_KIND_MESH_OPTIMIZATION,
  JOB_KIND_OG_GENERATION,
] as const;

export type JobKind = (typeof CANONICAL_JOB_KINDS)[number];

export const JOB_FAILURE_REASONS = {
  [JOB_KIND_MESH_OPTIMIZATION]: [
    "rejected",
    "source_not_found",
    "download_failed",
    "optimization_failed",
    "upload_failed",
    "internal",
  ],
  [JOB_KIND_OG_GENERATION]: [
    "invalid_claim",
    "source_rejected",
    "processing_failed",
    "integrity_failed",
    "completion_rejected",
  ],
} as const satisfies Record<JobKind, readonly string[]>;

export type JobFailureReason = (typeof JOB_FAILURE_REASONS)[JobKind][number];

const canonicalJobKinds = new Set<string>(CANONICAL_JOB_KINDS);

export function parseJobKind(value: string): JobKind {
  if (!canonicalJobKinds.has(value)) {
    throw new TypeError(`unknown canonical job kind ${JSON.stringify(value)}`);
  }
  return value as JobKind;
}

export function parseJobFailureReason(
  jobKind: JobKind,
  value: string,
): JobFailureReason {
  const reasons = JOB_FAILURE_REASONS[jobKind] as readonly string[] | undefined;
  if (reasons === undefined) {
    throw new TypeError(
      `unknown canonical job kind ${JSON.stringify(jobKind)}`,
    );
  }
  if (!reasons.includes(value)) {
    throw new TypeError(
      `unknown failure reason ${JSON.stringify(value)} for job kind ${JSON.stringify(jobKind)}`,
    );
  }
  return value as JobFailureReason;
}
