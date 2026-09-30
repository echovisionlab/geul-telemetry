import type { AuditRecord } from "../records.ts";
import { requireCatalogTarget } from "./catalog.ts";
import {
  requireContentVersionActor,
  requireOnly,
  type ContentTarget,
} from "./content-shared.ts";
import { requirePostUpdate } from "./post.ts";
import { requirePageUpdate } from "./page.ts";
import { requireWorkUpdate } from "./work.ts";

export {
  buildPostCommentAuditRecord,
  buildPostConfigurationAuditRecord,
  buildPostCreatedAuditRecord,
  buildPostDeletedAuditRecord,
  buildPostFeaturedImageAuditRecord,
  buildPostLifecycleAuditRecord,
  buildPostParticipantAuditRecord,
  buildPostShareLinkAuditRecord,
  buildPostVersionCreatedAuditRecord,
  buildPostVersionRestoreAuditRecord,
} from "./post.ts";
export {
  buildPageConfigurationAuditRecord,
  buildPageCreatedAuditRecord,
  buildPageDeletedAuditRecord,
  buildPageFeaturedImageAuditRecord,
  buildPageLifecycleAuditRecord,
  buildPageShareLinkAuditRecord,
  buildPageVersionCreatedAuditRecord,
  buildPageVersionRestoreAuditRecord,
} from "./page.ts";
export {
  buildWorkCreatedAuditRecord,
  buildWorkCreditAuditRecord,
  buildWorkDeletedAuditRecord,
  buildWorkFeaturedImageAuditRecord,
  buildWorkLifecycleAuditRecord,
  buildWorkMetadataAuditRecord,
  buildWorkShareLinkAuditRecord,
  buildWorkVersionCreatedAuditRecord,
  buildWorkVersionRestoreAuditRecord,
} from "./work.ts";

/** Returns true only for Post, Page, and Work catalog actions. */
export function validateContentAuditRecord(record: AuditRecord): boolean {
  if (
    !(["post", "page", "work"] as const).includes(
      record.target_type as ContentTarget,
    )
  ) {
    return false;
  }
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireOnly(record, []);
    return true;
  }
  // Version persistence is called exclusively by the authenticated collab
  // boundary. All other Content mutations must retain their Member actor.
  requireContentVersionActor(record);
  switch (record.target_type) {
    case "post":
      requirePostUpdate(record);
      break;
    case "page":
      requirePageUpdate(record);
      break;
    case "work":
      requireWorkUpdate(record);
      break;
  }
  return true;
}
