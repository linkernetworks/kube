package metricsclient

import "errors"

var (
	// ErrNotFound is returned when querying result is empty
	ErrNotFound = errors.New("not found")
	// ErrTypeConvertion is returned when type convertion (assertion) is not OK
	ErrTypeConvertion = errors.New("type convertion fail")
)
