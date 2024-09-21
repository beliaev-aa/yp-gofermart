package domain

import "time"

const (
	OrderStatusInvalid    = "INVALID"
	OrderStatusNew        = "NEW"
	OrderStatusProcessed  = "PROCESSED"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusRegistered = "REGISTERED"
)

type User struct {
	ID       int
	Login    string
	Password string
	Balance  float64
}

type Order struct {
	Number     string
	UserID     int
	Status     string
	Accrual    float64
	UploadedAt time.Time
	Processing bool
}

type Withdrawal struct {
	ID          int
	Order       string
	UserID      int
	Sum         float64
	ProcessedAt time.Time
}
