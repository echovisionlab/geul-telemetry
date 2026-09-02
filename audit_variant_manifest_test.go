package telemetry

import (
	"encoding/json"
	"os"
	"testing"
)

type auditVariantManifestEntry struct {
	Case         string         `json:"case"`
	Variant      string         `json:"variant"`
	Action       AuditAction    `json:"action"`
	TargetType   string         `json:"target_type"`
	TargetID     string         `json:"target_id"`
	ActorKind    ActorKind      `json:"actor_kind"`
	ActorService string         `json:"actor_service,omitempty"`
	Attributes   map[string]any `json:"attributes"`
}

func TestDomainAuditVariantManifest(t *testing.T) {
	contents, err := os.ReadFile("fixtures/domain-audit-wire-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var entries []auditVariantManifestEntry
	if err := json.Unmarshal(contents, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("variant manifest is empty")
	}
	for _, entry := range entries {
		t.Run(entry.Case, func(t *testing.T) {
			if entry.Variant == "" {
				t.Fatal("variant is required")
			}
			wire := map[string]any{
				"audit_id":    "00000000-0000-4000-8000-000000000001",
				"occurred_at": "2026-08-09T03:04:05Z",
				"action":      entry.Action,
				"target_type": entry.TargetType,
				"target_id":   entry.TargetID,
				"actor_kind":  entry.ActorKind,
			}
			if entry.ActorKind == ActorKindSystem {
				wire["actor_service"] = entry.ActorService
			} else if entry.ActorKind == ActorKindMember {
				wire["actor_member_id"] = "member-1"
			}
			for key, value := range entry.Attributes {
				wire[key] = value
			}
			encoded, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			var record AuditRecord
			if err := json.Unmarshal(encoded, &record); err != nil {
				t.Fatal(err)
			}
			if err := record.Validate(); err != nil {
				t.Fatal(err)
			}
			actualWire, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			var actual map[string]any
			if err := json.Unmarshal(actualWire, &actual); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{
				"audit_id", "occurred_at", "action", "target_type", "target_id",
				"request_id", "trace_id", "span_id",
				"actor_kind", "actor_member_id", "actor_service",
			} {
				delete(actual, key)
			}
			if !jsonValuesEqual(actual, entry.Attributes) {
				t.Fatalf("attributes = %#v, want %#v", actual, entry.Attributes)
			}
		})
	}
}
