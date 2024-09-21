package storage

import "beliaev-aa/yp-gofermart/internal/gofermart/domain"

type Storage interface {
	AddOrder(order domain.Order) error
	AddWithdrawal(withdrawal domain.Withdrawal) error
	GetOrderByNumber(number string) (*domain.Order, error)
	GetOrdersForProcessing() ([]domain.Order, error)
	GetOrdersByUserID(userID int) ([]domain.Order, error)
	GetUserByLogin(login string) (*domain.User, error)
	GetWithdrawalsByUserID(userID int) ([]domain.Withdrawal, error)
	SaveUser(user domain.User) error
	UpdateOrder(order domain.Order) error
	UpdateUserBalance(userID int, amount float64) error
	LockOrderForProcessing(orderNumber string) error
	UnlockOrder(orderNumber string) error
	GetUserBalance(login string) (balance, withdrawal float64, err error)
}