package apis

import (
	"errors"
)

const (
	// HTTP Request Fields
	IfModifiedSince   = "If-Modified-Since"
	IfUnmodifiedSince = "If-Unmodified-Since"
	IfMatch           = "If-Match"
	IfNoneMatch       = "If-None-Match"
	IfRange           = "If-Range"

	// HTTP Response Fields
	Location = "Location"
	ETag     = "ETag"

	// Self-defined Fields
	Filter        = "filter"
	Start         = "start"
	End           = "end"
	Select        = "select"
	Sort          = "sort"
	DESC          = "desc"
	ASC           = "asc"
	Limit         = "limit"
	Latest        = "latest"
	Interval      = "interval"
	OrphanRemoval = "orphanRemoval"
)

var (
	ErrMismatch     = errors.New("resource mismatch")
	ErrInternal     = errors.New("internal error")
	ErrInvalidValue = errors.New("invalid value")
	ErrImmutable    = errors.New("immutable")
)
