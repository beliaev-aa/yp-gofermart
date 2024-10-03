package storage

type Storage struct {
	UserRepo       UserRepository
	OrderRepo      OrderRepository
	WithdrawalRepo WithdrawalRepository
}
