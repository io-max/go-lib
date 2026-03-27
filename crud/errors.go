package gincrud

import "errors"

var (
	ErrRecordNotFound          = errors.New("record not found")
	ErrCannotDeleteHasChildren = errors.New("cannot delete record with children")
	ErrDuplicateEntry          = errors.New("duplicate entry")
	ErrEntityIDRequired        = errors.New("entity ID is required")
)
