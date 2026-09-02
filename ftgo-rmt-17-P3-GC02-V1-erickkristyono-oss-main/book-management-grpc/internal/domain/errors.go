package domain

import "errors"

// Sentinel errors dipakai lintas-layer agar layer delivery dapat memetakan
// error domain ke gRPC status code yang sesuai tanpa bergantung pada teks error.
var (
	ErrNotFound           = errors.New("resource not found")
	ErrUserAlreadyExists  = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrValidation         = errors.New("validation error")
	ErrBookNotAvailable   = errors.New("book is not available for borrowing")
	ErrForbidden          = errors.New("you are not allowed to perform this action")
	ErrUnauthenticated    = errors.New("unauthenticated")
)
