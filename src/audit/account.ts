import type {
  AccountSessionScope,
  AuditCollectionOperation,
  AuditRecord,
  AuditState,
} from "../records.ts";
import { canonicalAuditValues } from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import type { AuditMetadata } from "./types.ts";

function updated(
  metadata: AuditMetadata,
  memberId: string,
  attributes: Parameters<typeof buildAuditRecord>[2],
): AuditRecord {
  return buildAuditRecord(
    metadata,
    {
      action: "account.updated",
      target_type: "account",
      target_id: memberId,
    },
    attributes,
  );
}
export function buildAccountCanonicalEmailUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  previousEmail: string,
  newEmail: string,
): AuditRecord {
  return updated(metadata, memberId, {
    changed_fields: ["canonical_email"],
    previous_email: previousEmail,
    new_email: newEmail,
  });
}
function login(
  metadata: AuditMetadata,
  memberId: string,
  operation: AuditCollectionOperation,
  email: string,
): AuditRecord {
  return updated(metadata, memberId, {
    changed_fields: ["login_emails"],
    collection_operation: operation,
    email,
  });
}
export function buildAccountEmailLoginAddedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  email: string,
): AuditRecord {
  return login(metadata, memberId, "added", email);
}
export function buildAccountEmailLoginRemovedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  email: string,
): AuditRecord {
  return login(metadata, memberId, "removed", email);
}
function social(
  metadata: AuditMetadata,
  memberId: string,
  operation: AuditCollectionOperation,
  provider: string,
  providerSubject: string,
): AuditRecord {
  return updated(metadata, memberId, {
    changed_fields: ["social_logins"],
    collection_operation: operation,
    provider,
    provider_subject: providerSubject,
  });
}
export function buildAccountSocialLoginAddedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  provider: string,
  providerSubject: string,
): AuditRecord {
  return social(metadata, memberId, "added", provider, providerSubject);
}
export function buildAccountSocialLoginRemovedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  provider: string,
  providerSubject: string,
): AuditRecord {
  return social(metadata, memberId, "removed", provider, providerSubject);
}
function passkeys(
  metadata: AuditMetadata,
  memberId: string,
  operation: AuditCollectionOperation,
  passkeyIds: readonly string[],
): AuditRecord {
  return updated(metadata, memberId, {
    changed_fields: ["passkeys"],
    collection_operation: operation,
    passkey_ids: canonicalAuditValues(passkeyIds),
  });
}
export function buildAccountPasskeyAddedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  passkeyIds: readonly string[],
): AuditRecord {
  return passkeys(metadata, memberId, "added", passkeyIds);
}
export function buildAccountPasskeyRemovedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  passkeyIds: readonly string[],
): AuditRecord {
  return passkeys(metadata, memberId, "removed", passkeyIds);
}
export function buildAccountSessionRevokedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  sessionScope: AccountSessionScope,
  sessionIds: readonly string[],
): AuditRecord {
  return updated(metadata, memberId, {
    changed_fields: ["sessions"],
    collection_operation: "removed",
    session_scope: sessionScope,
    session_ids: canonicalAuditValues(sessionIds),
  });
}
export function buildAccountNewsletterSubscriptionUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  previousState: AuditState,
  newState: AuditState,
): AuditRecord {
  return updated(metadata, memberId, {
    changed_fields: ["newsletter_subscription"],
    previous_state: previousState,
    new_state: newState,
  });
}
function deletionState(
  metadata: AuditMetadata,
  memberId: string,
  previousState: AuditState,
  newState: AuditState,
): AuditRecord {
  return updated(metadata, memberId, {
    changed_fields: ["deletion_state"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildAccountDeletionRequestedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  previousState: AuditState,
): AuditRecord {
  return deletionState(
    metadata,
    memberId,
    previousState,
    "confirmation_pending",
  );
}
export function buildAccountDeletionScheduledAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  previousState: AuditState,
): AuditRecord {
  return deletionState(metadata, memberId, previousState, "scheduled");
}
export function buildAccountDeletionCancelledAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
): AuditRecord {
  return deletionState(metadata, memberId, "scheduled", "cancelled");
}
export function buildAccountDeletionRecoveredAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
): AuditRecord {
  return deletionState(
    metadata,
    memberId,
    "recovery_confirmation_pending",
    "recovered",
  );
}
export function buildAccountDeletedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
): AuditRecord {
  return buildAuditRecord(
    metadata,
    {
      action: "account.deleted",
      target_type: "account",
      target_id: memberId,
    },
    {},
  );
}
