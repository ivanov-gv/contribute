package config

import "errors"

// ErrNotAuthenticated is returned when no token is found.
var ErrNotAuthenticated = errors.New(
	"not authenticated — run 'gh contribute auth login' first",
)
