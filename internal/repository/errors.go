package repository

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrGoodNotFound      = errors.New("good not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrSelfTransfer      = errors.New("self-transfer not allowed")
	ErrCreateUser        = errors.New("failed to create user")
	ErrGetUser           = errors.New("failed to get user")
	ErrGetBalance        = errors.New("failed to get balance")
	ErrGetHistory        = errors.New("failed to get history")
	ErrGetInventory      = errors.New("failed to get inventory")
	ErrTransferCoins     = errors.New("failed to transfer coins")
	ErrBuyItem           = errors.New("failed to buy item")
	ErrGetGood           = errors.New("failed to get good")
)
