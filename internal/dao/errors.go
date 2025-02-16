package dao

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrGoodNotFound      = errors.New("good not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrSelfTransfer      = errors.New("self-transfer not allowed")
)
