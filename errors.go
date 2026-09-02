package telemetry

import "errors"

var (
	ErrInvalidRequestID = errors.New("request_id must be a canonical UUIDv4")
	ErrInvalidSourceIP  = errors.New("source_ip must be a canonical IPv4 or IPv6 address")
)
