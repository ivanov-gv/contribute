package auth

import "errors"

// ErrTokenInvalid is returned when GitHub rejects the token with 401.
var ErrTokenInvalid = errors.New(
	"token invalid or expired — run 'contribute login' to reauthenticate",
)
