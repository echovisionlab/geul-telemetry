package telemetry

import "fmt"

type ActorKind string

type ServiceName string

const (
	ServiceBackend           ServiceName = "geul-backend"
	ServiceWeb               ServiceName = "geul-web"
	ServiceIdentity          ServiceName = "geul-identity"
	ServiceEditorCollab      ServiceName = "geul-collab"
	ServiceCDN               ServiceName = "geul-cdn"
	ServiceOG                ServiceName = "geul-og"
	ServiceAssetOptimizer    ServiceName = "geul-asset-optimizer"
	ServiceTranscoder        ServiceName = "geul-transcoder"
	ServiceWaveformProcessor ServiceName = "geul-waveform-processor"
)

var canonicalServiceNames = map[ServiceName]struct{}{
	ServiceBackend:           {},
	ServiceWeb:               {},
	ServiceIdentity:          {},
	ServiceEditorCollab:      {},
	ServiceCDN:               {},
	ServiceOG:                {},
	ServiceAssetOptimizer:    {},
	ServiceTranscoder:        {},
	ServiceWaveformProcessor: {},
}

func ParseServiceName(value string) (ServiceName, error) {
	serviceName := ServiceName(value)
	if _, ok := canonicalServiceNames[serviceName]; !ok {
		return "", fmt.Errorf("unknown canonical service name %q", value)
	}
	return serviceName, nil
}

func (serviceName ServiceName) String() string {
	return string(serviceName)
}

func (serviceName ServiceName) Instrumentation(component string) string {
	return string(serviceName) + "/" + component
}

const (
	ActorKindAnonymous ActorKind = "anonymous"
	ActorKindMember    ActorKind = "member"
	ActorKindSystem    ActorKind = "system"
)

// Actor is the runtime subject associated with a request. Only the concrete
// actor types in this package can satisfy the interface.
type Actor interface {
	Kind() ActorKind
	actor() ActorKind
}

type AnonymousActor struct{}

func (AnonymousActor) Kind() ActorKind  { return ActorKindAnonymous }
func (AnonymousActor) actor() ActorKind { return ActorKindAnonymous }

type MemberActor struct {
	SessionID  string
	IdentityID string
	MemberID   string
}

func (MemberActor) Kind() ActorKind  { return ActorKindMember }
func (MemberActor) actor() ActorKind { return ActorKindMember }

type SystemActor struct {
	ServiceName ServiceName
}

func (SystemActor) Kind() ActorKind  { return ActorKindSystem }
func (SystemActor) actor() ActorKind { return ActorKindSystem }

// RecordActor is the safe actor projection written to logs and audit records.
// Session and identity identifiers deliberately have no wire fields here.
type RecordActor struct {
	Kind     ActorKind `json:"actor_kind"`
	MemberID string    `json:"actor_member_id,omitempty"`
	Service  string    `json:"actor_service,omitempty"`
}

func ActorForRecord(actor Actor) (RecordActor, error) {
	if actor == nil {
		actor = AnonymousActor{}
	}
	switch typed := actor.(type) {
	case AnonymousActor:
		return RecordActor{Kind: ActorKindAnonymous}, nil
	case MemberActor:
		if typed.MemberID == "" {
			return RecordActor{}, fmt.Errorf("member actor requires member ID")
		}
		return RecordActor{Kind: ActorKindMember, MemberID: typed.MemberID}, nil
	case SystemActor:
		serviceName, err := ParseServiceName(string(typed.ServiceName))
		if err != nil {
			return RecordActor{}, err
		}
		return RecordActor{Kind: ActorKindSystem, Service: string(serviceName)}, nil
	default:
		return RecordActor{}, fmt.Errorf("unsupported actor type %T", actor)
	}
}

func (actor RecordActor) Validate() error {
	switch actor.Kind {
	case ActorKindAnonymous:
		if actor.MemberID != "" || actor.Service != "" {
			return fmt.Errorf("anonymous actor cannot contain member or service identity")
		}
	case ActorKindMember:
		if actor.MemberID == "" || actor.Service != "" {
			return fmt.Errorf("member actor requires only actor_member_id")
		}
	case ActorKindSystem:
		if actor.MemberID != "" {
			return fmt.Errorf("system actor requires only actor_service")
		}
		if _, err := ParseServiceName(actor.Service); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid actor kind %q", actor.Kind)
	}
	return nil
}
