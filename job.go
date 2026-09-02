package telemetry

import "fmt"

type JobKind string

type JobFailureReason string

const (
	JobKindMeshOptimization JobKind = "mesh_optimization"
	JobKindOGGeneration     JobKind = "og_generation"
)

const (
	JobFailureRejected           JobFailureReason = "rejected"
	JobFailureSourceNotFound     JobFailureReason = "source_not_found"
	JobFailureDownloadFailed     JobFailureReason = "download_failed"
	JobFailureOptimizationFailed JobFailureReason = "optimization_failed"
	JobFailureUploadFailed       JobFailureReason = "upload_failed"
	JobFailureInternal           JobFailureReason = "internal"
	JobFailureInvalidClaim       JobFailureReason = "invalid_claim"
	JobFailureSourceRejected     JobFailureReason = "source_rejected"
	JobFailureProcessingFailed   JobFailureReason = "processing_failed"
	JobFailureIntegrityFailed    JobFailureReason = "integrity_failed"
	JobFailureCompletionRejected JobFailureReason = "completion_rejected"
)

var canonicalJobFailureReasons = map[JobKind]map[JobFailureReason]struct{}{
	JobKindMeshOptimization: {
		JobFailureRejected: {}, JobFailureSourceNotFound: {}, JobFailureDownloadFailed: {},
		JobFailureOptimizationFailed: {}, JobFailureUploadFailed: {}, JobFailureInternal: {},
	},
	JobKindOGGeneration: {
		JobFailureInvalidClaim: {}, JobFailureSourceRejected: {}, JobFailureProcessingFailed: {},
		JobFailureIntegrityFailed: {}, JobFailureCompletionRejected: {},
	},
}

func ParseJobKind(value string) (JobKind, error) {
	jobKind := JobKind(value)
	if _, ok := canonicalJobFailureReasons[jobKind]; !ok {
		return "", fmt.Errorf("unknown canonical job kind %q", value)
	}
	return jobKind, nil
}

func ParseJobFailureReason(jobKind JobKind, value string) (JobFailureReason, error) {
	reasons, ok := canonicalJobFailureReasons[jobKind]
	if !ok {
		return "", fmt.Errorf("unknown canonical job kind %q", jobKind)
	}
	reason := JobFailureReason(value)
	if _, ok := reasons[reason]; !ok {
		return "", fmt.Errorf("unknown failure reason %q for job kind %q", value, jobKind)
	}
	return reason, nil
}

func (jobKind JobKind) String() string {
	return string(jobKind)
}
