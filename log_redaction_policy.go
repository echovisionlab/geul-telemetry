package telemetry

import (
	_ "embed"
	"encoding/json"
)

//go:embed policy/log-redaction.json
var logRedactionPolicyJSON []byte

var (
	forbiddenLogKeys     map[string]struct{}
	forbiddenLogSuffixes []string
	forbiddenLogPrefixes []string
)

type logRedactionPolicy struct {
	Exact    []string `json:"exact"`
	Suffixes []string `json:"suffixes"`
	Prefixes []string `json:"prefixes"`
}

func parseLogRedactionPolicy(data []byte) logRedactionPolicy {
	var policy logRedactionPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		panic("telemetry: invalid embedded log redaction policy: " + err.Error())
	}
	return policy
}

func init() {
	policy := parseLogRedactionPolicy(logRedactionPolicyJSON)
	forbiddenLogKeys = make(map[string]struct{}, len(policy.Exact))
	for _, key := range policy.Exact {
		forbiddenLogKeys[key] = struct{}{}
	}
	forbiddenLogSuffixes = policy.Suffixes
	forbiddenLogPrefixes = policy.Prefixes
}
