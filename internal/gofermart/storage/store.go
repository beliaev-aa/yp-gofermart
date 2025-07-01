package storage

import "beliaev-aa/yp-gofermart/internal/gofermart/storage/repository"

type Storage struct {
	UserRepo       repository.UserRepository
	OrderRepo      repository.OrderRepository
	WithdrawalRepo repository.WithdrawalRepository
}
