package errors

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad request")
	ErrNotFound     = errors.New("not found")
	ErrInternal     = errors.New("internal server error")
)

var (
	ErrUserNotFound = errors.New("user not found")
)

var (
	ErrTenderNotFound        = errors.New("tender not found")
	ErrTenderHistoryNotFound = errors.New("tender history not found")
)

var (
	ErrBidNotFound      = errors.New("bid not found")
	ErrInvalidBidStatus = errors.New("invalid bid status")
	ErrInvalidUUID      = errors.New("invalid uuid")
)
