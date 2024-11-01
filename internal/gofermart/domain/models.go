package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

// Статусы заказов
const (
	OrderStatusInvalid    = "INVALID"
	OrderStatusNew        = "NEW"
	OrderStatusProcessed  = "PROCESSED"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusRegistered = "REGISTERED"
)

// User - представляет пользователя в системе.
type User struct {
	UserID   int             `gorm:"column:user_id;primaryKey;autoIncrement"`
	Login    string          `gorm:"column:login;unique;not null;index"`
	Password string          `gorm:"column:password;not null"`
	Balance  decimal.Decimal `gorm:"column:balance;type:numeric(18,2);default:0"`
}

// Order - представляет заказ, связанный с пользователем.
type Order struct {
	OrderNumber  string          `gorm:"column:order_number;primaryKey;index"`
	UserID       int             `gorm:"column:user_id;not null;index"`
	OrderStatus  string          `gorm:"column:order_status;not null"`
	Accrual      decimal.Decimal `gorm:"column:accrual;type:numeric(18,2);default:0"`
	UploadedAt   time.Time       `gorm:"column:uploaded_at;type:timestamp with time zone;not null;index"`
	IsProcessing bool            `gorm:"column:is_processing;default:false;index"`
}

// Withdrawal - представляет вывод средств пользователем.
type Withdrawal struct {
	WithdrawalID int             `gorm:"column:withdrawal_id;primaryKey;autoIncrement"`
	OrderNumber  string          `gorm:"column:order_number;not null;index"`
	UserID       int             `gorm:"column:user_id;not null;index"`
	Amount       decimal.Decimal `gorm:"column:amount;type:numeric(18,2);not null"`
	ProcessedAt  time.Time       `gorm:"column:processed_at;type:timestamp with time zone;not null"`
}
