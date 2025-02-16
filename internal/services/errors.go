package services

import "errors"

var (
	ErrInternal          = errors.New("internal error")
	ErrToUserRequired    = errors.New("toUser is required")
	ErrAmountBelowZero   = errors.New("amount must be greater than zero")
	ErrCantSelfTransfer  = errors.New("can't transfer to yourself")
	ErrRecieverNotFound  = errors.New("recipient not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrUserPassRequired  = errors.New("username and password required")
	ErrAuthFailed        = errors.New("authentication failed")
	ErrItemTypeRequired  = errors.New("item type is required")
	ErrItemNotFound      = errors.New("item not found")
)
