package telemetry

import (
	"fmt"
	"time"
)

type RequestReason string

const (
	RequestReasonAuthenticationRequired RequestReason = "authentication_required"
	RequestReasonPermissionDenied       RequestReason = "permission_denied"
	RequestReasonRateLimited            RequestReason = "rate_limited"
	RequestReasonClientError            RequestReason = "client_error"
	RequestReasonServerError            RequestReason = "server_error"
)

type RequestMetadata struct {
	OccurredAt time.Time
	Correlation
	RecordActor
}

type RequestResult struct {
	StatusCode int
	DurationMS int64
	Outcome    RequestOutcome
	ErrorCode  string
	Reason     RequestReason
}

func ClassifyHTTPResult(statusCode int, durationMS int64) (RequestResult, error) {
	if statusCode < 100 || statusCode > 599 || durationMS < 0 {
		return RequestResult{}, fmt.Errorf("HTTP result requires a valid status_code and non-negative duration_ms")
	}
	result := RequestResult{StatusCode: statusCode, DurationMS: durationMS}
	switch {
	case statusCode < 400:
		result.Outcome = RequestOutcomeSucceeded
	case statusCode == 401:
		result.Outcome, result.Reason = RequestOutcomeBlocked, RequestReasonAuthenticationRequired
	case statusCode == 403:
		result.Outcome, result.Reason = RequestOutcomeBlocked, RequestReasonPermissionDenied
	case statusCode == 429:
		result.Outcome, result.Reason = RequestOutcomeBlocked, RequestReasonRateLimited
	case statusCode < 500:
		result.Outcome, result.Reason = RequestOutcomeFailed, RequestReasonClientError
	default:
		result.Outcome, result.Reason = RequestOutcomeFailed, RequestReasonServerError
	}
	return result, nil
}

func NewHTTPRequestRecord(metadata RequestMetadata, httpMethod, httpRoute string, result RequestResult) (RequestRecord, error) {
	return newRequestRecord(metadata, RequestRecord{HTTPMethod: httpMethod, HTTPRoute: httpRoute}, result)
}

func NewRPCRequestRecord(metadata RequestMetadata, httpMethod, rpcService, rpcMethod string, result RequestResult) (RequestRecord, error) {
	return newRequestRecord(metadata, RequestRecord{HTTPMethod: httpMethod, RPCService: rpcService, RPCMethod: rpcMethod}, result)
}

func newRequestRecord(metadata RequestMetadata, boundary RequestRecord, result RequestResult) (RequestRecord, error) {
	boundary.Event = "request.completed"
	boundary.OccurredAt = metadata.OccurredAt
	boundary.Correlation = metadata.Correlation
	boundary.RecordActor = metadata.RecordActor
	boundary.StatusCode = result.StatusCode
	boundary.DurationMS = result.DurationMS
	boundary.Outcome = result.Outcome
	boundary.ErrorCode = result.ErrorCode
	boundary.Reason = string(result.Reason)
	if err := boundary.Validate(); err != nil {
		return RequestRecord{}, err
	}
	return boundary, nil
}
