// Package bobaerrors centralizes reusable error sentinels for BobaMixer subsystems.
package bobaerrors

import "errors"

var (
	// ErrConfig represents invalid configuration supplied by the user.
	ErrConfig = errors.New("configuration error")
	// ErrSecretsPerm indicates misconfigured permissions on secrets files.
	ErrSecretsPerm = errors.New("secrets permissions error")
	// ErrPricingUnavailable indicates pricing sources are unreachable or invalid.
	ErrPricingUnavailable = errors.New("pricing unavailable")
	// ErrHTTPUnauthorized indicates the upstream returned 401.
	ErrHTTPUnauthorized = errors.New("http unauthorized")
	// ErrHTTPForbidden indicates the upstream returned 403.
	ErrHTTPForbidden = errors.New("http forbidden")
	// ErrHTTPRetriable marks a transient HTTP failure eligible for retry.
	ErrHTTPRetriable = errors.New("http retriable error")
	// ErrToolExit signals a tool process exited with a non-zero code.
	ErrToolExit = errors.New("tool exit error")
	// ErrDB wraps database bootstrap or migration failures.
	ErrDB = errors.New("database error")
)
