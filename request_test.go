package telemetry

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestRequestBuildersFixBoundaryAndEvent(t *testing.T) {
	t.Parallel()
	metadata := RequestMetadata{
		OccurredAt:  time.Date(2026, 8, 9, 3, 4, 5, 0, time.UTC),
		Correlation: Correlation{RequestID: "018f47a2-8a3d-4e17-9d42-6f12c89b1234"},
		RecordActor: RecordActor{Kind: ActorKindAnonymous},
	}
	result := RequestResult{StatusCode: 200, DurationMS: 4, Outcome: RequestOutcomeSucceeded}
	httpRecord, err := NewHTTPRequestRecord(metadata, "GET", "/members/{id}", result)
	if err != nil {
		t.Fatal(err)
	}
	if httpRecord.Event != "request.completed" || httpRecord.HTTPRoute != "/members/{id}" {
		t.Fatalf("unexpected HTTP record: %#v", httpRecord)
	}
	rpcRecord, err := NewRPCRequestRecord(metadata, "POST", "geul.v1.MemberService", "GetMember", result)
	if err != nil {
		t.Fatal(err)
	}
	if rpcRecord.RPCMethod != "GetMember" || rpcRecord.HTTPRoute != "" {
		t.Fatalf("unexpected RPC record: %#v", rpcRecord)
	}
	if _, err := NewHTTPRequestRecord(metadata, "", "", result); err == nil {
		t.Fatal("invalid HTTP boundary accepted")
	}
}

func TestHTTPResultClassifierMatchesCrossLanguageFixture(t *testing.T) {
	t.Parallel()
	type fixtureResult struct {
		StatusCode int            `json:"status_code"`
		Outcome    RequestOutcome `json:"outcome"`
		Reason     RequestReason  `json:"reason,omitempty"`
	}
	data, err := os.ReadFile("fixtures/http-request-results.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture []fixtureResult
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, expected := range fixture {
		result, err := ClassifyHTTPResult(expected.StatusCode, 7)
		if err != nil {
			t.Fatal(err)
		}
		got := fixtureResult{StatusCode: result.StatusCode, Outcome: result.Outcome, Reason: result.Reason}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("ClassifyHTTPResult(%d) = %#v, want %#v", expected.StatusCode, got, expected)
		}
	}
	for _, invalid := range []struct {
		statusCode int
		durationMS int64
	}{{99, 0}, {600, 0}, {200, -1}} {
		if _, err := ClassifyHTTPResult(invalid.statusCode, invalid.durationMS); err == nil {
			t.Fatalf("ClassifyHTTPResult(%d, %d) accepted", invalid.statusCode, invalid.durationMS)
		}
	}
}
