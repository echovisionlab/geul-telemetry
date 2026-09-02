package telemetry

import "testing"

func TestParseLogRedactionPolicyRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("invalid redaction policy did not panic")
		}
	}()
	parseLogRedactionPolicy([]byte("{"))
}
