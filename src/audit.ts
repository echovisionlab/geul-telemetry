// Public compatibility facade. Domain modules remain internal implementation
// boundaries; package consumers import the supported root export.
export * from "./audit/account.ts";
export * from "./audit/attributes.ts";
export * from "./audit/artist-label.ts";
export * from "./audit/content.ts";
export * from "./audit/email-authoring.ts";
export * from "./audit/file.ts";
export * from "./audit/form.ts";
export * from "./audit/integrations.ts";
export * from "./audit/member.ts";
export * from "./audit/program.ts";
export * from "./audit/reference-data.ts";
export * from "./audit/reference-entities.ts";
export * from "./audit/relation-download-policy.ts";
export * from "./audit/release-campaign.ts";
export * from "./audit/settings.ts";
export * from "./audit/series-type.ts";
export * from "./audit/source-locale.ts";
export * from "./audit/locale-content.ts";
export * from "./audit/types.ts";
