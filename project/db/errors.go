package db

import "errors"

var (
	ErrExceedingTicketLimit       = errors.New("exceeding ticket limit")
)
