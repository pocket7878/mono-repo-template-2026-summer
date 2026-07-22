package todo

import "errors"

var (
	ErrNotFound        = errors.New("todo not found")
	ErrVersionConflict = errors.New("todo version conflict")
	ErrConflict        = errors.New("todo was modified by another request")
)
