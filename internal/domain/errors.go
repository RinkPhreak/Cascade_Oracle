package domain

import "errors"

var (
	// Core domain errors as defined in the architectural blueprint
	ErrNotFound        = errors.New("resource not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrPeerFlood       = errors.New("PEER_FLOOD")
	ErrAuthUnregistered= errors.New("AUTH_KEY_UNREGISTERED")
	ErrDeliveryFailed  = errors.New("delivery failed")
	ErrRateLimit       = errors.New("rate limited")
	ErrAccountBanned   = errors.New("account banned")
	ErrProxyDead       = errors.New("proxy dead")
	ErrDailyCheckCap   = errors.New("daily check cap reached")
	ErrSystemSuspended = errors.New("system suspended or halted")
)
