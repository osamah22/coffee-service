package authn

import "errors"

var (
	ErrMissingIssuer   = errors.New("auth issuer is required")
	ErrMissingAudience = errors.New("auth audience is required")
	ErrMissingJWKSURL  = errors.New("auth jwks url is required")
)
