package errors

import "errors"

var (
	ErrAccrualSystemUnavailable = errors.New("accrual system unavailable")
	ErrInsufficientFunds        = errors.New("insufficient funds")
	ErrLoginAlreadyExists       = errors.New("login already exists")
	ErrOrderAlreadyExists       = errors.New("order number already exists")
	ErrOrderAlreadyUploaded     = errors.New("order already uploaded by this user")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderUploadedByAnother   = errors.New("order already uploaded by another user")
	ErrUserNotFound             = errors.New("user not found")
)
