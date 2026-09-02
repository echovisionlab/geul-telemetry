package telemetry

import (
	"encoding/json"
	"os"
	"testing"
)

func TestJobCatalogMatchesCrossLanguageFixture(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("fixtures/job-catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture []struct {
		JobKind        string   `json:"job_kind"`
		FailureReasons []string `json:"failure_reasons"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture) != len(canonicalJobFailureReasons) {
		t.Fatalf("fixture has %d of %d job kinds", len(fixture), len(canonicalJobFailureReasons))
	}
	for _, entry := range fixture {
		jobKind, parseErr := ParseJobKind(entry.JobKind)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if len(entry.FailureReasons) != len(canonicalJobFailureReasons[jobKind]) {
			t.Fatalf("fixture job %q has incomplete failure reasons", entry.JobKind)
		}
		for _, reason := range entry.FailureReasons {
			if _, reasonErr := ParseJobFailureReason(jobKind, reason); reasonErr != nil {
				t.Fatal(reasonErr)
			}
		}
	}
	if _, err := ParseJobKind("projection"); err == nil {
		t.Fatal("unknown job kind accepted")
	}
	if _, err := ParseJobFailureReason(JobKindMeshOptimization, "failed"); err == nil {
		t.Fatal("unknown job failure reason accepted")
	}
	if _, err := ParseJobFailureReason("projection", "failed"); err == nil {
		t.Fatal("unknown job kind accepted for failure reason")
	}
	if got := JobKindMeshOptimization.String(); got != "mesh_optimization" {
		t.Fatalf("String() = %q", got)
	}
}
