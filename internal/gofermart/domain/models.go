package domain

import "time"

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type User struct {
	ID        int
	Login     string
	Password  string
	Balance   float64
	Withdrawn float64
}

type Order struct {
	Number     string
	UserID     int
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

type Withdrawal struct {
	ID          int
	Order       string
	UserID      int
	Sum         float64
	ProcessedAt time.Time
}
