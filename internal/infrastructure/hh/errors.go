package hh

import "errors"

var (
	ErrNotFound  = errors.New("vacancy not found")
	ErrRateLimit = errors.New("rate limit exceeded")
)
