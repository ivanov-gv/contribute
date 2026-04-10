package config

import "errors"

// ErrNotAuthenticated is returned when no token is found.
var ErrNotAuthenticated = errors.New(
	"not authenticated — set GH_CONTRIBUTE_APP_ID and GH_CONTRIBUTE_PRIVATE_KEY_PATH, or run 'contribute login'",
)
