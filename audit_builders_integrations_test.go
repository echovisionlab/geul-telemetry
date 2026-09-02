package telemetry

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestTranslationSettingsFieldsMatchFixture(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("fixtures/translation-settings-fields.json")
	if err != nil {
		t.Fatal(err)
	}
	var fields []string
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	want := []string{"default_locale", "protected_terms"}
	if len(fields) != len(want) {
		t.Fatalf("fixture has %d fields, want %d", len(fields), len(want))
	}
	for index, field := range fields {
		if field != want[index] {
			t.Fatalf("fixture field %d = %q, want %q", index, field, want[index])
		}
		if _, err := NewTranslationSettingsUpdatedAuditRecord(
			AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}},
			[]string{field},
		); err != nil {
			t.Fatalf("field %q rejected: %v", field, err)
		}
	}
}

func TestIntegrationAuditBuilders(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) { return NewMailAdapterCreatedAuditRecord(m, "adapter-1") }, func() (AuditRecord, error) {
			return NewMailAdapterConfigUpdatedAuditRecord(m, "adapter-1", []string{"config"})
		}, func() (AuditRecord, error) { return NewMailAdapterDeletedAuditRecord(m, "adapter-1") },
		func() (AuditRecord, error) {
			return NewEmailSuppressionReleasedAuditRecord(m, "suppression-1")
		}, func() (AuditRecord, error) {
			return NewTranslationSettingsUpdatedAuditRecord(m, []string{"default_locale"})
		},
		func() (AuditRecord, error) { return NewTranslationProviderCreatedAuditRecord(m, "provider-1") }, func() (AuditRecord, error) {
			return NewTranslationProviderConfigUpdatedAuditRecord(m, "provider-1", []string{"type"})
		}, func() (AuditRecord, error) { return NewTranslationProviderDeletedAuditRecord(m, "provider-1") },
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}
	record, err := NewEmailSuppressionReleasedAuditRecord(m, "suppression-1")
	if err != nil {
		t.Fatal(err)
	}
	if record.Email != "" {
		t.Fatal("email suppression builder emitted email")
	}
}

func TestEmailSuppressionRejectsEveryOtherTypedAttribute(t *testing.T) {
	t.Parallel()
	valid := AuditRecord{
		ChangedFields: []string{"status"},
		PreviousState: AuditStateActive,
		NewState:      AuditStateReleased,
	}
	allowed := map[string]bool{
		"ChangedFields": true,
		"PreviousState": true,
		"NewState":      true,
	}
	typeOfRecord := reflect.TypeOf(valid)
	for fieldIndex := 0; fieldIndex < typeOfRecord.NumField(); fieldIndex++ {
		field := typeOfRecord.Field(fieldIndex)
		if allowed[field.Name] || isAuditEnvelopeField(field.Name) {
			continue
		}
		t.Run(field.Name, func(t *testing.T) {
			record := valid
			setAuditTestAttribute(t, reflect.ValueOf(&record).Elem().Field(fieldIndex))
			if err := validateEmailSuppressionUpdate(record); err == nil {
				t.Fatalf("email suppression accepted %s", field.Name)
			}
		})
	}
}

func isAuditEnvelopeField(name string) bool {
	switch name {
	case "AuditID", "OccurredAt", "Action", "Correlation", "RecordActor", "TargetType", "TargetID":
		return true
	default:
		return false
	}
}

func setAuditTestAttribute(t *testing.T, value reflect.Value) {
	t.Helper()
	switch value.Kind() {
	case reflect.String:
		value.SetString("extra")
	case reflect.Slice:
		slice := reflect.MakeSlice(value.Type(), 1, 1)
		slice.Index(0).SetString("extra")
		value.Set(slice)
	case reflect.Pointer:
		pointer := reflect.New(value.Type().Elem())
		switch element := pointer.Elem(); element.Kind() {
		case reflect.Int64:
			element.SetInt(1)
		case reflect.Slice:
			slice := reflect.MakeSlice(element.Type(), 1, 1)
			slice.Index(0).SetString("extra")
			element.Set(slice)
		case reflect.Struct:
			element.Set(reflect.ValueOf(testOccurredAt))
		default:
			t.Fatalf("unsupported audit attribute pointer type %s", value.Type())
		}
		value.Set(pointer)
	default:
		t.Fatalf("unsupported audit attribute type %s", value.Type())
	}
}

func TestTranslationSettingsRejectRetiredFields(t *testing.T) {
	t.Parallel()
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	for _, field := range []string{
		"debounce_seconds", "english_fallback_enabled", "machine_generated_public_serve",
		"stale_english_enabled", "stale_exact_enabled",
	} {
		if _, err := NewTranslationSettingsUpdatedAuditRecord(m, []string{field}); err == nil {
			t.Fatalf("retired field %q accepted", field)
		}
	}
}
