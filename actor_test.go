package telemetry

import (
	"encoding/json"
	"os"
	"testing"
)

type unsupportedActor struct{}

func (unsupportedActor) Kind() ActorKind  { return "unsupported" }
func (unsupportedActor) actor() ActorKind { return "unsupported" }

func TestActorForRecord(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		actor   Actor
		want    RecordActor
		wantErr bool
	}{
		{name: "nil", want: RecordActor{Kind: ActorKindAnonymous}},
		{name: "anonymous", actor: AnonymousActor{}, want: RecordActor{Kind: ActorKindAnonymous}},
		{name: "member", actor: MemberActor{MemberID: "member-1"}, want: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}},
		{name: "system", actor: SystemActor{ServiceName: ServiceBackend}, want: RecordActor{Kind: ActorKindSystem, Service: "geul-backend"}},
		{name: "member missing ID", actor: MemberActor{}, wantErr: true},
		{name: "system missing service", actor: SystemActor{}, wantErr: true},
		{name: "system unknown service", actor: SystemActor{ServiceName: "api"}, wantErr: true},
		{name: "unsupported", actor: unsupportedActor{}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ActorForRecord(test.actor)
			if (err != nil) != test.wantErr {
				t.Fatalf("ActorForRecord() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("ActorForRecord() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestActorMarkersMatchKinds(t *testing.T) {
	t.Parallel()
	actors := []Actor{AnonymousActor{}, MemberActor{}, SystemActor{}}
	for _, actor := range actors {
		if actor.actor() != actor.Kind() {
			t.Fatalf("actor marker %q does not match kind %q", actor.actor(), actor.Kind())
		}
	}
}

func TestRecordActorValidate(t *testing.T) {
	t.Parallel()
	valid := []RecordActor{
		{Kind: ActorKindAnonymous},
		{Kind: ActorKindMember, MemberID: "member-1"},
		{Kind: ActorKindSystem, Service: "geul-backend"},
	}
	for _, actor := range valid {
		if err := actor.Validate(); err != nil {
			t.Fatalf("Validate() unexpected error = %v", err)
		}
	}
	invalid := []RecordActor{
		{Kind: ActorKindAnonymous, MemberID: "member-1"},
		{Kind: ActorKindMember},
		{Kind: ActorKindMember, MemberID: "member-1", Service: "api"},
		{Kind: ActorKindSystem},
		{Kind: ActorKindSystem, Service: "api"},
		{Kind: ActorKindSystem, MemberID: "member-1", Service: "geul-backend"},
		{Kind: "other"},
	}
	for _, actor := range invalid {
		if err := actor.Validate(); err == nil {
			t.Fatalf("Validate() accepted %#v", actor)
		}
	}
}

func TestCanonicalServiceNamesMatchFixture(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("fixtures/service-identities.json")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		t.Fatal(err)
	}
	want := []ServiceName{
		ServiceBackend,
		ServiceWeb,
		ServiceIdentity,
		ServiceEditorCollab,
		ServiceCDN,
		ServiceOG,
		ServiceAssetOptimizer,
		ServiceTranscoder,
		ServiceWaveformProcessor,
	}
	if len(names) != len(want) {
		t.Fatalf("fixture service count = %d, want %d", len(names), len(want))
	}
	for index, serviceName := range want {
		if names[index] != string(serviceName) {
			t.Fatalf("fixture service[%d] = %q, want %q", index, names[index], serviceName)
		}
		if parsed, parseErr := ParseServiceName(names[index]); parseErr != nil || parsed != serviceName {
			t.Fatalf("ParseServiceName(%q) = %q, %v", names[index], parsed, parseErr)
		}
	}
	if _, err := ParseServiceName("geul-kratos"); err == nil {
		t.Fatal("vendor telemetry resource accepted as GEUL service identity")
	}
	if got := ServiceBackend.Instrumentation("http"); got != "geul-backend/http" {
		t.Fatalf("Instrumentation() = %q", got)
	}
}
